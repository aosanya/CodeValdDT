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

> All in-scope MVP rows are complete or withdrawn — see [mvp_done.md](mvp_done.md).
> **`SHAREDLIB-014`** (`EntityFilter` time-range + default ordering for
> `dt_telemetry` / `dt_events`) remains open and is required before FR-004
> time-range telemetry queries can be exercised end-to-end. **FR-008** (DTDL v3
> export) and the parked open questions in
> [requirements.md §5](../1-SoftwareRequirements/requirements.md) (telemetry
> retention TTL, `TraverseGraph` max-depth ceiling) are candidates for the next
> task batch.

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
