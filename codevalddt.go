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
package codevalddt

import (
	"context"

	"github.com/aosanya/CodeValdSharedLib/entitygraph"
)

// DTSchemaManager is a type alias for [entitygraph.SchemaManager].
// Used by internal/app to seed [DefaultDTSchema] (an empty schema scaffold) so
// that the agency is bootstrapped and ready to accept agency-defined types.
type DTSchemaManager = entitygraph.SchemaManager

// CrossPublisher publishes DT lifecycle events to CodeValdCross.
// Implementations must be safe for concurrent use. A nil CrossPublisher is
// valid — publish calls are silently skipped.
type CrossPublisher interface {
	// Publish delivers an event for the given topic and agencyID to
	// CodeValdCross. Errors are non-fatal: implementations should log and
	// return nil for best-effort delivery.
	Publish(ctx context.Context, topic string, agencyID string) error
}
