package registrar

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	codevalddt "github.com/aosanya/CodeValdDT"
	"github.com/aosanya/CodeValdSharedLib/eventbus"
)

// TestNew_ReturnsRegistrar verifies that New succeeds with a valid-looking
// Cross address and produces a *Registrar that satisfies CrossPublisher.
func TestNew_ReturnsRegistrar(t *testing.T) {
	r, err := New("localhost:59001", ":50055", "agency-1", 10*time.Second, 1*time.Second)
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("New: expected non-nil Registrar")
	}
	defer r.Close()

	var _ codevalddt.CrossPublisher = r // compile-time assertion already in registrar.go
}

// TestRegistrar_PublishLogsTopic verifies the v1 best-effort contract: Publish
// returns nil and emits a log line carrying the topic and agencyID. Once
// CodeValdCross exposes a Publish RPC (CROSS-007) this test will need to fault
// in a fake server, but the public contract — nil error, observable side
// effect — must remain.
func TestRegistrar_PublishLogsTopic(t *testing.T) {
	r, err := New("localhost:59002", ":50055", "agency-7", 10*time.Second, 1*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer r.Close()

	var buf bytes.Buffer
	originalOut := log.Writer()
	originalFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(originalOut)
		log.SetFlags(originalFlags)
	})

	topic := "cross.dt.agency-7.entity.created"
	if err := r.Publish(context.Background(), eventbus.Event{Topic: topic, AgencyID: "agency-7"}); err != nil {
		t.Fatalf("Publish: unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, topic) {
		t.Errorf("Publish log missing topic %q; got %q", topic, got)
	}
	if !strings.Contains(got, "agency-7") {
		t.Errorf("Publish log missing agencyID; got %q", got)
	}
}

// TestRegistrar_RunExitsOnContextCancel verifies that Run returns after the
// context is cancelled, even with a Cross address that nothing listens on.
// Each ping fails silently and the heartbeat loop continues until ctx is done.
func TestRegistrar_RunExitsOnContextCancel(t *testing.T) {
	r, err := New("localhost:59003", ":50055", "agency-x", 50*time.Millisecond, 25*time.Millisecond)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer r.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.Run(ctx)
	}()

	cancel()
	select {
	case <-done:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit within 2 s after context cancellation")
	}
}

// TestDTRoutes_ContainsDTDLExport pins that dtRoutes always includes the
// service-level DTDL export route regardless of the schema content. Schema-
// derived routes start empty (DefaultDTSchema has no TypeDefinitions) and grow
// as the agency declares types at runtime; the DTDL route is unconditional.
func TestDTRoutes_ContainsDTDLExport(t *testing.T) {
	routes := dtRoutes()

	var found bool
	for _, r := range routes {
		if r.Method == "GET" && r.Pattern == "/{agencyId}/dt/schema/dtdl" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("dtRoutes(): DTDL export route not found in %+v", routes)
	}
}
