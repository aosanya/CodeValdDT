package codevalddt

import (
	"context"

	"github.com/aosanya/CodeValdSharedLib/entitygraph"
)

// DTDataManager is the CodeValdDT alias for [entitygraph.DataManager].
// gRPC handlers hold this interface — never the concrete type. The same
// interface backs both `dt_entities` writes and the routed `dt_telemetry`
// and `dt_events` writes; storage routing is driven by the resolved
// [types.TypeDefinition.StorageCollection].
type DTDataManager = entitygraph.DataManager

// DTSchemaManager is the CodeValdDT alias for [entitygraph.SchemaManager].
// cmd/main.go constructs the concrete implementation (e.g.
// arangodb.NewBackend) and injects it into the [DTDataManager].
type DTSchemaManager = entitygraph.SchemaManager

// CrossPublisher publishes digital-twin lifecycle events to CodeValdCross.
// Implementations must be safe for concurrent use. A nil CrossPublisher is
// valid — Publish calls are silently skipped by callers.
type CrossPublisher interface {
	// Publish delivers an event for the given topic and agencyID to
	// CodeValdCross. Errors are non-fatal: implementations should log and
	// return nil for best-effort delivery — the entity is already persisted.
	Publish(ctx context.Context, topic string, agencyID string) error
}
