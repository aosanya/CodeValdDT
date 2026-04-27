// Package config loads CodeValdDT runtime configuration from environment
// variables. All values have sensible defaults so the service starts in
// standalone mode (no Cross registration, default ArangoDB target) with zero
// environment variables set apart from CODEVALDDT_PORT.
package config

import (
	"time"

	"github.com/aosanya/CodeValdSharedLib/serverutil"
)

// Config holds all runtime configuration for the CodeValdDT service.
type Config struct {
	// GRPCPort is the port the gRPC server listens on. Required.
	GRPCPort string

	// ArangoEndpoint is the ArangoDB HTTP endpoint (default "http://localhost:8529").
	ArangoEndpoint string

	// ArangoUser is the ArangoDB username (default "root").
	ArangoUser string

	// ArangoPassword is the ArangoDB password.
	ArangoPassword string

	// ArangoDatabase is the ArangoDB database name (default "codevalddt").
	ArangoDatabase string

	// CrossGRPCAddr is the CodeValdCross gRPC address used for heartbeat
	// registration. Empty string disables registration (standalone mode).
	CrossGRPCAddr string

	// AdvertiseAddr is the address CodeValdCross dials back on.
	// Defaults to ":GRPCPort" when unset.
	AdvertiseAddr string

	// AgencyID is the agency this instance is scoped to, sent in every
	// Register heartbeat. Empty string means this instance serves all agencies.
	AgencyID string

	// PingInterval is the heartbeat cadence sent to CodeValdCross
	// (default 20s — the DT liveness contract).
	// Set to 0 to send only the initial registration ping.
	PingInterval time.Duration

	// PingTimeout is the per-RPC timeout for each Register call (default 5s).
	PingTimeout time.Duration
}

// Load reads runtime configuration from environment variables, falling back to
// sensible defaults for any variable that is unset or empty.
//
// Environment variables:
//
//	CODEVALDDT_PORT          gRPC listener port (required)
//	DT_ARANGO_ENDPOINT       ArangoDB endpoint (default "http://localhost:8529")
//	DT_ARANGO_USER           ArangoDB username (default "root")
//	DT_ARANGO_PASSWORD       ArangoDB password
//	DT_ARANGO_DATABASE       ArangoDB database name (default "codevalddt")
//	CROSS_GRPC_ADDR          CodeValdCross gRPC address (default ""; disables registration)
//	DT_GRPC_ADVERTISE_ADDR   address Cross dials back on (default ":PORT")
//	CODEVALDDT_AGENCY_ID     agency scope for this instance (default "")
//	CROSS_PING_INTERVAL      heartbeat cadence Go duration string (default "20s")
//	CROSS_PING_TIMEOUT       per-RPC Register timeout, Go duration string (default "5s")
func Load() Config {
	port := serverutil.MustGetEnv("CODEVALDDT_PORT")
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
