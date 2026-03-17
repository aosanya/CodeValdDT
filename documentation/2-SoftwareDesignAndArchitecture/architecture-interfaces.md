# CodeValdDT — Architecture: Interfaces & Models

> Part of [architecture.md](architecture.md)

## 1. Core Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Business-logic entry point | `DTManager` interface | gRPC handlers delegate to it; no logic in handlers |
| Downstream communication | gRPC only — no direct Go imports | Stable versioned contracts; independent deployment |
| Storage injection | `Backend` interface injected by `cmd/main.go` | Backend-agnostic core; easy to mock for tests |
| Graph storage | ArangoDB edge collection (`relationships`) | Native AQL graph traversal; no separate graph engine needed |
| Database isolation | One ArangoDB database per agency | Consistent with CodeVald platform convention |
| Schema definition | `DTSchema` per agency stored in `dt_schemas`; published via DT gRPC API by the Agency Owner | Versioned and immutable — each publish creates a new version |
| Schema enforcement | None in v1 | Keeps DT lean; enforcement can be added in v2 without API break |
| Pub/sub events | CodeValdCross topic-based pub/sub | Platform standard; agencyID-scoped topics |
| Telemetry model | Push + pub/sub | Caller pushes a reading via gRPC; DT stores it and publishes an event |
| Error types | `errors.go` at module root | All exported errors in one place |
| Value types | `models.go` at module root | Pure data structs, no methods |

---

## 2. DTManager Interface

```go
// DTManager is the sole business-logic entry point for all digital-twin operations.
// gRPC handlers hold this interface — never the concrete type.
// One instance per process; keyed internally by agencyID on every call.
type DTManager interface {
    // Entity operations
    CreateEntity(ctx context.Context, req CreateEntityRequest) (Entity, error)
    GetEntity(ctx context.Context, agencyID, entityID string) (Entity, error)
    UpdateEntity(ctx context.Context, agencyID, entityID string, req UpdateEntityRequest) (Entity, error)
    DeleteEntity(ctx context.Context, agencyID, entityID string) error
    ListEntities(ctx context.Context, filter EntityFilter) ([]Entity, error)

    // Graph operations
    CreateRelationship(ctx context.Context, req CreateRelationshipRequest) (Relationship, error)
    DeleteRelationship(ctx context.Context, agencyID, relationshipID string) error
    TraverseGraph(ctx context.Context, req TraverseGraphRequest) ([]Entity, error)

    // Telemetry operations
    RecordTelemetry(ctx context.Context, req RecordTelemetryRequest) (TelemetryReading, error)
    QueryTelemetry(ctx context.Context, filter TelemetryFilter) ([]TelemetryReading, error)

    // Event operations
    RecordEvent(ctx context.Context, req RecordEventRequest) (Event, error)
    ListEvents(ctx context.Context, filter EventFilter) ([]Event, error)

    // Schema management
    PublishSchema(ctx context.Context, agencyID string, types []types.TypeDefinition) (types.Schema, error)
    GetSchema(ctx context.Context, agencyID string, version int) (types.Schema, error)
    ListSchemaVersions(ctx context.Context, agencyID string) ([]types.Schema, error)
}
```

```go
// Backend is the storage contract injected into DTManager.
// cmd/main.go constructs the chosen implementation (e.g. arangodb.NewBackend).
// Each method receives agencyID so the implementation can route to the correct
// per-agency ArangoDB database.
type Backend interface {
    // Entity operations
    InsertEntity(ctx context.Context, req CreateEntityRequest) (Entity, error)
    GetEntity(ctx context.Context, agencyID, entityID string) (Entity, error)
    UpdateEntity(ctx context.Context, agencyID, entityID string, req UpdateEntityRequest) (Entity, error)
    DeleteEntity(ctx context.Context, agencyID, entityID string) error
    ListEntities(ctx context.Context, filter EntityFilter) ([]Entity, error)

    // Graph operations — relationships MUST be stored in an edge collection
    InsertRelationship(ctx context.Context, req CreateRelationshipRequest) (Relationship, error)
    DeleteRelationship(ctx context.Context, agencyID, relationshipID string) error
    TraverseGraph(ctx context.Context, req TraverseGraphRequest) ([]Entity, error)

    // Telemetry operations
    InsertTelemetry(ctx context.Context, req RecordTelemetryRequest) (TelemetryReading, error)
    QueryTelemetry(ctx context.Context, filter TelemetryFilter) ([]TelemetryReading, error)

    // Event operations
    InsertEvent(ctx context.Context, req RecordEventRequest) (Event, error)
    ListEvents(ctx context.Context, filter EventFilter) ([]Event, error)

    // Schema operations
    InsertSchema(ctx context.Context, schema types.Schema) (types.Schema, error)
    GetSchema(ctx context.Context, agencyID string, version int) (types.Schema, error)
    ListSchemaVersions(ctx context.Context, agencyID string) ([]types.Schema, error)
    NextSchemaVersion(ctx context.Context, agencyID string) (int, error)
}
```

---

## 3. Data Models

### DTSchema

`DTSchema` uses `types.Schema` from `CodeValdSharedLib`. The full type system—
`PropertyType`, `PropertyDefinition`, `RatingConfig`, `TypeDefinition`, `Schema`—
is defined in `CodeValdSharedLib/types/schema.go` and shared with `CodeValdComm`.

```go
// DTSchema is the CodeValdDT alias for types.Schema.
// Stored in the dt_schemas collection — one immutable document per agency per version.
// Agency Owners call PublishSchema to create a new version.
type DTSchema = types.Schema
```

---

### Entity and Runtime Types

```go
// Entity is an instance of a typed real-world object in a Digital Twin.
// TypeID matches TypeDefinition.Name in the agency's current DTSchema.
// Properties hold the current state values; no schema validation in v1.
type Entity struct {
    ID         string
    AgencyID   string
    TypeID     string            // matches TypeDefinition.Name in the agency's current DTSchema
    Properties map[string]any    // current state values
    CreatedAt  time.Time
    UpdatedAt  time.Time
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

// TelemetryReading is a single time-stamped sensor or metric value for an entity.
type TelemetryReading struct {
    ID        string
    AgencyID  string
    EntityID  string
    Name      string            // metric name (e.g. "temperature", "pressure")
    Value     any               // numeric, bool, or string value
    Timestamp time.Time
}

// RecordTelemetryRequest is the input for recording a telemetry value.
type RecordTelemetryRequest struct {
    AgencyID  string
    EntityID  string
    Name      string
    Value     any
    Timestamp time.Time         // caller provides; allows backfill
}

// TelemetryFilter scopes a historical telemetry query.
type TelemetryFilter struct {
    AgencyID  string
    EntityID  string
    Name      string            // empty = all metrics for entity
    Since     time.Time
    Until     time.Time
    Limit     int               // 0 = no limit
}

// Event is an occurrence that changed the state of an entity (discrete, not time-series).
type Event struct {
    ID        string
    AgencyID  string
    EntityID  string
    Name      string            // event type label (e.g. "pressure_exceeded", "valve_opened")
    Payload   map[string]any
    Timestamp time.Time
}

// RecordEventRequest is the input for appending a new event to an entity's log.
type RecordEventRequest struct {
    AgencyID  string
    EntityID  string
    Name      string
    Payload   map[string]any
    Timestamp time.Time
}

// EventFilter scopes an event log read.
type EventFilter struct {
    AgencyID string
    EntityID string
    Name     string    // empty = all event types
    Since    time.Time
    Until    time.Time
    Limit    int
}
```
