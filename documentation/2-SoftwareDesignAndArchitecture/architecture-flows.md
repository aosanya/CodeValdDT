# CodeValdDT — Architecture: Error Types, Flows & SharedLib

> Part of [architecture.md](architecture.md)

## 8. Error Types

Defined in `errors.go`:

```go
var (
    ErrEntityNotFound       = errors.New("entity not found")
    ErrRelationshipNotFound = errors.New("relationship not found")
    ErrSchemaNotFound       = errors.New("schema not found")
    ErrInvalidEntity        = errors.New("invalid entity: missing required fields")
    ErrInvalidRelationship  = errors.New("invalid relationship: missing required fields")
    ErrInvalidSchema        = errors.New("invalid schema: missing required fields")
    ErrImmutableType        = errors.New("entity type is immutable: update not allowed")
)
```

Map to gRPC status codes in `internal/server/server.go`:

| Error | gRPC code |
|---|---|
| `ErrEntityNotFound` | `codes.NotFound` |
| `ErrRelationshipNotFound` | `codes.NotFound` |
| `ErrSchemaNotFound` | `codes.NotFound` |
| `ErrInvalidEntity` | `codes.InvalidArgument` |
| `ErrInvalidRelationship` | `codes.InvalidArgument` |
| `ErrInvalidSchema` | `codes.InvalidArgument` |
| `ErrImmutableType` | `codes.FailedPrecondition` |
| all others | `codes.Internal` |

---

## 9. CreateEntity Flow (Critical Path)

```
gRPC handler
    │
    ▼
DTDataManager.CreateEntity(ctx, req)
    │
    ├── validate: req.AgencyID, req.TypeID non-empty
    │       → ErrInvalidEntity if violated
    │
    ├── dtSchemaManager.GetSchema(ctx, req.AgencyID, activeVersion)
    │       → resolve TypeDefinition for req.TypeID
    │       → ErrSchemaNotFound if no active schema exists
    │       → collection = TypeDefinition.StorageCollection
    │                       (default: "dt_entities" if empty)
    │
    ├── backend.InsertEntity(ctx, req, collection)    → resolved collection
    │       returns Entity{ID, AgencyID, TypeID, Properties, CreatedAt, UpdatedAt}
    │
    └── crossPublisher.Publish(ctx,
            "cross.dt.{agencyID}.entity.created",
            entity.ID)
            │
            ▼
        Cross routes event to subscribers
```

`cross.dt.{agencyID}.entity.created` **MUST be published** after every
successful create — regardless of the resolved collection (entities,
telemetry, events). Publish failures are logged but not returned to the
caller — the entity is already persisted.

Telemetry readings and events are created via this same flow with a TypeID
whose `TypeDefinition` has `StorageCollection: "dt_telemetry"` or
`"dt_events"` and `Immutable: true`.

---

## 10. UpdateEntity Flow (Immutability Guard)

```
gRPC handler
    │
    ▼
DTDataManager.UpdateEntity(ctx, agencyID, entityID, req)
    │
    ├── validate: agencyID, entityID non-empty
    │       → ErrInvalidEntity if violated
    │
    ├── backend.GetEntity(ctx, agencyID, entityID)
    │       → ErrEntityNotFound if not found or already deleted
    │
    ├── dtSchemaManager.GetSchema(ctx, agencyID, activeVersion)
    │       → resolve TypeDefinition for entity.TypeID
    │       → if TypeDefinition.Immutable == true → ErrImmutableType
    │
    └── backend.UpdateEntity(ctx, agencyID, entityID, req)
            returns updated Entity
```

Immutable types (e.g. telemetry readings, events) reject `UpdateEntity` with
`ErrImmutableType` (`codes.FailedPrecondition`). `CreateEntity` and
`DeleteEntity` remain valid for all types.

---

## 10a. DeleteEntity Flow (Soft Delete)

```
gRPC handler
    │
    ▼
DTDataManager.DeleteEntity(ctx, agencyID, entityID)
    │
    ├── validate: agencyID, entityID non-empty
    │       → ErrInvalidEntity if violated
    │
    ├── backend.GetEntity(ctx, agencyID, entityID)
    │       → ErrEntityNotFound if entity does not exist or is already deleted
    │
    └── backend.DeleteEntity(ctx, agencyID, entityID)
            ↳ sets deleted = true, deletedAt = now() on the ArangoDB document
            returns nil on success
```

Soft-delete rules:
- **No cascade** — relationships, telemetry, and events linked to the entity are **not** modified
- Soft-deleted entities are excluded from `ListEntities` and `TraverseGraph` results
- `GetEntity` returns `ErrEntityNotFound` for deleted entities in v1
- Hard delete is not exposed in v1; relationships become orphans and are cleaned up in a future version

---

## 11. SharedLib Dependency

CodeValdDT imports `github.com/aosanya/CodeValdSharedLib` for:

| SharedLib package | Provides |
|---|---|
| `entitygraph` | `DataManager` + `SchemaManager` interfaces; `Entity`, `Relationship`, all request/filter/result models — aliased as `DTDataManager` / `DTSchemaManager` |
| `types/schema` | `Schema`, `TypeDefinition`, `PropertyDefinition`, `PropertyType`, `RatingConfig` |
| `registrar` | Cross registration heartbeat loop |
| `serverutil` | `envOrDefault`, `parseDuration` helpers and gRPC server setup |
| `arangoutil` | ArangoDB driver connection and auth — connects to the pre-existing database; `DTDataManager` then ensures collections and indexes exist on startup |
| `gen/go/codevaldcross/v1` | Cross `Register` + `Publish` gRPC stubs — `DTDataManager` calls `Publish` on the generated client directly; no additional wrapper needed in v1 |

> **Principle**: Any infrastructure code shared across services lives in
> SharedLib. CodeValdDT retains only domain logic, domain errors, gRPC
> handlers, and storage collection schemas.

See task DT-012 in [mvp.md](../3-SofwareDevelopment/mvp.md) for SharedLib
migration scope.
