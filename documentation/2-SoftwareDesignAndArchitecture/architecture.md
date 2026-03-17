# CodeValdDT — Architecture

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

---

## 4. Package Structure

```
CodeValdDT/
├── cmd/
│   └── main.go                   # Wires dependencies only — no business logic
├── go.mod
├── errors.go                     # ErrEntityNotFound, ErrRelationshipNotFound, etc.
├── models.go                     # Entity, Relationship, TelemetryReading, Event, filter/request types
├── codevalddt.go                 # DTManager + Backend interfaces
├── internal/
│   ├── config/
│   │   └── config.go             # Config struct + loader (env / YAML)
│   ├── manager/
│   │   └── manager.go            # Concrete DTManager — holds Backend + CrossClient
│   ├── server/
│   │   └── server.go             # Inbound gRPC server — DTService handlers
│   └── registrar/
│       └── registrar.go          # Cross registration heartbeat loop
├── storage/
│   └── arangodb/
│       └── storage.go            # ArangoDB Backend implementation
├── proto/
│   └── codevalddt/
│       └── dt.proto              # DTService gRPC definition
├── gen/
│   └── go/                       # Generated protobuf code (buf generate — do not hand-edit)
└── bin/
    └── codevalddt                # Compiled binary
```

---

## 5. gRPC Service Definition

```protobuf
syntax = "proto3";
package codevalddt.v1;

service DTService {
    // Entity management
    rpc CreateEntity         (CreateEntityRequest)         returns (Entity);
    rpc GetEntity            (GetEntityRequest)            returns (Entity);
    rpc UpdateEntity         (UpdateEntityRequest)         returns (Entity);
    rpc DeleteEntity         (DeleteEntityRequest)         returns (google.protobuf.Empty);
    rpc ListEntities         (ListEntitiesRequest)         returns (ListEntitiesResponse);

    // Graph operations
    rpc CreateRelationship   (CreateRelationshipRequest)   returns (Relationship);
    rpc DeleteRelationship   (DeleteRelationshipRequest)   returns (google.protobuf.Empty);
    rpc TraverseGraph        (TraverseGraphRequest)        returns (TraverseGraphResponse);

    // Telemetry
    rpc RecordTelemetry      (RecordTelemetryRequest)      returns (TelemetryReading);
    rpc QueryTelemetry       (QueryTelemetryRequest)       returns (QueryTelemetryResponse);

    // Events
    rpc RecordEvent          (RecordEventRequest)          returns (Event);
    rpc ListEvents           (ListEventsRequest)           returns (ListEventsResponse);

    // Schema management
    rpc PublishSchema        (PublishSchemaRequest)        returns (DTSchema);
    rpc GetSchema            (GetSchemaRequest)            returns (DTSchema);
    rpc ListSchemaVersions   (ListSchemaVersionsRequest)   returns (ListSchemaVersionsResponse);
}
```

Generated Go code lives in `gen/go/`. **Never hand-edit generated files.**

---

## 6. CodeValdCross Registration

On startup, `cmd/main.go` starts a registration heartbeat. The loop calls
`OrchestratorService.Register` on CodeValdCross every **20 seconds**.

```go
RegisterRequest{
    ServiceName: "codevalddt",
    Addr:        ":50055",
    Produces: []string{
        "cross.dt.{agencyID}.entity.created",
        "cross.dt.{agencyID}.telemetry.recorded",
    },
    Consumes: []string{},
    Routes: []Route{
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "PUT",    Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "DELETE", Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/relationships"},
        {Method: "DELETE", Pattern: "/{agencyId}/dt/relationships/{relationshipId}"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities/{entityId}/traverse"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities/{entityId}/telemetry"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities/{entityId}/telemetry"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities/{entityId}/events"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities/{entityId}/events"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/schema"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/schema"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/schema/versions"},
    },
}
```

The repeat call is the **liveness signal** — Cross expires services that stop
registering. If Cross is not yet up, the loop retries silently.

---

## 7. ArangoDB Schema

### Per-Agency Database

Each agency gets its own ArangoDB database. The database name is derived from
the agency ID: `dt_{agencyID}`.

| Collection | Collection Type | Contents |
|---|---|---|
| `dt_schemas` | document | Versioned, immutable schema definitions (`TypeDefinition` list) |
| `entities` | document | Entity instances (typed objects) |
| `relationships` | **edge** | Graph edges between entities — `_from` + `_to` ref `entities/{id}` |
| `telemetry` | document | Time-series telemetry readings |
| `events` | document | Event log entries |

