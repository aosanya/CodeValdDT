# MVP — CodeValdDT

## Goal

Deliver a production-ready digital twin graph gRPC microservice with ArangoDB
persistence and CodeValdCross registration, backed by `entitygraph.DataManager`
and `entitygraph.SchemaManager` from CodeValdSharedLib.

---

## Task List

| Task ID | Title | Status | Depends On | Notes |
|---|---|---|---|---|
| ~~MVP-DT-003~~ | ~~gRPC Service Proto & Codegen~~ | ❌ Withdrawn (2026-04-27) | — | **Retired**: CodeValdDT exposes the shared `entitygraph.EntityService` (from CodeValdSharedLib) — agencies declare their own `TypeDefinition`s at runtime and storage routing is driven by `TypeDefinition.StorageCollection`, so a DT-specific proto would duplicate the shared surface. See [architecture-flows.md §9](../2-SoftwareDesignAndArchitecture/architecture-flows.md). |
| ~~MVP-DT-004~~ | ~~gRPC Server Implementation~~ | ❌ Withdrawn (2026-04-27) | — | **Retired**: `internal/server/entity_server.go` re-exports `egserver.NewEntityServer`, and [`internal/app/app.go`](../../internal/app/app.go) registers `entitygraphpb.RegisterEntityServiceServer`. Handler logic, `toGRPCError`, and entity↔proto conversion all live in `CodeValdSharedLib/entitygraph/server`. |
| MVP-DT-006 | Unit & Integration Tests | 🚧 In Progress | ~~MVP-DT-001~~ ✅, ~~MVP-DT-002~~ ✅ | Unit tests on DT-specific code (`internal/config`, `internal/registrar`, `storage/arangodb`, `schema.go`); integration test tagged `//go:build integration` boots `app.Run` against a real ArangoDB and exercises the EntityService gRPC surface end-to-end (entity CRUD, soft-delete exclusion, telemetry routing to `dt_telemetry` with `Immutable` enforcement, relationship + traversal). EntityService internals are tested in CodeValdSharedLib, not here. Branch: `feature/DT-006_unit-integration-tests` |

> SharedLib unblock notes: ~~SHAREDLIB-010~~ ✅ and ~~SHAREDLIB-011~~ ✅ are
> done; `entitygraph.DataManager`, `SchemaManager`, and all associated models
> are available. **`SHAREDLIB-014`** (`EntityFilter` time-range + default
> ordering for `dt_telemetry` / `dt_events`) is open and required before
> FR-004 time-range telemetry queries are implementable; not a blocker for
> MVP-DT-006.

---

## Success Criteria

| # | Criterion |
|---|---|
| 1 | `go build ./...` succeeds |
| 2 | `go test -race ./...` all pass |
| 3 | `go vet ./...` shows 0 issues |
| 4 | EntityService RPCs (from CodeValdSharedLib) work end-to-end against ArangoDB through `app.Run`, with `CreateEntity` routing to `dt_entities` / `dt_telemetry` / `dt_events` based on the resolved `TypeDefinition.StorageCollection` and `UpdateEntity` returning `FailedPrecondition` for immutable types |
| 5 | `dt_relationships` created as **edge collection** (cannot be changed post-creation) |
| 6 | `dt_graph` named graph bootstrapped on startup (idempotent) |
| 7 | CodeValdCross registration fires on startup and repeats every 20 seconds |
| 8 | `DTDataManager` and `DTSchemaManager` injected via constructor — never hardcoded |
| 9 | No direct imports of other CodeVald services |
