// Package registrar provides the CodeValdDT service registrar.
// It wraps the shared-library heartbeat registrar and additionally implements
// [codevalddt.CrossPublisher] so DT subsystems can notify CodeValdCross
// whenever an entity, telemetry record, or event record is created.
//
// Construct via [New]; start heartbeats by calling Run in a goroutine; stop
// by cancelling the context then calling Close.
package registrar

import (
	"context"
	"encoding/json"
	"log"
	"time"

	codevalddt "github.com/aosanya/CodeValdDT"
	egserver "github.com/aosanya/CodeValdSharedLib/entitygraph/server"
	"github.com/aosanya/CodeValdSharedLib/eventbus"
	sharedregistrar "github.com/aosanya/CodeValdSharedLib/registrar"
	"github.com/aosanya/CodeValdSharedLib/schemaroutes"
	"github.com/aosanya/CodeValdSharedLib/types"
)

// Registrar handles two responsibilities:
//  1. Sending periodic heartbeat registrations to CodeValdCross via the
//     shared-library registrar (Run / Close).
//  2. Implementing [codevalddt.CrossPublisher] so DT subsystems can fire
//     lifecycle events on successful operations.
type Registrar struct {
	heartbeat sharedregistrar.Registrar
	agencyID  string
}

// Compile-time assertion that *Registrar implements codevalddt.CrossPublisher.
var _ codevalddt.CrossPublisher = (*Registrar)(nil)

// New constructs a Registrar that heartbeats to the CodeValdCross gRPC server
// at crossAddr and can publish DT lifecycle events.
//
//   - crossAddr     — host:port of the CodeValdCross gRPC server
//   - advertiseAddr — host:port that Cross dials back on
//   - agencyID      — agency this instance serves (may be empty)
//   - pingInterval  — heartbeat cadence; ≤ 0 means only the initial ping
//   - pingTimeout   — per-RPC timeout for each Register call
//
// CodeValdDT produces three topic families, all scoped by agency:
//
//	cross.dt.{agencyID}.entity.created
//	cross.dt.{agencyID}.telemetry.recorded
//	cross.dt.{agencyID}.event.recorded
//
// It currently consumes no topics.
//
// Routes come solely from [schemaroutes.RoutesFromSchema] against
// [codevalddt.DefaultDTSchema] — DT has no service-specific gRPC service.
// Because the default schema is empty, the initial route set is empty too.
// Once the agency declares its types, the next heartbeat (or restart) will
// pick up the additional routes if/when the registrar is re-derived from a
// live schema. (For now, agency-time route rediscovery is out of scope.)
func New(
	crossAddr, advertiseAddr, agencyID string,
	pingInterval, pingTimeout time.Duration,
) (*Registrar, error) {
	produces := []string{
		"cross.dt." + agencyID + ".entity.created",
		"cross.dt." + agencyID + ".telemetry.recorded",
		"cross.dt." + agencyID + ".event.recorded",
	}

	hb, err := sharedregistrar.New(
		crossAddr,
		advertiseAddr,
		agencyID,
		"codevalddt",
		produces,
		[]string{},
		dtRoutes(),
		pingInterval,
		pingTimeout,
	)
	if err != nil {
		return nil, err
	}
	return &Registrar{heartbeat: hb, agencyID: agencyID}, nil
}

// Run starts the heartbeat loop, sending an immediate Register ping to
// CodeValdCross then repeating at the configured interval until ctx is
// cancelled. Must be called inside a goroutine.
func (r *Registrar) Run(ctx context.Context) {
	r.heartbeat.Run(ctx)
}

// Close releases the underlying gRPC connection used for heartbeats.
// Call after the context passed to Run has been cancelled.
func (r *Registrar) Close() {
	r.heartbeat.Close()
}

// Publish implements [eventbus.Publisher].
// Marshals the event payload to JSON and forwards it to CodeValdCross via the
// OrchestratorService.Publish RPC, which routes it on to CodeValdPubSub.
// Errors are logged but not returned — the operation is already persisted.
func (r *Registrar) Publish(ctx context.Context, e eventbus.Event) error {
	payload, err := json.Marshal(e.Payload)
	if err != nil {
		log.Printf("registrar[codevalddt]: marshal payload for topic=%q: %v", e.Topic, err)
		payload = []byte("{}")
	}
	if err := r.heartbeat.Publish(ctx, e.AgencyID, e.Topic, "codevalddt", string(payload)); err != nil {
		log.Printf("registrar[codevalddt]: Publish topic=%q: %v", e.Topic, err)
	}
	return nil
}

// dtRoutes returns the HTTP routes that CodeValdDT exposes via Cross.
//
// Schema-driven routes come from [schemaroutes.RoutesFromSchema]; because the
// default DT schema is empty, the initial set is empty and grows as the agency
// declares types at runtime.
//
// The DTDL export route is manually declared here because it is a
// service-level route, not tied to any TypeDefinition.
func dtRoutes() []types.RouteInfo {
	routes := schemaroutes.RoutesFromSchema(
		codevalddt.DefaultDTSchema(),
		"/dt/{agencyId}",
		"agencyId",
		egserver.GRPCServicePath,
	)
	routes = append(routes, types.RouteInfo{
		Method:     "GET",
		Pattern:    "/{agencyId}/dt/schema/dtdl",
		Capability: "export_dtdl",
	})
	return routes
}
