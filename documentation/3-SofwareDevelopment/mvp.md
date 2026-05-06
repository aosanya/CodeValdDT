# MVP — Active Task Backlog

## Overview
- **Objective**: Deliver CodeValdDT as a production-ready standalone gRPC microservice exposing SharedLib's `EntityService` with ArangoDB persistence, digital-twin schema (`DefaultDTSchema`), and CodeValdCross registration.
- **Completed tasks**: see [`mvp_done.md`](mvp_done.md)
- **Detailed specs**: see [`mvp-details/`](mvp-details/)

## Workflow

### Completion Process (MANDATORY)
1. Implement and validate (`go build ./...`, `go vet ./...`, `go test -race ./...`)
2. Add row to `mvp_done.md`
3. Remove task from this file
4. Mark dependency references as `~~DT-XXX~~ ✅`
5. Merge feature branch to main and delete it

### Branch Management
```bash
git checkout -b feature/DT-XXX_description
# implement + validate
git checkout main
git merge feature/DT-XXX_description --no-ff
git branch -d feature/DT-XXX_description
```

### Status Legend
- 📋 **Not Started** — ready to begin (dependencies met)
- 🚀 **In Progress** — currently being worked on
- ⏸️ **Blocked** — waiting on dependencies

---

## Architecture Pivot

~~### MVP-DT-003 — gRPC Service Proto & Codegen~~ ❌ Withdrawn (2026-04-27)

~~### MVP-DT-004 — gRPC Server Implementation~~ ❌ Withdrawn (2026-04-27)

**Rationale**: CodeValdDT re-uses the shared `entitygraph.EntityService` from CodeValdSharedLib.
Agencies declare their own `TypeDefinition`s at runtime; storage routing is driven by
`TypeDefinition.StorageCollection`, making a DT-specific proto a duplication of the shared surface.
`internal/server/entity_server.go` re-exports `egserver.NewEntityServer`; `internal/app/app.go`
registers `entitygraphpb.RegisterEntityServiceServer`. Handler logic, `toGRPCError`, and
entity↔proto conversion all live in CodeValdSharedLib.

---

## P2: Telemetry Query Unblocking

### SHAREDLIB-014 — EntityFilter Time-Range + Default Ordering

| Task | Status | Depends On |
|------|--------|------------|
| SHAREDLIB-014: Add time-range filter and default-descending ordering to `EntityFilter` in CodeValdSharedLib; enables end-to-end FR-004 telemetry queries against `dt_telemetry` and `dt_events` | ⏸️ Blocked | CodeValdSharedLib team |

**Scope**: `EntityFilter` does not yet support time-range predicates (`after`, `before`) or
default descending ordering by creation timestamp. Until this lands, `ListEntities` calls against
`dt_telemetry` and `dt_events` return unordered full-collection scans; FR-004 time-range window
queries cannot be exercised end-to-end. No DT code changes are required — this is owned entirely
by the CodeValdSharedLib team.

See: [mvp-details/telemetry-and-events.md](mvp-details/telemetry-and-events.md)

---

## P2.5: Dynamic Collection Provisioning

### DT-010 — TelemetryType / EventType → Dynamic Collection Bootstrap

| Task | Status | Depends On |
|------|--------|------------|
| DT-010: When a `TelemetryType` or `EventType` entity is created, derive the collection name (`dt_telemetry_{snake_plural(name)}` / `dt_events_{snake_plural(name)}`), create that ArangoDB document collection if absent (with the standard indexes), and add the corresponding `TypeDefinition` (Immutable: true) to the agency's active schema via `DTSchemaManager.SetSchema` | 📋 Not Started | MVP-DT-002 |

**Scope**: Requires a post-`CreateEntity` hook inside `DTDataManager` that checks
`TypeID ∈ {"TelemetryType", "EventType"}` and triggers collection + schema provisioning.
SharedLib `entitygraph/arangodb` must support `EnsureCollection(ctx, name, indexes)` or
equivalent to allow the DT service to create collections at runtime.

See: [mvp-details/telemetry-and-events.md](mvp-details/telemetry-and-events.md)

---

## P3: Next Feature Batch (Post-MVP)

~~### DT-007 — DTDL v3 Export Endpoint (FR-008)~~ ✅

~~### DT-008 — TraverseGraph Max-Depth Ceiling~~ ✅

### DT-009 — Telemetry Retention TTL Policy

| Task | Status | Depends On |
|------|--------|------------|
| DT-009: Define and implement a retention TTL for `dt_telemetry` and `dt_events`; add ArangoDB TTL index or a periodic cleanup job | 📋 Not Started | SHAREDLIB-014 |

**Scope**: `dt_telemetry` and `dt_events` are append-only collections with no deletion path.
Decide retention window (e.g. 90 days), implement an ArangoDB TTL index on `createdAt`, and
expose the window as a config env var (`DT_TELEMETRY_TTL_DAYS`).

See: [mvp-details/telemetry-and-events.md](mvp-details/telemetry-and-events.md)

---

## Success Criteria

| # | Criterion |
|---|---|
| 1 | `go build ./...` succeeds |
| 2 | `go test -race ./...` all pass |
| 3 | `go vet ./...` shows 0 issues |
| 4 | EntityService RPCs (from CodeValdSharedLib) work end-to-end against ArangoDB through `app.Run`, with `CreateEntity` routing to `dt_entities` / `dt_telemetry` / `dt_events` based on `TypeDefinition.StorageCollection` and `UpdateEntity` returning `FailedPrecondition` for immutable types |
| 5 | `dt_relationships` created as **edge collection** (cannot be changed post-creation) |
| 6 | `dt_graph` named graph bootstrapped on startup (idempotent) |
| 7 | CodeValdCross registration fires on startup and repeats every 20 seconds |
| 8 | `DTDataManager` and `DTSchemaManager` injected via constructor — never hardcoded |
| 9 | No direct imports of other CodeVald services |