> ⚠️ `relationships` **MUST** be created as an edge collection
> (`CollectionTypeEdge`). Creating it as a document collection prevents AQL
> graph traversal and breaks `TraverseGraph`. This is a one-time constraint —
> collection type cannot be changed after creation.

### Named Graph

A named graph `dt_graph` is created in each agency database:

```
Graph: dt_graph
Edge collection:   relationships
Vertex collection: entities
```

The named graph enables AQL traversal:
`FOR v, e, p IN 1..@depth @direction @startVertex GRAPH 'dt_graph'`

### Document Shapes

**`entities/{id}`**
```json
{
  "_key":       "entity-uuid",
  "agencyID":   "agency-123",
  "typeID":     "Pump",
  "properties": { "pressure": 4.2, "status": "running" },
  "createdAt":  "2026-01-01T00:00:00Z",
  "updatedAt":  "2026-01-01T00:00:00Z"
}
```

**`relationships/{id}`** (edge document)
```json
{
  "_key":       "rel-uuid",
  "_from":      "entities/entity-uuid-1",
  "_to":        "entities/entity-uuid-2",
  "agencyID":   "agency-123",
  "name":       "connects_to",
  "properties": {},
  "createdAt":  "2026-01-01T00:00:00Z"
}
```

**`telemetry/{id}`**
```json
{
  "_key":      "tel-uuid",
  "entityID":  "entity-uuid",
  "agencyID":  "agency-123",
  "name":      "temperature",
  "value":     42.5,
  "timestamp": "2026-01-01T00:00:00Z"
}
```

**`events/{id}`**
```json
{
  "_key":      "evt-uuid",
  "entityID":  "entity-uuid",
  "agencyID":  "agency-123",
  "name":      "valve_opened",
  "payload":   { "operator": "agent-7" },
  "timestamp": "2026-01-01T00:00:00Z"
}
```

### Indexes

| Collection | Field(s) | Type | Reason |
|---|---|---|---|
| `dt_schemas` | `agencyID` | persistent | All schema versions for an agency |
| `dt_schemas` | `agencyID, version` | unique persistent | One document per agency per version |
| `entities` | `agencyID` | persistent | All entity queries scope by agency |
| `entities` | `typeID` | persistent | ListEntities by type |
| `relationships` | `agencyID` | persistent | Scope edge queries |
| `telemetry` | `entityID, timestamp` | persistent | Time-range queries per entity |
| `events` | `entityID, timestamp` | persistent | Chronological event log per entity |

---

## 8. Error Types

Defined in `errors.go`:

```go
var (
    ErrEntityNotFound       = errors.New("entity not found")
    ErrRelationshipNotFound = errors.New("relationship not found")
    ErrInvalidEntity        = errors.New("invalid entity: missing required fields")
    ErrInvalidRelationship  = errors.New("invalid relationship: missing required fields")
    ErrTelemetryNotFound    = errors.New("telemetry not found")
    ErrEventNotFound        = errors.New("event not found")
    ErrSchemaNotFound       = errors.New("schema not found")
    ErrInvalidSchema        = errors.New("invalid schema: missing required fields")
)
```

Map to gRPC status codes in `internal/server/server.go`:

| Error | gRPC code |
|---|---|
| `ErrEntityNotFound` | `codes.NotFound` |
| `ErrRelationshipNotFound` | `codes.NotFound` |
| `ErrTelemetryNotFound` | `codes.NotFound` |
| `ErrEventNotFound` | `codes.NotFound` |
| `ErrInvalidEntity` | `codes.InvalidArgument` |
| `ErrInvalidRelationship` | `codes.InvalidArgument` |
| `ErrSchemaNotFound` | `codes.NotFound` |
| `ErrInvalidSchema` | `codes.InvalidArgument` |
| all others | `codes.Internal` |

---

## 9. CreateEntity Flow (Critical Path)

```
gRPC handler
    │
    ▼
DTManager.CreateEntity(ctx, req)
    │
    ├── validate: req.AgencyID, req.TypeID non-empty
    │       → ErrInvalidEntity if violated
    │
    ├── backend.InsertEntity(ctx, req)     → entities collection in dt_{agencyID} database
    │       returns Entity{ID, AgencyID, TypeID, Properties, CreatedAt, UpdatedAt}
    │
    └── crossPublisher.Publish(ctx,
            "cross.dt.{agencyID}.entity.created",
            entity.ID)
            │
            ▼
        Cross routes event to subscribers
```

