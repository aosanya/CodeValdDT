# MVP — CodeValdDT

## Goal

Deliver a production-ready digital twin graph gRPC microservice with ArangoDB
persistence and CodeValdCross registration, backed by `entitygraph.DataManager`
and `entitygraph.SchemaManager` from CodeValdSharedLib.

---

## Task List

| Task ID | Title | Status | Depends On | Notes |
|---|---|---|---|---|
| MVP-DT-001 | Module Scaffolding | ⏸️ Blocked | SHAREDLIB-010 | `go.mod`, `errors.go`, `models.go`, `internal/config/`, `cmd/main.go` skeleton; `feature/DT-001_module-scaffolding` |
| MVP-DT-002 | ArangoDB Backend | ⏸️ Blocked | MVP-DT-001 | `storage/arangodb/storage.go`; bootstrap `dt_entities`, `dt_relationships` (**edge**), `dt_telemetry`, `dt_events`, `dt_schemas`; named graph `dt_graph`; `feature/DT-002_arangodb-backend` |
| MVP-DT-003 | gRPC Service Proto & Codegen | ⏸️ Blocked | MVP-DT-001 | `proto/codevalddt/v1/dt.proto`; 10 RPCs; `buf generate` → `gen/go/`; `feature/DT-003_grpc-proto-codegen` |
| MVP-DT-004 | gRPC Server Implementation | ⏸️ Blocked | MVP-DT-001, MVP-DT-002, MVP-DT-003 | `internal/server/server.go`; thin handlers delegate to `DTDataManager`; `toGRPCError` mapping; `feature/DT-004_grpc-server-implementation` |
| MVP-DT-005 | CodeValdCross Registration | ⏸️ Blocked | MVP-DT-004 | SharedLib `registrar`; heartbeat every 20 s; full `cmd/main.go` wiring via `serverutil`; `feature/DT-005_cross-registration` |
| MVP-DT-006 | Unit & Integration Tests | ⏸️ Blocked | MVP-DT-001, MVP-DT-002, MVP-DT-004 | Table-driven unit tests with mock `DTDataManager`; integration tests tagged `//go:build integration`; coverage ≥ 80% on `internal/server/`; `feature/DT-006_unit-integration-tests` |

> All tasks blocked on `SHAREDLIB-010` — see [CodeValdSharedLib mvp.md](../../../CodeValdSharedLib/documentation/3-SofwareDevelopment/mvp.md).

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
