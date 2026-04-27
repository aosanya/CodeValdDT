// Command server starts the CodeValdDT gRPC microservice.
//
// MVP-DT-001 ships only the entry point and configuration loader. Storage
// (MVP-DT-002), the gRPC server (MVP-DT-003 + MVP-DT-004), and Cross
// registration (MVP-DT-005) are wired in by their respective tasks; until
// then, the binary loads its configuration, logs it, and exits cleanly.
//
// Configuration is via environment variables — see
// [config.Load] for the full list:
//
//	CODEVALDDT_GRPC_PORT       gRPC listener port (required, set in .env)
//	CROSS_GRPC_ADDR            CodeValdCross gRPC address for service
//	                           registration heartbeats and event publishing
//	                           (optional; omit to disable)
//	DT_GRPC_ADVERTISE_ADDR     address CodeValdCross dials back (default ":PORT")
//	CODEVALDDT_AGENCY_ID       agency ID sent in every Register heartbeat
//	                           (required when CROSS_GRPC_ADDR is set)
//	CROSS_PING_INTERVAL        heartbeat cadence (default "20s")
//	CROSS_PING_TIMEOUT         per-RPC timeout for each Register call (default "5s")
//
// ArangoDB backend (connects to a pre-existing database — CodeValdDT does
// not create it):
//
//	DT_ARANGO_ENDPOINT         ArangoDB endpoint URL (default "http://localhost:8529")
//	DT_ARANGO_USER             ArangoDB username (default "root")
//	DT_ARANGO_PASSWORD         ArangoDB password
//	DT_ARANGO_DATABASE         ArangoDB database name (default "codevalddt")
package main

import (
	"log"

	"github.com/aosanya/CodeValdDT/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("codevalddt: configuration loaded (GRPCPort=%s ArangoDB=%s/%s CrossAddr=%q AgencyID=%q)",
		cfg.GRPCPort, cfg.ArangoEndpoint, cfg.ArangoDatabase, cfg.CrossGRPCAddr, cfg.AgencyID)
	log.Println("codevalddt: scaffolding only — server wiring lands in MVP-DT-002 through MVP-DT-005")
}
