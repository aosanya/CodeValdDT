// Package codevalddt — pre-delivered schema scaffold.
//
// This file exposes [DefaultDTSchema], which returns the fixed [types.Schema]
// scaffold for CodeValdDT. The schema is seeded idempotently on startup via
// entitygraph.SeedSchema (see internal/app) so the agency is bootstrapped and
// ready to accept agency-defined TypeDefinitions at runtime.
//
// Unlike CodeValdWork, CodeValdDT does NOT pre-bake any TypeDefinition.
// CodeValdDT is the digital-twin store: agencies declare their own entity
// types, telemetry channels, and event channels at runtime via the
// [DTSchemaManager]. Telemetry and events are routed to dedicated storage
// collections (`dt_telemetry`, `dt_events`) by setting
// [types.TypeDefinition.StorageCollection] on the agency-defined types —
// there are no separate Go types or gRPC services for telemetry/events.
//
// Entities without a per-type StorageCollection fall back to `dt_entities`.
// All edges live in the `dt_relationships` edge collection. The named graph
// is `dt_graph`.
package codevalddt

import "github.com/aosanya/CodeValdSharedLib/types"

// DefaultDTSchema returns the pre-delivered [types.Schema] seeded on startup.
// The schema is intentionally empty — DT types are agency-defined dynamically
// via the [DTSchemaManager]. Seeding the empty scaffold is idempotent and
// activates the agency so subsequent SetSchema / Publish / Activate calls
// against the [DTSchemaManager] succeed.
func DefaultDTSchema() types.Schema {
	return types.Schema{
		ID:      "dt-schema-v1",
		Version: 1,
		Tag:     "v1",
		Types:   nil,
	}
}