`cross.dt.{agencyID}.entity.created` **MUST be published** after every
successful create. Publish failures are logged but not returned to the caller —
the entity is already persisted.

---

## 10. RecordTelemetry Flow (Critical Path)

```
gRPC handler
    │
    ▼
DTManager.RecordTelemetry(ctx, req)
    │
    ├── validate: req.AgencyID, req.EntityID, req.Name non-empty
    │
    ├── backend.InsertTelemetry(ctx, req)  → telemetry collection in dt_{agencyID} database
    │       returns TelemetryReading
    │
    └── crossPublisher.Publish(ctx,
            "cross.dt.{agencyID}.telemetry.recorded",
            telemetryReading.ID)
            │
            ▼
        Cross routes event to subscribers
```

`cross.dt.{agencyID}.telemetry.recorded` **MUST be published** after every
successful record. Same best-effort rule applies.

---

## 11. SharedLib Dependency

CodeValdDT imports `github.com/aosanya/CodeValdSharedLib` for:

| SharedLib package | Provides |
|---|---|
| `types/schema` | `Schema`, `TypeDefinition`, `PropertyDefinition`, `PropertyType`, `RatingConfig` |
| `registrar` | Cross registration heartbeat loop |
| `serverutil` | `envOrDefault`, `parseDuration` helpers and gRPC server setup |
| `arangoutil` | ArangoDB driver connection, auth, database bootstrap |
| `gen/go/codevaldcross/v1` | Cross stubs for Register + Publish calls |

> **Principle**: Any infrastructure code shared across services lives in
> SharedLib. CodeValdDT retains only domain logic, domain errors, gRPC
> handlers, and storage collection schemas.

See task DT-012 in [mvp.md](../3-SofwareDevelopment/mvp.md) for SharedLib
migration scope.
| Repo granularity | 1 repo per Agency | Mirrors CodeValdCortex's database-per-agency isolation |
| Agent write policy | Always on a branch, never `main` | Prevents concurrent agent writes from corrupting shared history |
| Branch naming | `task/{task-id}` | Short-lived, traceable back to CodeValdCortex task records |
| Merge strategy | Auto-merge on task completion | No human approval gate for now; policy layer can extend this later |
| Storage backend | Pluggable via `storage.Storer` interface | go-git's open/closed design; caller injects the storer — filesystem and ArangoDB are both valid implementations |
| Worktree filesystem | Pluggable via `billy.Filesystem` interface | go-git separates object storage from the working tree; both are independently injectable |

---

## 2. Storage Backends

### go-git Pluggable Interfaces

go-git separates storage into two injectable interfaces:

| Interface | Package | Purpose |
|---|---|---|
| `storage.Storer` | `github.com/go-git/go-git/v5/storage` | Git objects, refs, index, config |
| `billy.Filesystem` | `github.com/go-git/go-billy/v5` | Working tree (checked-out files) |

### CodeValdGit `Backend` Interface

CodeValdGit adds a thin `Backend` interface on top of `storage.Storer`. It captures the operations that differ per storage type — repo lifecycle (init, archive/flag, purge) and storer construction — while the shared `Repo` implementation (branches, files, history) sits in `internal/repo/` and is backend-agnostic.

```go
// Backend abstracts storage-specific repo lifecycle.
// Implemented by storage/filesystem and storage/arangodb.
type Backend interface {
    // InitRepo provisions a new store for agencyID.
    InitRepo(ctx context.Context, agencyID string) error
    // OpenStorer returns a go-git storage.Storer and billy.Filesystem for agencyID.
    OpenStorer(ctx context.Context, agencyID string) (storage.Storer, billy.Filesystem, error)
    // DeleteRepo archives or flags the repo as deleted (behaviour is backend-specific).
    DeleteRepo(ctx context.Context, agencyID string) error
    // PurgeRepo permanently removes all storage for agencyID.
    PurgeRepo(ctx context.Context, agencyID string) error
}
```

The single `repoManager` implementation in `internal/manager/` holds a `Backend` and delegates lifecycle calls to it. `NewRepoManager(b Backend)` is the sole constructor — the caller (CodeValdCortex) picks and constructs the backend.

