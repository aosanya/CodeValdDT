//go:build integration

// Package app_test — end-to-end test of the CodeValdDT runtime wiring against a
// real ArangoDB instance. Mirrors the dependency graph that internal/app.Run
// constructs (ArangoDB Backend → EntityServer → grpc.Server) so a regression in
// the wiring path that app.Run takes is caught here.
//
// Run with:
//
//	DT_INTEGRATION_DATABASE=codevald_dt_test go test -tags=integration ./internal/app/...
//
// Optional env vars: DT_ARANGO_ENDPOINT (default "http://localhost:8529"),
// DT_ARANGO_USER (default "root"), DT_ARANGO_PASSWORD.
package app_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	driver "github.com/arangodb/go-driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/aosanya/CodeValdDT/internal/server"
	dtarangodb "github.com/aosanya/CodeValdDT/storage/arangodb"
	"github.com/aosanya/CodeValdSharedLib/arangoutil"
	"github.com/aosanya/CodeValdSharedLib/entitygraph"
	pb "github.com/aosanya/CodeValdSharedLib/gen/go/entitygraph/v1"
	"github.com/aosanya/CodeValdSharedLib/types"
)

// testRig bundles everything an integration test needs: a connected gRPC
// client, the ArangoDB handle (for direct collection inspection), and the
// agencyID this test run was scoped to.
type testRig struct {
	client   pb.EntityServiceClient
	db       driver.Database
	agencyID string
}

// setupRig connects to ArangoDB, seeds a test schema for a unique agency, and
// stands up a gRPC server backed by the same backend that internal/app.Run
// uses. Skips the test cleanly when DT_INTEGRATION_DATABASE is unset.
func setupRig(t *testing.T) *testRig {
	t.Helper()

	dbName := os.Getenv("DT_INTEGRATION_DATABASE")
	if dbName == "" {
		t.Skip("DT_INTEGRATION_DATABASE not set; skipping integration test")
	}

	endpoint := envOr("DT_ARANGO_ENDPOINT", "http://localhost:8529")
	user := envOr("DT_ARANGO_USER", "root")
	password := os.Getenv("DT_ARANGO_PASSWORD")

	connCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := arangoutil.Connect(connCtx, arangoutil.Config{
		Endpoint: endpoint,
		Username: user,
		Password: password,
		Database: dbName,
	})
	if err != nil {
		t.Skipf("ArangoDB unreachable at %s: %v", endpoint, err)
	}

	agencyID := fmt.Sprintf("dt-itest-%d", time.Now().UnixNano())
	schema := testSchema(agencyID)

	dm, sm, err := dtarangodb.New(db, schema)
	if err != nil {
		t.Fatalf("dtarangodb.New: %v", err)
	}

	// Seed the schema (SetSchema → Publish → Activate(1)) for this agency only.
	seedCtx, seedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer seedCancel()
	if err := entitygraph.SeedSchema(seedCtx, sm, agencyID, schema); err != nil {
		t.Fatalf("SeedSchema: %v", err)
	}

	// Boot the gRPC server on a random port — same wiring as internal/app.Run.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterEntityServiceServer(grpcServer, server.NewEntityServer(dm))

	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		_ = grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		grpcServer.GracefulStop()
		t.Fatalf("grpc dial: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		grpcServer.GracefulStop()
		<-serveDone
	})

	return &testRig{
		client:   pb.NewEntityServiceClient(conn),
		db:       db,
		agencyID: agencyID,
	}
}

