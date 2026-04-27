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
    ├── dtSchemaManager.GetSchema(ctx, req.AgencyID, /* Schema.Active == true */)
    │       → resolve TypeDefinition for req.TypeID
    │       → ErrSchemaNotFound if no active schema exists
    │       → collection = TypeDefinition.StorageCollection
    │                       (default: "dt_entities" if empty)
    │
    ├── backend.InsertEntity(ctx, req, collection)    → resolved collection
    │       returns Entity{ID, AgencyID, TypeID, Properties, CreatedAt, UpdatedAt}
    │
    └── crossPublisher.Publish(ctx, topic, entity.ID)
            │   topic is selected from the resolved StorageCollection:
            │     "dt_entities"  → cross.dt.{agencyID}.entity.created
            │     "dt_telemetry" → cross.dt.{agencyID}.telemetry.recorded
            │     "dt_events"    → cross.dt.{agencyID}.event.recorded
            ▼
        Cross routes the event to subscribers
```

A topic **MUST be published** after every successful create. Publish failures
are logged but not returned to the caller — the entity is already persisted.

**Telemetry readings** are created via this same flow with a `typeID` whose
`TypeDefinition` has `StorageCollection: "dt_telemetry"` and `Immutable: true`.
The reading's `value`, source `entityID`, and reading `timestamp` are carried
inside `properties`.

**Events** follow the identical pattern with `StorageCollection: "dt_events"`
and `Immutable: true`. Payload, source `entityID`, and event `timestamp` live
inside `properties`.

> Telemetry and events are **never** modelled as separate Go types — they are
> `Entity` instances throughout. There is no `RecordTelemetry` /
> `RecordEvent` / `QueryTelemetry` / `ListEvents` RPC; reads use
> `ListEntities` filtered by `TypeID` and (when implemented in
> `entitygraph.EntityFilter`) a time range against `properties.timestamp`.

> **Active schema selection**: `DTSchemaManager.GetSchema` returns the single
> `Schema` document for the agency whose `Active` flag is `true`. Only one
> published schema per agency is active at a time; drafts have `Active:
> false`. See [`types.Schema.Active`](file:///workspaces/CodeVald-AIProject/CodeValdSharedLib/types/schema.go) in
> `CodeValdSharedLib/types/schema.go` for the field definition.

> **Default ordering on time-series collections**: when the resolved
> `TypeDefinition.StorageCollection` is `"dt_telemetry"` or `"dt_events"`,
> `ListEntities` returns rows sorted by `properties.timestamp ASC`. This is a
> property of the `entitygraph.EntityFilter` contract (tracked in
> `SHAREDLIB-014`) — the AQL query orders on `properties.timestamp`, served
> by the existing `(properties.entityID, properties.timestamp)` persistent
> index.

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
    ├── dtSchemaManager.GetSchema(ctx, agencyID, /* Schema.Active == true */)
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

See [mvp.md](../3-SofwareDevelopment/mvp.md) — every MVP-DT-* task depends on
`SHAREDLIB-010`, which is the SharedLib migration that exposes the
`entitygraph` and `arangoutil` packages used here.
