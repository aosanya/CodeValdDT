# CodeValdDT

Digital-twin store for CodeValdCortex agencies. Entities, telemetry, and
event records are persisted as typed entities in the agency-scoped graph via
[CodeValdSharedLib/entitygraph](../CodeValdSharedLib/entitygraph).

CodeValdDT does NOT pre-deliver any TypeDefinition — agencies declare their
own entity, telemetry, and event types at runtime via the SchemaManager.
Telemetry / event records are routed to dedicated storage collections by
setting `TypeDefinition.StorageCollection` to `dt_telemetry` / `dt_events` —
they are not separate Go types and have no separate gRPC service.

## Layout

- `codevalddt.go`, `models.go`, `errors.go`, `schema.go` — public API surface.
- `cmd/server`, `cmd/dev` — slim shims that delegate to `internal/app.Run`.
- `internal/app`, `internal/config` — bootstrap wiring; configuration loaded
  from env vars (see `internal/config/config.go`).
- `internal/registrar` — Cross heartbeat + `CrossPublisher` for
  `cross.dt.{agencyID}.entity.created`,
  `cross.dt.{agencyID}.telemetry.recorded`,
  `cross.dt.{agencyID}.event.recorded`.
- `internal/server` — re-export of the shared `EntityServer` (DT has no
  service-specific gRPC service).
- `storage/arangodb` — thin shim over
  [`CodeValdSharedLib/entitygraph/arangodb`](../CodeValdSharedLib/entitygraph/arangodb)
  fixing the `dt_*` collection / graph names.

## Local dev

```sh
make build         # compile everything
make test          # unit tests
make test-arango   # integration tests (requires ArangoDB; reads .env)
make dev           # build + run with .env loaded
```
