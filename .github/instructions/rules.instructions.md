---
applyTo: '**'
---

# CodeValdDT — Code Structure Rules

## Service Design Principles

CodeValdDT is a **Go gRPC microservice** — not a library and not a monolith.

- **Has a `cmd/main.go` binary entry point** — wires all dependencies and
  starts the server
- **No business logic in `cmd/`** — `main.go` only constructs dependencies
  and calls `server.Run`
- **Callers inject dependencies** — `Backend` and any cross-service clients
  are never hardcoded
- **Exported API surface is minimal** — expose only what other packages within
  this module need
- **No AI agent logic, LLM integration, or frontend code** — this service
  owns digital twin data only

---

## Interface-First Design

**Always define interfaces before concrete types.**

```go
// ✅ CORRECT — interface at root package level; concrete impl is unexported in internal/manager/
type DTManager interface {
    CreateEntity(ctx context.Context, req CreateEntityRequest) (Entity, error)
    GetEntity(ctx context.Context, agencyID, entityID string) (Entity, error)
    UpdateEntity(ctx context.Context, agencyID, entityID string, req UpdateEntityRequest) (Entity, error)
    DeleteEntity(ctx context.Context, agencyID, entityID string) error
    ListEntities(ctx context.Context, filter EntityFilter) ([]Entity, error)
    CreateRelationship(ctx context.Context, req CreateRelationshipRequest) (Relationship, error)
    DeleteRelationship(ctx context.Context, agencyID, relationshipID string) error
    TraverseGraph(ctx context.Context, req TraverseGraphRequest) ([]Entity, error)
    RecordTelemetry(ctx context.Context, req RecordTelemetryRequest) (TelemetryReading, error)
    QueryTelemetry(ctx context.Context, filter TelemetryFilter) ([]TelemetryReading, error)
    RecordEvent(ctx context.Context, req RecordEventRequest) (Event, error)
    ListEvents(ctx context.Context, filter EventFilter) ([]Event, error)
}

// ❌ WRONG — leaking a concrete storage struct to callers
type ArangoDTManager struct {
    db driver.Database
}
```

**File layout — one primary concern per file:**

```
errors.go                            → ErrEntityNotFound, ErrRelationshipNotFound, etc.
models.go                            → Entity, Relationship, TelemetryReading, Event, filters
codevalddt.go                        → DTManager + Backend interfaces
internal/manager/manager.go          → Concrete DTManager implementation
internal/server/server.go            → Inbound gRPC server (DTService handlers)
internal/config/config.go            → Configuration struct + loader
internal/registrar/registrar.go      → Cross registration heartbeat
storage/arangodb/storage.go          → ArangoDB Backend implementation
cmd/main.go                          → Dependency wiring only
```

---

## Graph Edge Rules

**Relationships are ArangoDB edge collection documents — never regular documents.**

```go
// ✅ CORRECT — insert into the edge collection
func (b *backend) InsertRelationship(ctx context.Context, req CreateRelationshipRequest) (Relationship, error) {
    edgeDoc := map[string]any{
        "_from": "entities/" + req.FromEntityID,
        "_to":   "entities/" + req.ToEntityID,
        "name":  req.Name,
    }
    _, err := b.db.Collection(ctx, "relationships").CreateDocument(ctx, edgeDoc)
    ...
}

// ❌ WRONG — storing a relationship as a regular document
func (b *backend) InsertRelationship(...) {
    b.db.Collection(ctx, "entities").CreateDocument(ctx, rel) // wrong collection type
}
```

Use ArangoDB AQL graph traversal for `TraverseGraph`:

```aql
FOR v, e, p IN 1..{depth} {direction} 'entities/{startID}'
  GRAPH 'dt_graph'
  RETURN v
```

---

## Pub/Sub Event Rules

**Two events must be published in v1 — never skip them.**

```go
// ✅ CORRECT — publish after CreateEntity
func (m *manager) CreateEntity(ctx context.Context, req CreateEntityRequest) (Entity, error) {
    entity, err := m.backend.InsertEntity(ctx, req)
    if err != nil {
        return Entity{}, err
    }
    topic := fmt.Sprintf("cross.dt.%s.entity.created", req.AgencyID)
    m.crossClient.Publish(ctx, topic, entity.ID)
    return entity, nil
}

// ✅ CORRECT — publish after RecordTelemetry
func (m *manager) RecordTelemetry(ctx context.Context, req RecordTelemetryRequest) (TelemetryReading, error) {
    reading, err := m.backend.InsertTelemetry(ctx, req)
    if err != nil {
        return TelemetryReading{}, err
    }
    topic := fmt.Sprintf("cross.dt.%s.telemetry.recorded", req.AgencyID)
    m.crossClient.Publish(ctx, topic, reading.ID)
    return reading, nil
}

// ❌ WRONG — silent return without publishing
func (m *manager) CreateEntity(...) (Entity, error) {
    return m.backend.InsertEntity(ctx, req)
}
```

