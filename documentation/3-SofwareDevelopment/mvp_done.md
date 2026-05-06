# MVP Done — Completed Tasks

Completed tasks are removed from `mvp.md` and recorded here with their completion date.

| Task ID | Title | Completion Date | Branch | Coding Session |
|---------|-------|-----------------|--------|----------------|
| MVP-DT-001 | Module Scaffolding | 2026-04-27 | `feature/DT-001_module-scaffolding` | Bundled commit absorbed DT-002 + DT-005 work — see notes below. |
| MVP-DT-002 | ArangoDB Backend | 2026-04-27 | `feature/DT-001_module-scaffolding` | Landed as `storage/arangodb/arangodb.go` — thin shim over `CodeValdSharedLib/entitygraph/arangodb` fixing `dt_*` collection / graph names. Indexes are inherited from the SharedLib `entitygraph/arangodb` bootstrap. |
| MVP-DT-005 | CodeValdCross Registration | 2026-04-27 | `feature/DT-001_module-scaffolding` | Landed as `internal/registrar/registrar.go`; SharedLib heartbeat at 20 s default; `Produces:` enumerates `cross.dt.{agencyID}.entity.created`, `…telemetry.recorded`, `…event.recorded`. |
| MVP-DT-006 | Unit & Integration Tests | 2026-04-28 | `feature/DT-006_unit-integration-tests` | Unit tests on DT-specific code (`internal/config`, `internal/registrar`, `storage/arangodb`, `schema.go`); integration test tagged `//go:build integration` in `internal/app/app_integration_test.go` boots `app.Run` against a real ArangoDB and exercises the EntityService gRPC surface end-to-end. `go build`, `go vet`, and `go test -race ./...` all clean on `main` (merge commit `dfe6ff2`). |
| DT-007 | DTDL v3 Export Endpoint | 2026-05-06 | `feature/DT-007_dtdl-export-and-depth-ceiling` | `internal/dtdl/export.go` — pure `ExportSchema(agencyID, schema)` conversion (TypeDef→DTDL Interface; telemetry/events collections→`Telemetry` items; properties→`Property writable:true`; relationships→`Relationship` with populated `target`). `internal/httphandler/handler.go` — `GET /{agencyId}/dt/schema/dtdl` via cmux HTTP listener. `app.go` switched from plain gRPC listener to cmux (gRPC + HTTP on same port). DTDL export route registered in registrar. 6 export unit tests + updated registrar test. |
| DT-008 | TraverseGraph Max-Depth Ceiling | 2026-05-06 | `feature/DT-007_dtdl-export-and-depth-ceiling` | `internal/server/interceptor.go` — `TraverseDepthInterceptor(maxDepth int32)` gRPC unary interceptor; `MaxTraverseDepth = 10`. `app.go` switched from `serverutil.NewGRPCServer()` to `grpc.NewServer(grpc.ChainUnaryInterceptor(...))` + `reflection.Register`. Returns `codes.InvalidArgument` when `Depth > 10`; passes through at or below limit and for all other methods. 5 interceptor unit tests. |

> Architecture pivot resolved: the merged implementation re-uses the shared
> `entitygraph.EntityService` and re-exports the shared `EntityServer`,
> eliminating the need for the DT-specific proto and server originally
> specified in MVP-DT-003 / MVP-DT-004. Those two tasks were formally withdrawn
> on 2026-04-27 (see [mvp.md](mvp.md)).
