// Package arangodb implements the ArangoDB backend for CodeValdDT.
// All implementation logic lives in
// [github.com/aosanya/CodeValdSharedLib/entitygraph/arangodb]; this package
// is a thin service-scoped adapter that fixes the collection and graph names
// to their CodeValdDT-specific values.
//
// Entity collections:
//   - dt_entities  — fallback document collection for any TypeID without a
//     StorageCollection override
//   - dt_telemetry — telemetry records (TypeDefinitions whose StorageCollection
//     is set to "dt_telemetry")
//   - dt_events    — event records (TypeDefinitions whose StorageCollection
//     is set to "dt_events")
//
// Infrastructure collections:
//   - dt_relationships     — ArangoDB edge collection for all directed graph edges
//   - dt_schemas_draft     — one mutable draft schema document per agency
//   - dt_schemas_published — immutable append-only published schema snapshots
//
// Named graph: dt_graph
//
// Use [New] to obtain a (DataManager, SchemaManager) pair from an open database.
// Use [NewBackend] to connect and construct in a single call.
// Use [NewBackendFromDB] in tests that manage their own database lifecycle.
package arangodb

import (
	"fmt"

	driver "github.com/arangodb/go-driver"

	"github.com/aosanya/CodeValdSharedLib/entitygraph"
	sharedadb "github.com/aosanya/CodeValdSharedLib/entitygraph/arangodb"
	"github.com/aosanya/CodeValdSharedLib/types"
)

// Backend is a type alias for the shared ArangoDB Backend.
// Callers holding *Backend references continue to compile unchanged.
type Backend = sharedadb.Backend

// Config is the connection parameters for the CodeValdDT ArangoDB backend.
// It is an alias of [sharedadb.ConnConfig]; see that type for field docs.
// NewBackend requires Database to be set (e.g. "codevalddt").
type Config = sharedadb.ConnConfig

// toSharedConfig expands a CodeValdDT Config into a full SharedLib Config,
// filling in the fixed CodeValdDT-specific collection and graph names.
func toSharedConfig(cfg Config) sharedadb.Config {
	return sharedadb.Config{
		Endpoint:            cfg.Endpoint,
		Username:            cfg.Username,
		Password:            cfg.Password,
		Database:            cfg.Database,
		Schema:              cfg.Schema,
		EntityCollection:    "dt_entities",
		RelCollection:       "dt_relationships",
		SchemasDraftCol:     "dt_schemas_draft",
		SchemasPublishedCol: "dt_schemas_published",
		GraphName:           "dt_graph",
	}
}

// New constructs a Backend from an already-open driver.Database using the
// provided schema, ensures all collections and the named graph exist, and
// returns the Backend as both a DataManager and a SchemaManager.
func New(db driver.Database, schema types.Schema) (entitygraph.DataManager, entitygraph.SchemaManager, error) {
	if db == nil {
		return nil, nil, fmt.Errorf("arangodb: New: database must not be nil")
	}
	scfg := toSharedConfig(Config{Schema: schema})
	return sharedadb.New(db, scfg)
}

// NewBackend connects to ArangoDB using cfg, ensures all collections exist,
// and returns a ready-to-use Backend. cfg.Database is required.
func NewBackend(cfg Config) (*Backend, error) {
	if cfg.Database == "" {
		return nil, fmt.Errorf("arangodb: NewBackend: Database must be set (e.g. \"codevalddt\")")
	}
	scfg := toSharedConfig(cfg)
	return sharedadb.NewBackend(scfg)
}

// NewBackendFromDB constructs a Backend from an already-open driver.Database
// using the provided schema. Intended for tests that manage their own database
// lifecycle.
func NewBackendFromDB(db driver.Database, schema types.Schema) (*Backend, error) {
	if db == nil {
		return nil, fmt.Errorf("arangodb: NewBackendFromDB: database must not be nil")
	}
	scfg := toSharedConfig(Config{Schema: schema})
	return sharedadb.NewBackendFromDB(db, scfg)
}
