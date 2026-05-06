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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	codevalddt "github.com/aosanya/CodeValdDT"
	"github.com/aosanya/CodeValdDT/internal/config"
	"github.com/aosanya/CodeValdDT/internal/httphandler"
	"github.com/aosanya/CodeValdDT/internal/registrar"
	"github.com/aosanya/CodeValdDT/internal/server"
	dtarangodb "github.com/aosanya/CodeValdDT/storage/arangodb"
	"github.com/aosanya/CodeValdSharedLib/entitygraph"
	healthpb "github.com/aosanya/CodeValdSharedLib/gen/go/codevaldhealth/v1"
	entitygraphpb "github.com/aosanya/CodeValdSharedLib/gen/go/entitygraph/v1"
	"github.com/aosanya/CodeValdSharedLib/health"
)

// Run starts all CodeValdDT subsystems (Cross registrar, ArangoDB
// entitygraph backend, gRPC + HTTP via cmux) and blocks until SIGINT/SIGTERM
// triggers graceful shutdown.
func Run(cfg config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Signal handling ───────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		log.Println("codevalddt: shutdown signal received")
		cancel()
	}()

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

	// ── gRPC server (DT-008: TraverseGraph depth ceiling via interceptor) ─────
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(server.TraverseDepthInterceptor(server.MaxTraverseDepth)))
	reflection.Register(grpcServer)
	entitygraphpb.RegisterEntityServiceServer(grpcServer, server.NewEntityServer(backend))
	healthpb.RegisterHealthServiceServer(grpcServer, health.New("codevalddt"))

	// ── HTTP handler (DT-007: DTDL v3 export) ────────────────────────────────
	httpHandler := httphandler.New(backend)
	httpServer := &http.Server{
		Handler:      httpHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// ── TCP listener + cmux ───────────────────────────────────────────────────
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("listen on :%s: %w", cfg.GRPCPort, err)
	}

	mux := cmux.New(lis)
	grpcLis := mux.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
	)
	httpLis := mux.Match(cmux.Any())

	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil && err != grpc.ErrServerStopped {
			log.Printf("codevalddt: gRPC server error: %v", err)
		}
	}()
	go func() {
		if err := httpServer.Serve(httpLis); err != nil && err != http.ErrServerClosed {
			log.Printf("codevalddt: HTTP server error: %v", err)
		}
	}()
	go func() {
		if err := mux.Serve(); err != nil {
			log.Printf("codevalddt: cmux error: %v", err)
		}
	}()

	log.Printf("codevalddt: listening on :%s (gRPC + HTTP via cmux)", cfg.GRPCPort)

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	<-ctx.Done()
	log.Println("codevalddt: shutting down")

	grpcServer.GracefulStop()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("codevalddt: HTTP shutdown error: %v", err)
	}

	log.Println("codevalddt: stopped")
	return nil
}
