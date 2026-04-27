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
	"log"
	"time"

	codevalddt "github.com/aosanya/CodeValdDT"
	egserver "github.com/aosanya/CodeValdSharedLib/entitygraph/server"
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

// Publish implements [codevalddt.CrossPublisher].
// It fires a best-effort notification for topic and agencyID.
// Currently logs the event; a future iteration will call a Cross Publish RPC
// once CodeValdCross exposes one. Errors are always nil — the DT operation
// has already been persisted and must not be rolled back.
func (r *Registrar) Publish(ctx context.Context, topic string, agencyID string) error {
	log.Printf("registrar[codevalddt]: publish topic=%q agencyID=%q", topic, agencyID)
	// TODO(CROSS-007): call OrchestratorService.Publish RPC when available.
	return nil
}

// dtRoutes returns the HTTP routes that CodeValdDT exposes via Cross.
//
// All routes are derived dynamically from [codevalddt.DefaultDTSchema] using
// [schemaroutes.RoutesFromSchema]. DT has no service-specific gRPC service —
// every route maps to the shared EntityService at
// [egserver.GRPCServicePath]. The default schema is empty, so the initial
// route set is empty; types are added by the agency at runtime.
func dtRoutes() []types.RouteInfo {
	return schemaroutes.RoutesFromSchema(
		codevalddt.DefaultDTSchema(),
		"/dt/{agencyId}",
		"agencyId",
		egserver.GRPCServicePath,
	)
}