### Filesystem Backend (`storage/filesystem/`)

```
{base_path}/
└── {agency-id}/          ← One real .git repo per Agency
    └── .git/
```

| Operation | Implementation |
|---|---|
| `InitRepo` | `git.PlainInit` on disk; empty commit on `main` |
| `DeleteRepo` | `os.Rename` to `{archive_path}/{agency-id}/` (non-destructive) |
| `PurgeRepo` | `os.RemoveAll` of archive directory |
| `OpenStorer` | `filesystem.NewStorage` + `osfs.New` |

Simple, portable, works on any mounted volume (local disk, PVC, NFS).

### ArangoDB Backend (`storage/arangodb/`)

| Operation | Implementation |
|---|---|
| `InitRepo` | Insert seed documents into `git_objects`, `git_refs`, `git_config`, `git_index` |
| `DeleteRepo` | Set `deleted: true` flag on all agency documents (non-destructive; auditable) |
| `PurgeRepo` | Delete all documents where `agencyID == target` from all four collections |
| `OpenStorer` | `arango.NewStorage(db, agencyID)` + `memfs.New()` (or `osfs` for a durable worktree) |

The working tree (`billy.Filesystem`) remains on a local or in-memory filesystem — only the Git object store moves to ArangoDB. This mirrors the existing database-per-agency model in CodeValdCortex and means repos survive container restarts without a mounted volume.

| Collection | Contents |
|---|---|
| `git_objects` | Encoded Git objects (blobs, trees, commits, tags) keyed by SHA |
| `git_refs` | Branch and tag references |
| `git_index` | Staging area index |
| `git_config` | Per-repo Git config |

> **Selection**: The caller (CodeValdCortex) constructs the desired `Backend` implementation and passes it to `NewRepoManager`. CodeValdGit's core logic is backend-agnostic.

### Package Layout

```
github.com/aosanya/CodeValdGit/
├── codevaldgit.go          # RepoManager + Repo + Backend interfaces
├── types.go                # FileEntry, Commit, FileDiff, AuthorInfo, ErrMergeConflict
├── errors.go               # Sentinel errors (ErrRepoNotFound, ErrBranchNotFound, etc.)
├── config.go               # NewRepoManager constructor
├── internal/
│   ├── manager/            # Concrete repoManager — implements RepoManager, delegates to Backend
│   ├── repo/               # Shared Repo implementation — used by both storage backends
│   └── gitutil/            # Shared go-git helper utilities
└── storage/
    ├── filesystem/         # NewFilesystemBackend() — implements Backend (filesystem lifecycle)
    └── arangodb/           # NewArangoBackend()    — implements Backend (ArangoDB lifecycle)
```

---

## 3. Repository Identity

Naming convention: the Agency ID is the repository key in both backends.
- Filesystem: `{base_path}/{agency-id}/.git`
- ArangoDB: documents in `git_objects` etc. carry an `agency_id` field as the partition key (mirrors the existing database-per-agency isolation).

---

## 4. Branching Model

```
main
 │
 ├── task/task-abc-001     ← Agent A works here
 │     commits...
 │     └── auto-merged → main on task completion
 │
 └── task/task-xyz-002     ← Agent B works here (concurrent, isolated)
       commits...
       └── auto-merged → main on task completion
```

### Branch Lifecycle
1. **Task starts** → `CreateBranch("task/{task-id}", from: "main")`
2. **Agent writes files** → `Commit(branch: "task/{task-id}", files, author, message)`
3. **Task completes** → `MergeBranch("task/{task-id}", into: "main")`
   - If fast-forward is possible → merge directly
   - If `main` has advanced → **auto-rebase** task branch onto `main`, then fast-forward merge
   - If rebase conflicts → return `ErrMergeConflict{Files: [...]}` to caller; branch left clean for retry
4. **Branch deleted** → `DeleteBranch("task/{task-id}")`

> **Implementation note**: go-git only supports `FastForwardMerge`. The rebase step must be implemented by cherry-picking commits from the task branch onto the latest `main` using go-git's plumbing layer (`object.Commit`, `Worktree.Commit`).

---

## 5. Proposed Library API (Draft)