// testSchema returns the agency-scoped schema used by the integration tests:
// a Pump entity (default storage), a TempReading routed to dt_telemetry with
// Immutable=true, and a self-referential connects_to relationship on Pump.
func testSchema(agencyID string) types.Schema {
	return types.Schema{
		ID:       "dt-itest-schema",
		AgencyID: agencyID,
		Version:  1,
		Tag:      "v1",
		Types: []types.TypeDefinition{
			{
				Name:        "Pump",
				DisplayName: "Pump",
				Properties: []types.PropertyDefinition{
					{Name: "label", Type: types.PropertyTypeString},
					{Name: "pressure", Type: types.PropertyTypeNumber},
				},
				Relationships: []types.RelationshipDefinition{
					{Name: "connects_to", ToType: "Pump", ToMany: true},
				},
			},
			{
				Name:              "TempReading",
				DisplayName:       "Temperature Reading",
				StorageCollection: "dt_telemetry",
				Immutable:         true,
				Properties: []types.PropertyDefinition{
					{Name: "entityID", Type: types.PropertyTypeString},
					{Name: "value", Type: types.PropertyTypeNumber},
					{Name: "timestamp", Type: types.PropertyTypeString},
				},
			},
		},
	}
}

// TestEntityCRUD covers the FR-001 happy path through the gRPC surface:
// CreateEntity → GetEntity → UpdateEntity → DeleteEntity is reflected in
// ListEntities and a follow-up GetEntity returns NotFound.
func TestEntityCRUD(t *testing.T) {
	rig := setupRig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	created, err := rig.client.CreateEntity(ctx, &pb.CreateEntityRequest{
		AgencyId:   rig.agencyID,
		TypeId:     "Pump",
		Properties: mustStruct(t, map[string]any{"label": "P1", "pressure": 4.2}),
	})
	if err != nil {
		t.Fatalf("CreateEntity Pump: %v", err)
	}
	if created.GetId() == "" {
		t.Fatal("CreateEntity: returned empty ID")
	}

	got, err := rig.client.GetEntity(ctx, &pb.GetEntityRequest{
		AgencyId: rig.agencyID,
		EntityId: created.GetId(),
	})
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}
	if got.GetTypeId() != "Pump" {
		t.Errorf("GetEntity TypeId: got %q, want %q", got.GetTypeId(), "Pump")
	}

	updated, err := rig.client.UpdateEntity(ctx, &pb.UpdateEntityRequest{
		AgencyId:   rig.agencyID,
		EntityId:   created.GetId(),
		Properties: mustStruct(t, map[string]any{"label": "P1", "pressure": 7.1}),
	})
	if err != nil {
		t.Fatalf("UpdateEntity: %v", err)
	}
	if got, want := updated.GetProperties().AsMap()["pressure"], 7.1; got != want {
		t.Errorf("UpdateEntity pressure: got %v, want %v", got, want)
	}

	if _, err := rig.client.DeleteEntity(ctx, &pb.DeleteEntityRequest{
		AgencyId: rig.agencyID,
		EntityId: created.GetId(),
	}); err != nil {
		t.Fatalf("DeleteEntity: %v", err)
	}

	_, err = rig.client.GetEntity(ctx, &pb.GetEntityRequest{
		AgencyId: rig.agencyID,
		EntityId: created.GetId(),
	})
	if code := status.Code(err); code != codes.NotFound {
		t.Errorf("GetEntity after delete: got code %v, want NotFound", code)
	}

	list, err := rig.client.ListEntities(ctx, &pb.ListEntitiesRequest{
		AgencyId: rig.agencyID,
		TypeId:   "Pump",
	})
	if err != nil {
		t.Fatalf("ListEntities: %v", err)
	}
	for _, e := range list.GetEntities() {
		if e.GetId() == created.GetId() {
			t.Errorf("ListEntities: soft-deleted entity %q still returned", created.GetId())
		}
	}
}

