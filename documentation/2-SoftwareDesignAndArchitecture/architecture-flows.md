# CodeValdDT — Architecture: Error Types, Flows & SharedLib

> Part of [architecture.md](architecture.md)

## 8. Error Types

Defined in `errors.go`:

```go
var (
    ErrEntityNotFound       = errors.New("entity not found")
    ErrRelationshipNotFound = errors.New("relationship not found")
    ErrInvalidEntity        = errors.New("invalid entity: missing required fields")
    ErrInvalidRelationship  = errors.New("invalid relationship: missing required fields")
    ErrTelemetryNotFound    = errors.New("telemetry not found")
    ErrEventNotFound        = errors.New("event not found")
    ErrSchemaNotFound       = errors.New("schema not found")
    ErrInvalidSchema        = errors.New("invalid schema: missing required fields")
)
```

Map to gRPC status codes in `internal/server/server.go`:

| Error | gRPC code |
|---|---|
| `ErrEntityNotFound` | `codes.NotFound` |
| `ErrRelationshipNotFound` | `codes.NotFound` |
| `ErrTelemetryNotFound` | `codes.NotFound` |
| `ErrEventNotFound` | `codes.NotFound` |
| `ErrInvalidEntity` | `codes.InvalidArgument` |
| `ErrInvalidRelationship` | `codes.InvalidArgument` |
| `ErrSchemaNotFound` | `codes.NotFound` |
| `ErrInvalidSchema` | `codes.InvalidArgument` |
| all others | `codes.Internal` |

---

## 9. CreateEntity Flow (Critical Path)

```
gRPC handler
    │
    ▼
DTManager.CreateEntity(ctx, req)
    │
    ├── validate: req.AgencyID, req.TypeID non-empty
    │       → ErrInvalidEntity if violated
    │
    ├── backend.InsertEntity(ctx, req)     → entities collection in dt_{agencyID} database
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
successful create. Publish failures are logged but not returned to the caller —
the entity is already persisted.

---

## 10. RecordTelemetry Flow (Critical Path)

```
gRPC handler
    │
    ▼
DTManager.RecordTelemetry(ctx, req)
    │
    ├── validate: req.AgencyID, req.EntityID, req.Name non-empty
    │
    ├── backend.InsertTelemetry(ctx, req)  → telemetry collection in dt_{agencyID} database
    │       returns TelemetryReading
    │
    └── crossPublisher.Publish(ctx,
            "cross.dt.{agencyID}.telemetry.recorded",
            telemetryReading.ID)
            │
            ▼
        Cross routes event to subscribers
```

`cross.dt.{agencyID}.telemetry.recorded` **MUST be published** after every
successful record. Same best-effort rule applies.

---

## 11. SharedLib Dependency

CodeValdDT imports `github.com/aosanya/CodeValdSharedLib` for:

| SharedLib package | Provides |
|---|---|
| `types/schema` | `Schema`, `TypeDefinition`, `PropertyDefinition`, `PropertyType`, `RatingConfig` |
| `registrar` | Cross registration heartbeat loop |
| `serverutil` | `envOrDefault`, `parseDuration` helpers and gRPC server setup |
| `arangoutil` | ArangoDB driver connection, auth, database bootstrap |
| `gen/go/codevaldcross/v1` | Cross stubs for Register + Publish calls |

> **Principle**: Any infrastructure code shared across services lives in
> SharedLib. CodeValdDT retains only domain logic, domain errors, gRPC
> handlers, and storage collection schemas.

See task DT-012 in [mvp.md](../3-SofwareDevelopment/mvp.md) for SharedLib
migration scope.
