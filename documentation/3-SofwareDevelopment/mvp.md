# MVP — CodeValdDT

## Goal

Deliver a production-ready digital twin graph gRPC microservice with ArangoDB
persistence and CodeValdCross registration, backed by `entitygraph.DataManager`
and `entitygraph.SchemaManager` from CodeValdSharedLib.

---

## Task List

| Task ID | Title | Status | Depends On | Notes |
|---|---|---|---|---|
| MVP-DT-001 | Module Scaffolding | 🚀 In Progress | ~~SHAREDLIB-010~~ ✅, ~~SHAREDLIB-011~~ ✅ | `go.mod`, `errors.go`, `models.go` (DTDataManager + DTSchemaManager aliases), `internal/config/`, `cmd/main.go` skeleton; `feature/DT-001_module-scaffolding` |
| MVP-DT-002 | ArangoDB Backend | ⏸️ Blocked | MVP-DT-001 | `storage/arangodb/storage.go`; bootstrap `dt_schemas`, `dt_entities`, `dt_relationships` (**edge**), `dt_telemetry`, `dt_events`; named graph `dt_graph`; indexes per [architecture-storage.md](../2-SoftwareDesignAndArchitecture/architecture-storage.md); `feature/DT-002_arangodb-backend` |
| MVP-DT-003 | gRPC Service Proto & Codegen | ⏸️ Blocked | MVP-DT-001 | `proto/codevalddt/v1/dt.proto`; 10 RPCs (entity + relationship + traverse — no telemetry/event RPCs, those route via `CreateEntity`); `buf generate` → `gen/go/`; `feature/DT-003_grpc-proto-codegen` |
| MVP-DT-004 | gRPC Server Implementation | ⏸️ Blocked | MVP-DT-001, MVP-DT-002, MVP-DT-003 | `internal/server/server.go`; thin handlers delegate to `DTDataManager`; `toGRPCError` mapping; `CreateEntity` flow branches on resolved `TypeDefinition.StorageCollection` for the Cross publish topic (`entity.created` / `telemetry.recorded` / `event.recorded`); `feature/DT-004_grpc-server-implementation` |
| MVP-DT-005 | CodeValdCross Registration | ⏸️ Blocked | MVP-DT-004 | SharedLib `registrar`; heartbeat every 20 s; full `cmd/main.go` wiring via `serverutil`; `Produces:` lists all three Cross topics; `feature/DT-005_cross-registration` |
| MVP-DT-006 | Unit & Integration Tests | ⏸️ Blocked | MVP-DT-001, MVP-DT-002, MVP-DT-004 | Table-driven unit tests with mock `DTDataManager`; integration tests tagged `//go:build integration`; coverage ≥ 80% on `internal/server/`; `feature/DT-006_unit-integration-tests` |

> SharedLib unblock notes: ~~SHAREDLIB-010~~ ✅ and ~~SHAREDLIB-011~~ ✅ are
> done; `entitygraph.DataManager`, `SchemaManager`, and all associated models
> are available. **`SHAREDLIB-014`** (`EntityFilter` time-range + default
> ordering for `dt_telemetry` / `dt_events`) is open and required before
> FR-004 time-range telemetry queries are implementable; not a blocker for
> MVP-DT-001..MVP-DT-005.

---

## Success Criteria

| # | Criterion |
|---|---|
| 1 | `go build ./...` succeeds |
| 2 | `go test -race ./...` all pass |
| 3 | `go vet ./...` shows 0 issues |
| 4 | All 10 `DTService` RPCs work end-to-end with ArangoDB |
| 5 | `dt_relationships` created as **edge collection** (cannot be changed post-creation) |
| 6 | `dt_graph` named graph bootstrapped on startup (idempotent) |
| 7 | CodeValdCross registration fires on startup and repeats every 20 seconds |
| 8 | `DTDataManager` and `DTSchemaManager` injected via constructor — never hardcoded |
| 9 | No direct imports of other CodeVald services |
