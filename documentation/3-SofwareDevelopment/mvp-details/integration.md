# Cross Integration & Service Registration

Topics: Cross Registration · Pub/Sub Topics · Route Table · SharedLib Wiring · cmd/main.go

---

## Tasks

| Task | Status | Depends On |
|---|---|---|
| MVP-DT-005 — Cross registration heartbeat (`internal/registrar/registrar.go`) | ✅ Done (2026-04-27) | MVP-DT-001 |
| MVP-DT-006 — Registrar unit tests | ✅ Done (2026-04-28) | MVP-DT-005 |

Architecture ref: [architecture-service.md §6](../../2-SoftwareDesignAndArchitecture/architecture-service.md),
[architecture-flows.md §11](../../2-SoftwareDesignAndArchitecture/architecture-flows.md)

---

## Overview

CodeValdDT registers with CodeValdCross on startup and repeats the registration
every **20 seconds** as a liveness heartbeat. Cross expires services that stop
registering. If Cross is unreachable, the loop retries silently.

The registrar is the **only** outbound Cross call at the service layer.
`CreateEntity` publishes events via the `crossPublisher` interface injected into
`DTDataManager` — that is handled by the SharedLib `entitygraph` layer, not by
`internal/registrar`.

---

## Acceptance Criteria

- [x] `internal/registrar/registrar.go` starts a goroutine on `app.Run` that calls `OrchestratorService.Register` every 20 seconds
- [x] `RegisterRequest.ServiceName == "codevalddt"`
- [x] `RegisterRequest.Addr == ":50055"`
- [x] `RegisterRequest.Produces` enumerates the three Cross topics (see below)
- [x] `RegisterRequest.Consumes` is empty (CodeValdDT does not subscribe to any topics in v1)
- [x] `RegisterRequest.Routes` lists all 10 HTTP route patterns (see below)
- [x] Registration failures are logged and retried — service startup is not blocked
- [x] `go build ./...`, `go vet ./...`, `go test -race ./...` all pass

---

## Cross Registration Payload

```go
RegisterRequest{
    ServiceName: "codevalddt",
    Addr:        ":50055",
    Produces: []string{
        "cross.dt.{agencyID}.entity.created",
        "cross.dt.{agencyID}.telemetry.recorded",
        "cross.dt.{agencyID}.event.recorded",
    },
    Consumes: []string{},
    Routes: []Route{
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "PUT",    Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "DELETE", Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/relationships"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/relationships/{relationshipId}"},
        {Method: "DELETE", Pattern: "/{agencyId}/dt/relationships/{relationshipId}"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/relationships"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities/{entityId}/traverse"},
    },
}
```

---

## Pub/Sub Topic Semantics

Topics are published by `DTDataManager` (SharedLib `entitygraph.DataManager`)
after every successful `CreateEntity`. The topic is chosen from the resolved
`TypeDefinition.StorageCollection`:

| `StorageCollection` | Published topic |
|---|---|
| `"dt_entities"` (or empty) | `cross.dt.{agencyID}.entity.created` |
| `"dt_telemetry"` | `cross.dt.{agencyID}.telemetry.recorded` |
| `"dt_events"` | `cross.dt.{agencyID}.event.recorded` |

**No consumers are declared for telemetry or event topics in v1.** The topics
exist on the Cross bus; subscriber registration is deferred until a consuming
service scopes the requirement.

Publish failures are logged but **not** returned to the caller — the entity is
already persisted.

---

## SharedLib Dependencies

CodeValdDT delegates all infrastructure to SharedLib. No infrastructure is
duplicated in this repo.

| SharedLib package | Provides |
|---|---|
| `entitygraph` | `DataManager` + `SchemaManager` interfaces; all entity/relationship/schema models |
| `entitygraph/arangodb` | Concrete ArangoDB backend; bootstraps collections and indexes |
| `entitygraph/server` | `EntityServer` — SharedLib gRPC handlers; CodeValdDT re-exports via `internal/server/entity_server.go` |
| `types/schema` | `Schema`, `TypeDefinition`, `PropertyDefinition`, `PropertyType` |
| `registrar` | Cross registration heartbeat loop (20 s default) |
| `serverutil` | `envOrDefault`, `parseDuration` helpers; gRPC server setup |
| `arangoutil` | ArangoDB driver connection and auth |
| `gen/go/codevaldcross/v1` | Cross `Register` + `Publish` gRPC stubs |

> **Principle**: Any infrastructure shared across CodeVald services lives in
> SharedLib. CodeValdDT retains only domain errors, storage collection
> configuration, and the Cross route table.

---

## cmd/main.go Wiring

```
cmd/main.go
  │
  ├── Load config (internal/config)
  ├── Connect ArangoDB (arangoutil.Connect)
  ├── Build backend (storage/arangodb.NewBackend — fixes dt_* collection names)
  ├── Build DTDataManager (entitygraph.NewDataManager with backend)
  ├── Seed DefaultDTSchema idempotently (entitygraph.SeedSchema)
  ├── Register EntityServer (entitygraphpb.RegisterEntityServiceServer)
  ├── Start Cross registrar heartbeat (internal/registrar)
  └── Start gRPC server on :50055
```

---

## Config Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DT_ARANGO_ENDPOINT` | `http://localhost:8529` | ArangoDB server URL |
| `DT_ARANGO_DATABASE` | `codevald` | Shared ArangoDB database name |
| `DT_ARANGO_USERNAME` | `root` | ArangoDB username |
| `DT_ARANGO_PASSWORD` | `""` | ArangoDB password |
| `CROSS_ADDR` | `localhost:50051` | CodeValdCross gRPC address |
| `DT_PORT` | `:50055` | gRPC listen address |

---

## Tests

| Test | File | Coverage |
|---|---|---|
| `TestRegistrar_CallsRegister` | `internal/registrar/registrar_test.go` | Register called on startup with correct payload |
| `TestRegistrar_Heartbeat` | `internal/registrar/registrar_test.go` | Second call fires within 20 s window |
| `TestRegistrar_CrossUnavailable` | `internal/registrar/registrar_test.go` | Failure logged; startup not blocked |
| `TestConfig_Defaults` | `internal/config/config_test.go` | All env var defaults applied correctly |
| `TestConfig_Override` | `internal/config/config_test.go` | Env vars override defaults |
| `TestApp_Run_EndToEnd` | `internal/app/app_integration_test.go` | Full startup against real ArangoDB; gRPC surface exercised |

Integration tests tagged `//go:build integration`; skip without `DT_ARANGO_ENDPOINT`.
