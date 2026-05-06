# Entity Lifecycle

Topics: Entity CRUD · Storage Routing · Immutability Guard · Soft Delete · gRPC Error Mapping

---

## Tasks

| Task | Status | Depends On |
|---|---|---|
| MVP-DT-001 — Module scaffolding: `errors.go`, `models.go`, `schema.go`, `cmd/main.go` | ✅ Done (2026-04-27) | — |
| MVP-DT-002 — ArangoDB backend: `storage/arangodb/storage.go`, collection bootstrap, indexes | ✅ Done (2026-04-27) | MVP-DT-001 |
| MVP-DT-006 — Unit & integration tests covering entity CRUD and immutability | ✅ Done (2026-04-28) | MVP-DT-001, MVP-DT-002 |

Architecture ref: [architecture-interfaces.md](../../2-SoftwareDesignAndArchitecture/architecture-interfaces.md),
[architecture-flows.md](../../2-SoftwareDesignAndArchitecture/architecture-flows.md)

---

## Overview

Entity lifecycle is the core capability of CodeValdDT. The gRPC surface is the
shared `EntityService` from `CodeValdSharedLib/entitygraph/server` — CodeValdDT
does **not** define its own proto; it re-exports the shared server via
`internal/server/entity_server.go`.

The five entity RPCs delegate to `DTDataManager` (= `entitygraph.DataManager`):

| RPC | Go method | Notes |
|---|---|---|
| `CreateEntity` | `dm.CreateEntity` | Routes to collection by `TypeDefinition.StorageCollection`; publishes Cross topic |
| `GetEntity` | `dm.GetEntity` | Returns `ErrEntityNotFound` for soft-deleted entities |
| `UpdateEntity` | `dm.UpdateEntity` | Returns `ErrImmutableType` (`FailedPrecondition`) for immutable type definitions |
| `DeleteEntity` | `dm.DeleteEntity` | Soft delete — sets `deleted=true`, `deletedAt=now()` |
| `ListEntities` | `dm.ListEntities` | Scoped by `agencyID` + optional `typeID` filter |

---

## Acceptance Criteria

- [x] `errors.go`: `ErrEntityNotFound`, `ErrInvalidEntity`, `ErrImmutableType`, and other exported sentinels
- [x] `models.go`: `DTDataManager = entitygraph.DataManager`, `DTSchemaManager = entitygraph.SchemaManager`
- [x] `schema.go`: `DefaultDTSchema()` returns an empty scaffold (no built-in TypeDefinitions — all types are agency-defined at runtime)
- [x] `storage/arangodb/storage.go`: thin shim over `CodeValdSharedLib/entitygraph/arangodb` with fixed `dt_*` collection and graph names
- [x] `internal/server/entity_server.go`: re-exports `egserver.NewEntityServer` from SharedLib
- [x] `internal/app/app.go`: seeds `DefaultDTSchema` idempotently on startup via `entitygraph.SeedSchema`; registers `entitygraphpb.RegisterEntityServiceServer`
- [x] `CreateEntity` routes to the collection named in `TypeDefinition.StorageCollection`; defaults to `dt_entities`
- [x] `UpdateEntity` returns `codes.FailedPrecondition` (`ErrImmutableType`) for types with `Immutable: true`
- [x] `DeleteEntity` sets `deleted=true` and `deletedAt` — never hard-deletes
- [x] Soft-deleted entities are excluded from `ListEntities` and `GetEntity` responses
- [x] `go build ./...`, `go vet ./...`, `go test -race ./...` all pass

---

## CreateEntity Flow (Critical Path)

```
gRPC handler (SharedLib EntityServer)
    │
    ▼
DTDataManager.CreateEntity(ctx, req)
    │
    ├── validate: req.AgencyID, req.TypeID non-empty → ErrInvalidEntity
    │
    ├── DTSchemaManager.GetSchema(ctx, req.AgencyID, active=true)
    │       → resolve TypeDefinition for req.TypeID
    │       → collection = TypeDefinition.StorageCollection
    │                       (default: "dt_entities" if empty)
    │
    ├── backend.InsertEntity(ctx, req, collection)
    │       returns Entity{ID, AgencyID, TypeID, Properties, CreatedAt, UpdatedAt}
    │
    └── crossPublisher.Publish(ctx, topic, entity.ID)
            topic selected by resolved StorageCollection:
              "dt_entities"  → cross.dt.{agencyID}.entity.created
              "dt_telemetry" → cross.dt.{agencyID}.telemetry.recorded
              "dt_events"    → cross.dt.{agencyID}.event.recorded
            Publish failures logged, never returned to caller.
```

---

## UpdateEntity Flow (Immutability Guard)

