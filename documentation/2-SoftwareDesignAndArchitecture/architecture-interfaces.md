# CodeValdDT — Architecture: Interfaces & Models

> Part of [architecture.md](architecture.md)

## 1. Core Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Business-logic entry point | `DTDataManager = entitygraph.DataManager` (SharedLib) | gRPC handlers delegate to it; owns entity lifecycle and graph operations; shared with CodeValdComm |
| Downstream communication | gRPC only — no direct Go imports | Stable versioned contracts; independent deployment |
| Storage injection | `DTSchemaManager = entitygraph.SchemaManager` injected by `cmd/main.go` | Owns schema reads and writes (`SetSchema`, `GetSchema`, `ListSchemaVersions`) on `dt_schemas`; backend-agnostic |
| Graph storage | ArangoDB edge collection (`dt_relationships`) | Native AQL graph traversal; no separate graph engine needed |
| Database isolation | Single shared ArangoDB database (`DT_ARANGO_DATABASE` env var) | All collections scoped by `agencyID` field; consistent with CodeVald env-var convention |
| Soft delete | `DeleteEntity` sets `deleted: true` and `deletedAt` — no cascade | Preserves telemetry and event history; relationships retained as orphans; hard delete deferred to v2 |
| Traversal result | `TraverseGraph` returns both vertices and edges as `TraverseGraphResult` | Callers need edge properties (name, metadata) to render the graph without a second round-trip |
| Relationship reads | `GetRelationship` + `ListRelationships` on `DTDataManager` | Required for UI entity-connection views |
| Cross publisher | `DTDataManager` uses SharedLib `gen/go/codevaldcross/v1` stubs to call `Publish` — no extra wrapper | Use SharedLib infrastructure; no new abstraction needed for v1 |
| Schema | `DTSchemaManager` owns `SetSchema`, `GetSchema`, and `ListSchemaVersions` on `dt_schemas`; `DTDataManager` has no schema methods | Schema storage is a `DTSchemaManager` concern; business-logic layer stays clean |
| Schema enforcement | `TypeDefinition.Immutable` — `UpdateEntity` returns `ErrImmutableType` when the resolved type has `Immutable: true`; storage routing via `TypeDefinition.StorageCollection` | Immutability and storage are schema-driven; no hard-coded type checks in the manager |
| Pub/sub events | CodeValdCross topic-based pub/sub | Platform standard; agencyID-scoped topics |
| Error types | `errors.go` at module root | All exported errors in one place |
| Value types | `models.go` at module root | Pure data structs, no methods |

---

## 2. DTDataManager & DTSchemaManager Interfaces

Both interfaces are defined in **`CodeValdSharedLib/entitygraph`** and aliased
locally. CodeValdComm uses the same interfaces for its own entity-graph store.

```go
import "github.com/aosanya/CodeValdSharedLib/entitygraph"

// DTDataManager is the CodeValdDT alias for entitygraph.DataManager.
// gRPC handlers hold this interface — never the concrete type.
// Telemetry and events are schema-defined entity types routed to their
// respective collections via TypeDefinition.StorageCollection.
type DTDataManager = entitygraph.DataManager

// DTSchemaManager is the CodeValdDT alias for entitygraph.SchemaManager.
// cmd/main.go constructs the concrete implementation (e.g. arangodb.NewDTSchemaManager)
// and injects it into the concrete DTDataManager.
type DTSchemaManager = entitygraph.SchemaManager
```

Full interface and model definitions live in SharedLib:

```go
// entitygraph.DataManager (defined in CodeValdSharedLib/entitygraph/entitygraph.go)
type DataManager interface {
    // Entity operations
    CreateEntity(ctx context.Context, req CreateEntityRequest) (Entity, error)
    GetEntity(ctx context.Context, agencyID, entityID string) (Entity, error)
    UpdateEntity(ctx context.Context, agencyID, entityID string, req UpdateEntityRequest) (Entity, error)
    DeleteEntity(ctx context.Context, agencyID, entityID string) error
    ListEntities(ctx context.Context, filter EntityFilter) ([]Entity, error)

    // Graph operations
    CreateRelationship(ctx context.Context, req CreateRelationshipRequest) (Relationship, error)
    GetRelationship(ctx context.Context, agencyID, relationshipID string) (Relationship, error)
    DeleteRelationship(ctx context.Context, agencyID, relationshipID string) error
    ListRelationships(ctx context.Context, filter RelationshipFilter) ([]Relationship, error)
    TraverseGraph(ctx context.Context, req TraverseGraphRequest) (TraverseGraphResult, error)
}

// entitygraph.SchemaManager (defined in CodeValdSharedLib/entitygraph/entitygraph.go)
type SchemaManager interface {
    SetSchema(ctx context.Context, schema types.Schema) error
    GetSchema(ctx context.Context, agencyID string, version int) (types.Schema, error)
    ListSchemaVersions(ctx context.Context, agencyID string) ([]types.Schema, error)
}
```

