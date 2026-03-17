---
agent: agent
---

# Refactor Go Code

Safe, incremental Go refactoring for **CodeValdDT**.

---

## When to Refactor

- File exceeds **500 lines** (hard limit)
- Function exceeds **50 lines**
- Multiple concerns in one file
- Business logic leaked into `cmd/main.go` or gRPC handler
- AI/LLM/frontend logic crept in — remove it

---

## Refactoring Workflow

### Step 1: Understand the File

```bash
wc -l internal/manager/manager.go
grep -n "^func " internal/manager/manager.go
```

### Step 2: Plan the Split

Typical file responsibilities for CodeValdDT:

```
errors.go                            # Sentinel error types
models.go                            # Entity, Relationship, TelemetryReading, Event, filters
codevalddt.go                        # DTManager + Backend interfaces
internal/manager/manager.go          # Concrete DTManager implementation
internal/server/server.go            # gRPC server handlers
internal/config/config.go            # Config loading
internal/registrar/registrar.go      # Cross registration heartbeat
storage/arangodb/storage.go          # ArangoDB Backend
cmd/main.go                          # Wiring only
```

If `storage/arangodb/storage.go` grows too large, split by concern:

```
storage/arangodb/entities.go         # Entity CRUD
storage/arangodb/relationships.go    # Edge collection operations
storage/arangodb/telemetry.go        # Telemetry insert + query
storage/arangodb/events.go           # Event insert + list
storage/arangodb/graph.go            # TraverseGraph AQL
storage/arangodb/db.go               # DB init + collection setup
```

### Step 3: Extract — One File at a Time

1. Create the new file with its package declaration
2. Move types / functions
3. Update imports
4. `go build ./...` — must succeed after each move
5. `go test -v -race ./...`

### Step 4: Validate

```bash
go build ./...
go vet ./...
go test -v -race ./...
golangci-lint run ./...
```

---

## Specific Concerns for CodeValdDT

### Edge collection must not be refactored into document collection
- `relationships` is always created as `CollectionTypeEdge` in ArangoDB
- Any refactor must preserve this — verify with a test

### Keep pub/sub in manager, not in storage
- `crossClient.Publish(...)` lives in `internal/manager/manager.go`
- `storage/arangodb/` only does data operations — no pub/sub calls

### Keep Cross registration separate from business logic
- Registration heartbeat lives in `internal/registrar/`
- Never mix heartbeat retry logic with `CreateEntity` business logic
