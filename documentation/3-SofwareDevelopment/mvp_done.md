# MVP Done — Completed Tasks

Completed tasks are removed from `mvp.md` and recorded here with their completion date.

| Task ID | Title | Completion Date | Branch | Coding Session |
|---------|-------|-----------------|--------|----------------|
| MVP-DT-001 | Module Scaffolding | 2026-04-27 | `feature/DT-001_module-scaffolding` | Bundled commit absorbed DT-002 + DT-005 work — see notes below. |
| MVP-DT-002 | ArangoDB Backend | 2026-04-27 | `feature/DT-001_module-scaffolding` | Landed as `storage/arangodb/arangodb.go` — thin shim over `CodeValdSharedLib/entitygraph/arangodb` fixing `dt_*` collection / graph names. Indexes are inherited from the SharedLib `entitygraph/arangodb` bootstrap. |
| MVP-DT-005 | CodeValdCross Registration | 2026-04-27 | `feature/DT-001_module-scaffolding` | Landed as `internal/registrar/registrar.go`; SharedLib heartbeat at 20 s default; `Produces:` enumerates `cross.dt.{agencyID}.entity.created`, `…telemetry.recorded`, `…event.recorded`. |

> Architecture pivot deferred for review (does not affect the three rows above):
> the merged WIP re-uses the shared `entitygraph.EntityService` and re-exports
> the shared `EntityServer`, eliminating the need for the DT-specific proto and
> server originally specified in MVP-DT-003 / MVP-DT-004. Those two tasks
> remain in `mvp.md` flagged for a follow-up decision.