---

## 3. Data Models

`Entity`, `Relationship`, and all associated request/filter/result types are
defined in `CodeValdSharedLib/entitygraph/entitygraph.go` and imported directly.
The models below are reproduced here for reference only.

### DTSchema

`DTSchema` uses `types.Schema` from `CodeValdSharedLib`. The full type system—
`PropertyType`, `PropertyDefinition`, `RatingConfig`, `TypeDefinition`, `Schema`—
is defined in `CodeValdSharedLib/types/schema.go` and shared with `CodeValdComm`.

```go
// DTSchema is the CodeValdDT alias for types.Schema.
// Stored in the dt_schemas collection — one immutable document per agency per version.
// Written via DTSchemaManager.SetSchema; read via GetSchema and ListSchemaVersions.
// DTDataManager has no schema methods.
type DTSchema = types.Schema
```

`TypeDefinition` carries two DT-relevant fields beyond the base property list:

```go
// TypeDefinition (excerpt — full definition in CodeValdSharedLib/types/schema.go)
type TypeDefinition struct {
    Name              string
    DisplayName       string
    Properties        []PropertyDefinition

    // StorageCollection is the backing ArangoDB collection for instances of this
    // type. Empty means the service default ("dt_entities").
    // Set to "dt_telemetry" or "dt_events" to route writes to a specialised collection.
    StorageCollection string

    // Immutable indicates that instances of this type cannot be updated after
    // creation. UpdateEntity returns ErrImmutableType for immutable types.
    // Only CreateEntity and DeleteEntity are valid.
    Immutable bool
}
```

---

### Entity and Runtime Types

```go
// Entity is an instance of a typed real-world object in a Digital Twin.
// TypeID matches TypeDefinition.Name in the agency's current DTSchema.
// Properties hold the current state values; no schema validation in v1.
// Deleted and DeletedAt are set by DeleteEntity (soft delete) — the entity
// is never hard-deleted in v1.
type Entity struct {
    ID         string
    AgencyID   string
    TypeID     string            // matches TypeDefinition.Name in the agency's current DTSchema
    Properties map[string]any    // current state values
    CreatedAt  time.Time
    UpdatedAt  time.Time
    Deleted    bool              // true once DeleteEntity has been called
    DeletedAt  *time.Time        // nil until deleted
}

// CreateEntityRequest is the input for creating a new entity.
type CreateEntityRequest struct {
    AgencyID   string
    TypeID     string
    Properties map[string]any
}

// UpdateEntityRequest is the input for patching entity properties.
type UpdateEntityRequest struct {
    Properties map[string]any
}

// EntityFilter scopes a list operation. Zero values mean "no filter".
type EntityFilter struct {
    AgencyID string
    TypeID   string
}

// Relationship is a directed graph edge between two entities.
// Stored in an ArangoDB edge collection — _from and _to reference entities/ documents.
type Relationship struct {
    ID         string
    AgencyID   string
    Name       string            // semantic label (e.g. "connects_to", "reports_to")
    FromID     string            // source entity ID
    ToID       string            // target entity ID
    Properties map[string]any
    CreatedAt  time.Time
}

// CreateRelationshipRequest is the input for creating a graph edge.
type CreateRelationshipRequest struct {
    AgencyID   string
    Name       string
    FromID     string
    ToID       string
    Properties map[string]any
}

// TraverseGraphRequest walks the entity graph from a starting entity.
type TraverseGraphRequest struct {
    AgencyID  string
    StartID   string
    Direction string // "outbound" | "inbound" | "any"
    Depth     int    // max traversal depth; 0 means 1
}

// TraverseGraphResult is returned by TraverseGraph.
// Both visited vertices and traversed edges are included so callers can
// inspect relationship names and properties without a second round-trip.
// Soft-deleted entities are excluded from Vertices.
type TraverseGraphResult struct {
    Vertices []Entity       // reachable entities (excludes soft-deleted)
    Edges    []Relationship // traversed edges in order of discovery
}

// RelationshipFilter scopes a ListRelationships query.
// Zero-value fields are ignored (no filter applied for that field).
type RelationshipFilter struct {
    AgencyID string
    FromID   string // filter by source entity ID; empty = any source
    ToID     string // filter by target entity ID; empty = any target
    Name     string // filter by relationship type label; empty = all types
}
```