---

## gRPC Handler Rules

**Handlers are thin — delegate immediately to `DTManager`.**

```go
// ✅ CORRECT — handler delegates to interface
func (s *server) CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (*pb.Entity, error) {
    entity, err := s.manager.CreateEntity(ctx, toModel(req))
    if err != nil {
        return nil, toGRPCError(err)
    }
    return toProto(entity), nil
}

// ❌ WRONG — business logic inside handler
func (s *server) CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (*pb.Entity, error) {
    doc, err := s.db.Collection("entities").CreateDocument(ctx, req)
    ...
}
```

---

## Storage Backend Rules

The `Backend` interface is the injection point. `cmd/main.go` constructs the
desired `Backend` and passes it to `NewDTManager`.

```go
// ✅ CORRECT — Backend injected by cmd/main.go
b, _ := arangodb.NewBackend(cfg.ArangoDB)
mgr := manager.NewDTManager(b, crossClient)

// ❌ WRONG — hardcoded driver inside manager
func NewDTManager() DTManager {
    db, _ := arangodb.NewDatabase(...)
    return &dtManager{db: db}
}
```

---

## CodeValdCross Registration Rules

**Registration must happen on startup and repeat as a liveness heartbeat.**

```go
// ✅ CORRECT — register on startup with heartbeat loop
func register(ctx context.Context, crossAddr string) {
    req := &pb.RegisterRequest{
        ServiceName: "codevalddt",
        Addr:        ":50055",
        Produces: []string{
            "cross.dt.{agencyID}.entity.created",
            "cross.dt.{agencyID}.telemetry.recorded",
        },
        Consumes: []string{},
        Routes:   dtRoutes(),
    }
    for {
        if err := crossClient.Register(ctx, req); err != nil {
            log.Printf("codevalddt: register error: %v", err)
        }
        select {
        case <-ctx.Done():
            return
        case <-time.After(20 * time.Second):
        }
    }
}

// ❌ WRONG — register once and forget
func main() {
    crossClient.Register(ctx, req)
    server.Run(ctx)
}
```

---

## Error Types

All exported errors live in `errors.go`.

```go
var (
    ErrEntityNotFound       = errors.New("entity not found")
    ErrRelationshipNotFound = errors.New("relationship not found")
    ErrEntityAlreadyExists  = errors.New("entity already exists")
)
```

Map errors to gRPC status codes in the server layer only:

```go
func toGRPCError(err error) error {
    switch {
    case errors.Is(err, ErrEntityNotFound):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, ErrEntityAlreadyExists):
        return status.Error(codes.AlreadyExists, err.Error())
    default:
        return status.Error(codes.Internal, err.Error())
    }
}
```

---

## Naming Conventions

| Category | Convention | Example |
|---|---|---|
| Branch | `feature/DT-XXX_description` | `feature/DT-001_create-entity` |
| Commit | `DT-XXX: message` | `DT-001: Add CreateEntity gRPC handler` |
| Package | lowercase, no abbreviations | `codevalddt`, `manager`, `server` |
| Interfaces | noun-only | `DTManager`, `Backend` |
| Exported types | PascalCase | `Entity`, `Relationship`, `TelemetryReading` |
| gRPC service | `DTService` | in `proto/codevalddt/dt.proto` |
| Topic constants | package-level const | `TopicEntityCreated = "cross.dt.%s.entity.created"` |

---

## Anti-Patterns

- ❌ **AI/LLM calls** — not in this service
- ❌ **Frontend routes or HTML templates** — CodeValdHi only
- ❌ **Work item or task management** — CodeValdWork only
- ❌ **Git operations** — CodeValdGit only
- ❌ **Pub/sub topic strings as raw literals** — define as constants
- ❌ **Business logic in gRPC handlers** — delegate to `DTManager`
- ❌ **Hardcoded ArangoDB connection in manager** — inject `Backend`
- ❌ **Relationships in a document collection** — use the edge collection
- ❌ **Skipping pub/sub events** — always publish on entity creation and
  telemetry recording