```go
// Backend abstracts storage-specific repo lifecycle.
// Implemented by storage/filesystem and storage/arangodb.
// The caller constructs the desired backend and passes it to NewRepoManager.
type Backend interface {
    InitRepo(ctx context.Context, agencyID string) error
    OpenStorer(ctx context.Context, agencyID string) (storage.Storer, billy.Filesystem, error)
    DeleteRepo(ctx context.Context, agencyID string) error
    PurgeRepo(ctx context.Context, agencyID string) error
}

// NewRepoManager constructs the shared RepoManager backed by the given Backend.
// Use storage/filesystem.NewFilesystemBackend or storage/arangodb.NewArangoBackend
// to obtain a Backend, then pass it here.
func NewRepoManager(b Backend) RepoManager

// RepoManager is the top-level entry point for managing per-agency Git repositories.
// Obtain via NewRepoManager. One instance is typically shared process-wide.
type RepoManager interface {
    InitRepo(ctx context.Context, agencyID string) error                   // delegates to Backend.InitRepo
    OpenRepo(ctx context.Context, agencyID string) (Repo, error)
    DeleteRepo(ctx context.Context, agencyID string) error                 // delegates to Backend.DeleteRepo
    PurgeRepo(ctx context.Context, agencyID string) error                  // delegates to Backend.PurgeRepo
}

// Repo represents a single agency's Git repository. Obtained via RepoManager.OpenRepo.
// Backed by internal/repo — backend-agnostic; works over any storage.Storer.
type Repo interface {
    // Branch operations
    CreateBranch(ctx context.Context, taskID string) error
    MergeBranch(ctx context.Context, taskID string) error
    DeleteBranch(ctx context.Context, taskID string) error

    // File operations (always on a task branch)
    WriteFile(ctx context.Context, taskID, path, content, author, message string) error
    ReadFile(ctx context.Context, ref, path string) (string, error)
    DeleteFile(ctx context.Context, taskID, path, author, message string) error
    ListDirectory(ctx context.Context, ref, path string) ([]FileEntry, error)

    // History
    Log(ctx context.Context, ref, path string) ([]Commit, error)
    Diff(ctx context.Context, fromRef, toRef string) ([]FileDiff, error)
}
```

---

## 6. Integration with CodeValdCortex

CodeValdCortex will call CodeValdGit at these lifecycle points:

| CodeValdCortex Event | CodeValdGit Call |
|---|---|
| Agency created | `RepoManager.InitRepo(agencyID)` |
| Task started | `Repo.CreateBranch(taskID)` |
| Agent writes output | `Repo.WriteFile(taskID, path, content, ...)` |
| Task completed | `Repo.MergeBranch(taskID)` → `Repo.DeleteBranch(taskID)` |
| Agency deleted | `RepoManager.DeleteRepo(agencyID)` |
| UI file browser | `Repo.ListDirectory("main", path)` |
| UI file view | `Repo.ReadFile("main", path)` |
| UI history view | `Repo.Log("main", path)` |

---

## 7. CodeValdSharedLib Dependency

CodeValdGit imports `github.com/aosanya/CodeValdSharedLib` for:

| SharedLib package | Replaces |
|---|---|
| `registrar` | `internal/registrar/registrar.go` (identical struct; service-specific metadata passed as constructor args) |
| `serverutil` | `envOrDefault`, `parseDuration` helpers and gRPC server setup block in `cmd/server/main.go` |
| `arangoutil` | ArangoDB `driverhttp.NewConnection` / auth / database bootstrap in `storage/arangodb/arangodb.go` |
| `gen/go/codevaldcross/v1` | Local copy of generated Cross stubs in `gen/go/codevaldcross/v1/` and `cmd/cross.go` |

> **Principle**: Any infrastructure code used by more than one service lives in
> SharedLib. CodeValdGit retains only domain logic, domain errors, gRPC
> handlers, and storage collection schemas.

See task MVP-GIT-012 in [mvp.md](../3-SofwareDevelopment/mvp.md) for migration scope.

---

## 8. What Gets Removed from CodeValdCortex

Once CodeValdGit is integrated, the following will be deleted:

- `internal/git/ops/operations.go` — custom SHA-1 blob/tree/commit engine
- `internal/git/storage/repository.go` — ArangoDB Git object storage
- `internal/git/fileindex/service.go` — ArangoDB file index service
- `internal/git/fileindex/repository.go` — ArangoDB file index repository
- `internal/git/models/` — custom Git object models
- ArangoDB collections: `git_objects`, `git_refs`, `repositories`
