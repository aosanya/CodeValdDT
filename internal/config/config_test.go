package config_test

import (
	"testing"
	"time"

	"github.com/aosanya/CodeValdDT/internal/config"
)

// TestLoad_Defaults verifies that with only CODEVALDDT_PORT set, every other
// field falls back to its documented default.
func TestLoad_Defaults(t *testing.T) {
	t.Setenv("CODEVALDDT_PORT", "50055")
	t.Setenv("DT_ARANGO_ENDPOINT", "")
	t.Setenv("DT_ARANGO_USER", "")
	t.Setenv("DT_ARANGO_PASSWORD", "")
	t.Setenv("DT_ARANGO_DATABASE", "")
	t.Setenv("CROSS_GRPC_ADDR", "")
	t.Setenv("DT_GRPC_ADVERTISE_ADDR", "")
	t.Setenv("CODEVALDDT_AGENCY_ID", "")
	t.Setenv("CROSS_PING_INTERVAL", "")
	t.Setenv("CROSS_PING_TIMEOUT", "")

	got := config.Load()

	want := config.Config{
		GRPCPort:       "50055",
		ArangoEndpoint: "http://localhost:8529",
		ArangoUser:     "root",
		ArangoPassword: "",
		ArangoDatabase: "codevalddt",
		CrossGRPCAddr:  "",
		AdvertiseAddr:  ":50055",
		AgencyID:       "",
		PingInterval:   20 * time.Second,
		PingTimeout:    5 * time.Second,
	}

	if got != want {
		t.Errorf("Load with defaults:\n got  %+v\n want %+v", got, want)
	}
}

// TestLoad_Overrides verifies that every env var override flows through to the
// matching Config field.
func TestLoad_Overrides(t *testing.T) {
	t.Setenv("CODEVALDDT_PORT", "60001")
	t.Setenv("DT_ARANGO_ENDPOINT", "http://arangodb:9000")
	t.Setenv("DT_ARANGO_USER", "dtuser")
	t.Setenv("DT_ARANGO_PASSWORD", "secret")
	t.Setenv("DT_ARANGO_DATABASE", "codevald_demo")
	t.Setenv("CROSS_GRPC_ADDR", "cross:50050")
	t.Setenv("DT_GRPC_ADVERTISE_ADDR", "dt:60001")
	t.Setenv("CODEVALDDT_AGENCY_ID", "agency-123")
	t.Setenv("CROSS_PING_INTERVAL", "7s")
	t.Setenv("CROSS_PING_TIMEOUT", "750ms")

	got := config.Load()

	want := config.Config{
		GRPCPort:       "60001",
		ArangoEndpoint: "http://arangodb:9000",
		ArangoUser:     "dtuser",
		ArangoPassword: "secret",
		ArangoDatabase: "codevald_demo",
		CrossGRPCAddr:  "cross:50050",
		AdvertiseAddr:  "dt:60001",
		AgencyID:       "agency-123",
		PingInterval:   7 * time.Second,
		PingTimeout:    750 * time.Millisecond,
	}

	if got != want {
		t.Errorf("Load with overrides:\n got  %+v\n want %+v", got, want)
	}
}

// TestLoad_AdvertiseAddrDefaultsToPort verifies that when DT_GRPC_ADVERTISE_ADDR
// is unset the loader synthesises ":<GRPCPort>" — the address Cross will dial
// back on. This is documented in [config.Config.AdvertiseAddr].
func TestLoad_AdvertiseAddrDefaultsToPort(t *testing.T) {
	t.Setenv("CODEVALDDT_PORT", "55555")
	t.Setenv("DT_GRPC_ADVERTISE_ADDR", "")

	got := config.Load()

	if got.AdvertiseAddr != ":55555" {
		t.Errorf("AdvertiseAddr default: got %q, want %q", got.AdvertiseAddr, ":55555")
	}
}

// TestLoad_InvalidDurationsFallBackToDefaults verifies the documented behaviour
// of serverutil.ParseDurationString — bad values log and fall back, they don't
// fail the loader.
func TestLoad_InvalidDurationsFallBackToDefaults(t *testing.T) {
	t.Setenv("CODEVALDDT_PORT", "50055")
	t.Setenv("CROSS_PING_INTERVAL", "not-a-duration")
	t.Setenv("CROSS_PING_TIMEOUT", "")

	got := config.Load()

	if got.PingInterval != 20*time.Second {
		t.Errorf("PingInterval invalid fallback: got %v, want 20s", got.PingInterval)
	}
	if got.PingTimeout != 5*time.Second {
		t.Errorf("PingTimeout default: got %v, want 5s", got.PingTimeout)
	}
}
