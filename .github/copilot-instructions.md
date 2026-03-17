````instructions
# CodeValdDT — AI Agent Development Instructions

## Project Overview

**CodeValdDT** is a **Go gRPC microservice** that manages the **Digital Twin**
layer of the CodeVald platform.

A Digital Twin is a live, graph-structured model of an Agency's real-world
entities — machines, people, locations, networks, or any typed object the
agency cares about. CodeValdDT owns entities, the graph of relationships
between them, telemetry readings, and events. It knows nothing about tasks,
git, AI agents, or communication. Those concerns belong to other services.

**Core Concept**: CodeValdDT has exactly one job — store and expose the digital
twin graph for an agency: entities, relationships (graph edges), telemetry, and
events.

---

## Service Architecture

> **Full architecture details live in the documentation.**
> See `documentation/2-SoftwareDesignAndArchitecture/architecture.md` for:
> - `DTManager` and `Backend` interface contracts with full method signatures
> - `Entity`, `Relationship`, `TelemetryReading`, `Event` data models
> - Project directory structure and file responsibilities
> - gRPC `DTService` proto definition and generated-code location
> - ArangoDB collection schema (one DB per agency: `entities`, `relationships`,
>   `telemetry`, `events`)
> - CodeValdCross `RegisterRequest` payload — service name, addr, topics
> - Heartbeat / liveness registration pattern

**Key invariants to keep in mind while coding:**

- CodeValdDT **never** imports CodeValdGit, CodeValdWork, CodeValdAgency, or
  CodeValdCross packages — gRPC only
- All cross-service communication flows through the `Register` RPC on
  CodeValdCross
- The `DTManager` interface is the only business-logic entry point — gRPC
  handlers delegate to it
- `relationships` is an **ArangoDB edge collection** — use ArangoDB graph
  traversal for `TraverseGraph`
- `cross.dt.{agencyID}.entity.created` **must** be published after every
  successful `CreateEntity`
- `cross.dt.{agencyID}.telemetry.recorded` **must** be published after every
  successful `RecordTelemetry`
- Storage backends are injected — the core library is backend-agnostic
- EntityType schema (DTDL Interface) lives in CodeValdAgency — CodeValdDT
  trusts the caller in v1 (no schema enforcement)

---

## Developer Workflows

### Build & Run Commands

```bash
# Build the binary
go build -o bin/codevalddt-server ./cmd/...

# Run the service
./bin/codevalddt-server

# Run all tests with race detector
go test -v -race ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Static analysis
go vet ./...

# Format code
go fmt ./...

# Lint
golangci-lint run ./...

# Regenerate protobuf (requires buf)
buf generate
```

### Task Management Workflow

**Every task follows strict branch management:**

```bash
# 1. Create feature branch from main
git checkout -b feature/DT-XXX_description

# 2. Implement changes

# 3. Build validation before merge
go build ./...           # Must succeed
go vet ./...             # Must show 0 issues
go test -v -race ./...   # Must pass
golangci-lint run ./...  # Must pass

# 4. Merge when complete
git checkout main
git merge feature/DT-XXX_description --no-ff
git branch -d feature/DT-XXX_description
```

---

## Technology Stack

| Component | Choice | Rationale |
|---|---|---|
| Language | Go 1.21+ | Matches all other CodeVald services |
| Service communication | gRPC + protobuf | Typed contracts; Cross dials DT via gRPC |
| Storage | ArangoDB | Native graph; edge collections for relationships |
| Schema standard | DTDL v3 compatible | Azure Digital Twins migration path |
| Configuration | YAML + env overrides | Consistent with other services |
| Registration | CodeValdCross `Register` RPC | Standard onboarding pattern |

---

## Code Quality Rules

### Service-Specific Rules

- **No business logic in `cmd/main.go`** — wire dependencies only; logic lives
  in `internal/`
- **Interface-first for `DTManager`** — the gRPC server holds the interface,
  not the concrete type
- **No direct imports of other CodeVald services** — all cross-service calls
  go through gRPC
- **All public functions must have godoc comments**
- **Context propagation** — every public method takes `context.Context` as
  first argument
- **`cross.dt.{agencyID}.entity.created` must be published** on every
  successful `CreateEntity`
- **`cross.dt.{agencyID}.telemetry.recorded` must be published** on every
  successful `RecordTelemetry`
- **`relationships` is an edge collection** — never store relationships in a
  regular document collection

### Naming Conventions

- **Package name**: `codevalddt` (root), `manager`, `server`, `config`,
  `arangodb`
- **Interfaces**: noun-only, no `I` prefix — `DTManager`, `Backend`
- **Branch naming**: `feature/DT-XXX_description` (lowercase with underscores)
- **gRPC service**: `DTService`
- **No abbreviations in exported names** — prefer `EntityID` over `EID`

### File Organisation

- **Max file size**: 500 lines (prefer smaller, focused files)
- **Max function length**: 50 lines (prefer 20-30)
- **One primary concern per file**
- **Error types in `errors.go`** — `ErrEntityNotFound`,
  `ErrRelationshipNotFound`
- **Value types in `models.go`** — `Entity`, `Relationship`,
  `TelemetryReading`, `Event`

### Anti-Patterns to Avoid

- ❌ **AI agent logic, LLM calls** — not in this service
- ❌ **Frontend serving, HTML templates** — belongs in CodeValdHi
- ❌ **Task or work item management** — belongs in CodeValdWork
- ❌ **Git operations** — belongs in CodeValdGit
- ❌ **Business logic in gRPC handlers** — handlers delegate to `DTManager`
- ❌ **Hardcoded storage** — inject `Backend` via constructor
- ❌ **Regular collection for relationships** — must use ArangoDB edge
  collection
- ❌ **Skipping pub/sub events** — always publish on entity creation and
  telemetry recording

---

## Integration with CodeValdCross

CodeValdDT registers with CodeValdCross on startup:

```go
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
```

Heartbeat: call `Register` every 20 seconds.
````
