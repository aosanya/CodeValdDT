// Package config loads CodeValdDT runtime configuration from environment
// variables.
package config

import (
	"time"

	"github.com/aosanya/CodeValdSharedLib/serverutil"
)

// Config holds all runtime configuration for the CodeValdDT service.
type Config struct {
	// GRPCPort is the port the gRPC server listens on (required, set in .env).
	GRPCPort string

	// ArangoEndpoint is the ArangoDB HTTP endpoint
	// (default "http://localhost:8529").
	ArangoEndpoint string

	// ArangoUser is the ArangoDB username (default "root").
	ArangoUser string

	// ArangoPassword is the ArangoDB password.
	ArangoPassword string

	// ArangoDatabase is the pre-existing ArangoDB database name
	// (e.g. "codevald_demo"). CodeValdDT does NOT create the database — it
	// connects to an existing one and ensures its collections exist on boot.
	ArangoDatabase string

	// CrossGRPCAddr is the CodeValdCross gRPC address for registration
	// heartbeats and event publishing. Empty string disables registration.
	CrossGRPCAddr string

	// AdvertiseAddr is the address CodeValdCross dials back on
	// (default ":GRPCPort").
	AdvertiseAddr string

	// AgencyID is the agency identifier sent in every Register heartbeat to
	// CodeValdCross. Required when CrossGRPCAddr is set.
	AgencyID string

	// PingInterval is the heartbeat cadence sent to CodeValdCross
	// (default 20s — the platform liveness contract).
	PingInterval time.Duration

	// PingTimeout is the per-RPC timeout for each Register call (default 5s).
	PingTimeout time.Duration
}

// Load reads configuration from environment variables, falling back to
// defaults for any variable that is unset or empty.
//
// Required:
//   - CODEVALDDT_GRPC_PORT
//
// Optional (all have defaults):
//   - DT_ARANGO_ENDPOINT, DT_ARANGO_USER, DT_ARANGO_PASSWORD, DT_ARANGO_DATABASE
//   - CROSS_GRPC_ADDR, DT_GRPC_ADVERTISE_ADDR, CODEVALDDT_AGENCY_ID
//   - CROSS_PING_INTERVAL, CROSS_PING_TIMEOUT
func Load() Config {
	port := serverutil.MustGetEnv("CODEVALDDT_GRPC_PORT")
	return Config{
		GRPCPort:       port,
		ArangoEndpoint: serverutil.EnvOrDefault("DT_ARANGO_ENDPOINT", "http://localhost:8529"),
		ArangoUser:     serverutil.EnvOrDefault("DT_ARANGO_USER", "root"),
		ArangoPassword: serverutil.EnvOrDefault("DT_ARANGO_PASSWORD", ""),
		ArangoDatabase: serverutil.EnvOrDefault("DT_ARANGO_DATABASE", "codevalddt"),
		CrossGRPCAddr:  serverutil.EnvOrDefault("CROSS_GRPC_ADDR", ""),
		AdvertiseAddr:  serverutil.EnvOrDefault("DT_GRPC_ADVERTISE_ADDR", ":"+port),
		AgencyID:       serverutil.EnvOrDefault("CODEVALDDT_AGENCY_ID", ""),
		PingInterval:   serverutil.ParseDurationString("CROSS_PING_INTERVAL", 20*time.Second),
		PingTimeout:    serverutil.ParseDurationString("CROSS_PING_TIMEOUT", 5*time.Second),
	}
}
