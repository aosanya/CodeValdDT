// Package app holds the shared runtime wiring for CodeValdDT. Both the
// production binary (cmd/server) and the local dev binary (cmd/dev) call
// Run; they differ only in which environment variables they set before
// loading config.
package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	codevalddt "github.com/aosanya/CodeValdDT"
	"github.com/aosanya/CodeValdDT/internal/config"
	"github.com/aosanya/CodeValdDT/internal/registrar"
	"github.com/aosanya/CodeValdDT/internal/server"
	dtarangodb "github.com/aosanya/CodeValdDT/storage/arangodb"
	"github.com/aosanya/CodeValdSharedLib/entitygraph"
	healthpb "github.com/aosanya/CodeValdSharedLib/gen/go/codevaldhealth/v1"
	entitygraphpb "github.com/aosanya/CodeValdSharedLib/gen/go/entitygraph/v1"
	"github.com/aosanya/CodeValdSharedLib/health"
	"github.com/aosanya/CodeValdSharedLib/serverutil"
)

// Run starts all CodeValdDT subsystems (Cross registrar, ArangoDB
// entitygraph backend, gRPC server) and blocks until SIGINT/SIGTERM triggers
// graceful shutdown.
func Run(cfg config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Cross registrar (optional) ───────────────────────────────────────────
	var pub codevalddt.CrossPublisher
	if cfg.CrossGRPCAddr != "" {
		reg, err := registrar.New(
			cfg.CrossGRPCAddr,
			cfg.AdvertiseAddr,
			cfg.AgencyID,
			cfg.PingInterval,
			cfg.PingTimeout,
		)
		if err != nil {
			log.Printf("codevalddt: registrar: failed to create: %v — continuing without registration", err)
		} else {
			defer reg.Close()
			go reg.Run(ctx)
			pub = reg
		}
	} else {
		log.Println("codevalddt: CROSS_GRPC_ADDR not set — skipping CodeValdCross registration")
	}
	_ = pub // CrossPublisher is wired for future event emission from DT subsystems.

	// ── ArangoDB entitygraph backend ─────────────────────────────────────────
	backend, err := dtarangodb.NewBackend(dtarangodb.Config{
		Endpoint: cfg.ArangoEndpoint,
		Username: cfg.ArangoUser,
		Password: cfg.ArangoPassword,
		Database: cfg.ArangoDatabase,
		Schema:   codevalddt.DefaultDTSchema(),
	})
	if err != nil {
		return fmt.Errorf("ArangoDB backend: %w", err)
	}

	// ── Schema seed (idempotent on startup) ──────────────────────────────────
	// DT's default schema is intentionally empty — seeding activates the
	// agency so subsequent agency-defined SetSchema/Publish calls succeed.
	if cfg.AgencyID != "" {
		seedCtx, seedCancel := context.WithTimeout(ctx, 10*time.Second)
		if err := entitygraph.SeedSchema(seedCtx, backend, cfg.AgencyID, codevalddt.DefaultDTSchema()); err != nil {
			log.Printf("codevalddt: schema seed: %v", err)
		}
		seedCancel()
	} else {
		log.Println("codevalddt: CODEVALDDT_AGENCY_ID not set — skipping schema seed")
	}

	// ── gRPC server ───────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("listen on :%s: %w", cfg.GRPCPort, err)
	}

	grpcServer, _ := serverutil.NewGRPCServer()
	// CodeValdDT exposes only the shared EntityService — agencies declare
	// their own TypeDefinitions at runtime. There is no DT-specific gRPC
	// service for now.
	entitygraphpb.RegisterEntityServiceServer(grpcServer, server.NewEntityServer(backend))
	healthpb.RegisterHealthServiceServer(grpcServer, health.New("codevalddt"))

	// ── Signal handling ───────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		log.Println("codevalddt: shutdown signal received")
		cancel()
	}()

	log.Printf("codevalddt: gRPC server listening on :%s", cfg.GRPCPort)
	serverutil.RunWithGracefulShutdown(ctx, grpcServer, lis, 30*time.Second)
	return nil
}