// TestTelemetryRoutingAndImmutability covers FR-004 and the
// TypeDefinition.Immutable rule on Update: telemetry-routed entities land in
// dt_telemetry (not dt_entities), and UpdateEntity rejects them with
// FailedPrecondition.
func TestTelemetryRoutingAndImmutability(t *testing.T) {
	rig := setupRig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reading, err := rig.client.CreateEntity(ctx, &pb.CreateEntityRequest{
		AgencyId: rig.agencyID,
		TypeId:   "TempReading",
		Properties: mustStruct(t, map[string]any{
			"entityID":  "ext-pump-1",
			"value":     42.5,
			"timestamp": "2026-04-27T00:00:00Z",
		}),
	})
	if err != nil {
		t.Fatalf("CreateEntity TempReading: %v", err)
	}

	// Direct collection check: the document must exist in dt_telemetry, not
	// in dt_entities. This pins FR-004 and Success Criterion #4 — storage
	// routing is driven by TypeDefinition.StorageCollection.
	if got := documentExists(ctx, t, rig.db, "dt_telemetry", reading.GetId()); !got {
		t.Errorf("dt_telemetry: expected document %q to exist", reading.GetId())
	}
	if got := documentExists(ctx, t, rig.db, "dt_entities", reading.GetId()); got {
		t.Errorf("dt_entities: expected no document %q (telemetry should not land here)", reading.GetId())
	}

	// UpdateEntity on an immutable type → FailedPrecondition.
	_, err = rig.client.UpdateEntity(ctx, &pb.UpdateEntityRequest{
		AgencyId:   rig.agencyID,
		EntityId:   reading.GetId(),
		Properties: mustStruct(t, map[string]any{"value": 99.9}),
	})
	if code := status.Code(err); code != codes.FailedPrecondition {
		t.Errorf("UpdateEntity on immutable type: got code %v, want FailedPrecondition", code)
	}
}

// TestRelationshipAndTraversal covers FR-002 + FR-003: relationships are
// persisted as edges and TraverseGraph reaches connected vertices.
func TestRelationshipAndTraversal(t *testing.T) {
	rig := setupRig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	a, err := rig.client.CreateEntity(ctx, &pb.CreateEntityRequest{
		AgencyId: rig.agencyID, TypeId: "Pump",
		Properties: mustStruct(t, map[string]any{"label": "A"}),
	})
	if err != nil {
		t.Fatalf("CreateEntity A: %v", err)
	}
	b, err := rig.client.CreateEntity(ctx, &pb.CreateEntityRequest{
		AgencyId: rig.agencyID, TypeId: "Pump",
		Properties: mustStruct(t, map[string]any{"label": "B"}),
	})
	if err != nil {
		t.Fatalf("CreateEntity B: %v", err)
	}

	rel, err := rig.client.CreateRelationship(ctx, &pb.CreateRelationshipRequest{
		AgencyId: rig.agencyID,
		EntityId: a.GetId(),
		Name:     "connects_to",
		ToId:     b.GetId(),
	})
	if err != nil {
		t.Fatalf("CreateRelationship: %v", err)
	}
	if rel.GetId() == "" {
		t.Fatal("CreateRelationship: empty ID")
	}

	trav, err := rig.client.TraverseGraph(ctx, &pb.TraverseGraphRequest{
		AgencyId:  rig.agencyID,
		StartId:   a.GetId(),
		Direction: "outbound",
		Depth:     1,
	})
	if err != nil {
		t.Fatalf("TraverseGraph: %v", err)
	}

	var sawB bool
	for _, v := range trav.GetVertices() {
		if v.GetId() == b.GetId() {
			sawB = true
			break
		}
	}
	if !sawB {
		t.Errorf("TraverseGraph: B (%q) not found in vertices: %+v", b.GetId(), trav.GetVertices())
	}
	if len(trav.GetEdges()) == 0 {
		t.Error("TraverseGraph: expected at least one edge")
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustStruct(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	if err != nil {
		t.Fatalf("structpb.NewStruct(%v): %v", m, err)
	}
	return s
}

// documentExists returns true when a document with the given _key exists in the
// named collection. Errors other than NotFound fail the test.
func documentExists(ctx context.Context, t *testing.T, db driver.Database, collection, key string) bool {
	t.Helper()
	col, err := db.Collection(ctx, collection)
	if err != nil {
		t.Fatalf("collection %q: %v", collection, err)
	}
	exists, err := col.DocumentExists(ctx, key)
	if err != nil {
		t.Fatalf("DocumentExists %q/%q: %v", collection, key, err)
	}
	return exists
}