```
DTDataManager.UpdateEntity(ctx, agencyID, entityID, req)
    │
    ├── validate: agencyID, entityID non-empty → ErrInvalidEntity
    ├── backend.GetEntity → ErrEntityNotFound if deleted or not found
    ├── DTSchemaManager.GetSchema → resolve TypeDefinition for entity.TypeID
    │       if TypeDefinition.Immutable == true → ErrImmutableType
    └── backend.UpdateEntity → returns updated Entity
```

Immutable types (telemetry readings, events) reject `UpdateEntity` with
`codes.FailedPrecondition`. `CreateEntity` and `DeleteEntity` remain valid.

---

## DeleteEntity Flow (Soft Delete)

```
DTDataManager.DeleteEntity(ctx, agencyID, entityID)
    │
    ├── validate: agencyID, entityID non-empty → ErrInvalidEntity
    ├── backend.GetEntity → ErrEntityNotFound if already deleted
    └── backend.DeleteEntity
            ↳ sets deleted=true, deletedAt=now() on the ArangoDB document
            No cascade — relationships, telemetry, events are untouched.
```

Soft-delete rules:
- Deleted entities excluded from `ListEntities` and `TraverseGraph`
- `GetEntity` returns `ErrEntityNotFound` for deleted entities in v1
- Relationships become orphans; hard delete and orphan cleanup deferred to v2

---

## Storage Routing

`TypeDefinition.StorageCollection` is the **only** routing mechanism. No hard-coded
type checks exist in the manager or handlers.

| `StorageCollection` | Backing collection | Immutable? | Cross topic |
|---|---|---|---|
| `""` (empty) | `dt_entities` | No | `cross.dt.{agencyID}.entity.created` |
| `"dt_telemetry"` | `dt_telemetry` | Yes | `cross.dt.{agencyID}.telemetry.recorded` |
| `"dt_events"` | `dt_events` | Yes | `cross.dt.{agencyID}.event.recorded` |

`DefaultDTSchema()` ships **no** built-in TypeDefinitions — every agency populates
its own `TypeDefinition` list at runtime via `DTSchemaManager.SetSchema`.

---

## gRPC Error Code Mapping

| Go error | gRPC code |
|---|---|
| `ErrEntityNotFound` | `codes.NotFound` |
| `ErrInvalidEntity` | `codes.InvalidArgument` |
| `ErrImmutableType` | `codes.FailedPrecondition` |
| `ErrSchemaNotFound` | `codes.NotFound` |
| all others | `codes.Internal` |

Error mapping lives in `CodeValdSharedLib/entitygraph/server` — CodeValdDT does
not duplicate it.

---

## Document Shape

**`dt_entities/{id}`**

```json
{
  "_key":       "entity-uuid",
  "agencyID":   "agency-123",
  "typeID":     "Pump",
  "properties": { "pressure": 4.2, "status": "running" },
  "createdAt":  "2026-01-01T00:00:00Z",
  "updatedAt":  "2026-01-01T00:00:00Z",
  "deleted":    false,
  "deletedAt":  null
}
```

---

## Indexes (dt_entities)

| Field(s) | Type | Purpose |
|---|---|---|
| `agencyID` | persistent | Scope all entity queries to the agency |
| `typeID` | persistent | `ListEntities` by type |
| `agencyID, deleted` | persistent | Efficiently exclude soft-deleted entities |

---

## Tests

| Test | File | Coverage |
|---|---|---|
| `TestDefaultDTSchema_Metadata` | `schema_test.go` | Schema ID, Version, Tag |
| `TestDefaultDTSchema_PlatformMetaTypes` | `schema_test.go` | TelemetryType and EventType present; no domain types |
| `TestDefaultDTSchema_StorageCollections` | `schema_test.go` | TelemetryType→dt_telemetry_types, EventType→dt_event_types |
| `TestDefaultDTSchema_TelemetryTypeRequiredProperties` | `schema_test.go` | TelemetryType.name is required string |
| `TestDefaultDTSchema_EventTypeRequiredProperties` | `schema_test.go` | EventType.name is required string |
| `TestNewBackend_*` | `storage/arangodb/storage_test.go` | Backend construction; correct collection/graph names |
| `TestConfig_*` | `internal/config/config_test.go` | Env var loading; defaults |
| `TestEntityServer_CreateEntity` | `internal/app/app_integration_test.go` | End-to-end via gRPC against ArangoDB |
| `TestEntityServer_UpdateEntity_Immutable` | `internal/app/app_integration_test.go` | `FailedPrecondition` for immutable types |
| `TestEntityServer_DeleteEntity_SoftDelete` | `internal/app/app_integration_test.go` | Deleted entity excluded from Get/List |

Integration tests tagged `//go:build integration`; skip without `DT_ARANGO_ENDPOINT`.
