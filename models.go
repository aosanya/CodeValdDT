package codevalddt

import (
	"github.com/aosanya/CodeValdSharedLib/entitygraph"
	"github.com/aosanya/CodeValdSharedLib/eventbus"
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

// CrossPublisher is the historical name for the event-publishing contract
// CodeValdDT callers inject. As of MVP-WORK-014 it is a type alias for
// [eventbus.Publisher] — the SharedLib package that unifies the publish
// contract across CodeValdAgency, CodeValdComm, CodeValdDT, and CodeValdWork.
//
// New callers should refer to [eventbus.Publisher] directly; this alias
// remains for source compatibility.
type CrossPublisher = eventbus.Publisher
