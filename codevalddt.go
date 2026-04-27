// Package codevalddt provides the CodeValdDT digital-twin service. CodeValdDT
// stores agency-defined entity types, telemetry, and event records in the
// agency-scoped graph via
// [github.com/aosanya/CodeValdSharedLib/entitygraph]. Unlike services that
// pre-deliver a fixed schema (e.g. CodeValdWork), CodeValdDT does not bake any
// TypeDefinition into its own schema — agencies declare their entity types,
// telemetry channels, and event channels at runtime via the
// [DTSchemaManager].
//
// All persistence and CRUD is handled by the shared
// [github.com/aosanya/CodeValdSharedLib/entitygraph/server.EntityServer],
// re-exported under [internal/server]. Telemetry and events are routed to
// dedicated storage collections (`dt_telemetry`, `dt_events`) by setting
// [types.TypeDefinition.StorageCollection] on the agency-defined types — they
// are NOT separate Go types and have no separate gRPC service.
//
// Public API surface:
//   - [DTDataManager] / [DTSchemaManager] — entitygraph aliases (see models.go)
//   - [CrossPublisher] — best-effort lifecycle event publisher (see models.go)
//   - [DefaultDTSchema] — empty pre-delivered schema scaffold (see schema.go)
//   - [Err...] — typed errors (see errors.go)
package codevalddt
