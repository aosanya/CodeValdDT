# CodeValdDT — Requirements

## 1. Purpose

CodeValdDT is a **Go gRPC microservice** that manages the Digital Twin layer
of the CodeVald platform.

A Digital Twin is a live, graph-structured model of an Agency's real-world
entities — machines, people, locations, industrial assets, individuals in a
population, or any typed object the agency cares about. CodeValdDT stores
those entities, the graph of relationships between them, telemetry readings,
and events.

---

## 2. Scope

### In Scope
- Full entity lifecycle (create, read, update, delete, list)
- Graph of relationships between entities (ArangoDB edge collection)
- Graph traversal by depth and direction
- Telemetry recording and historical querying per entity
- Event log recording and reading per entity
- Pub/sub event publishing via CodeValdCross
- Registration and liveness heartbeat with CodeValdCross
- DTDL v3 compatible schema model (Azure Digital Twins migration path)

### Out of Scope
- Entity type schema definition and enforcement — schema lives in CodeValdAgency; v1 trusts the caller
- Task or work item management (CodeValdWork)
- Git artifact management (CodeValdGit)
- AI agent orchestration (CodeValdAI)
- Real-time streaming subscriptions (deferred)

---

## 3. Functional Requirements

### FR-001: Entity Management
- CodeValdDT must support creating, reading, updating, deleting, and listing
  entity instances scoped by `agencyID`
- Each entity has a `typeID` (matching a `TypeDefinition.Name` in the agency's
  active DT schema), a `properties` map, and timestamps
- The agency declares its own `TypeDefinition`s at runtime via
  `DTSchemaManager.SetSchema` — no domain types are pre-wired in the service
- Entity creation is allowed regardless of Agency publication status in v1

### FR-002: Graph Relationships
- Relationships between entities must be stored as **ArangoDB edge collection**
  documents with `_from` and `_to` pointing to entity document handles
- The `relationships` collection must be created as `CollectionTypeEdge`
- Relationship names are **data-driven**: the agency declares
  `RelationshipDefinition`s on each `TypeDefinition` in the schema, then
  creates relationship instances at runtime via `CreateRelationship`
- The engine validates that the relationship `name` matches a declared
  `RelationshipDefinition.Name` on the source entity's `TypeDefinition`, and
  that the target entity's `TypeID` matches `RelationshipDefinition.ToType`
- Relationships have a `name` and optional `properties`

### FR-003: Graph Traversal
- The service must support traversing the graph from a starting entity with
  configurable depth and direction (inbound, outbound, any)
- Traversal is implemented using ArangoDB AQL graph queries on a named graph
  that includes the `relationships` edge collection

### FR-004: Telemetry
- Telemetry readings are recorded as `Entity` instances — not as a separate
  type — by calling `CreateEntity` with a `typeID` whose `TypeDefinition` has
  `StorageCollection: "dt_telemetry"` and `Immutable: true`. The reading's
  source `entityID`, `value`, and `timestamp` are carried in `properties`
- Historical telemetry must be queryable per producing entity, optionally
  filtered by time range, via `ListEntities` against the `dt_telemetry`
  collection
- The service must NOT expose a `RecordTelemetry` / `QueryTelemetry` RPC —
  telemetry travels through the entity API

### FR-005: Events
- Events are recorded as `Entity` instances — not as a separate type — by
  calling `CreateEntity` with a `typeID` whose `TypeDefinition` has
  `StorageCollection: "dt_events"` and `Immutable: true`. The event's source
  `entityID`, `payload`, and `timestamp` are carried in `properties`
- Events must be listable per producing entity in chronological order via
  `ListEntities` against the `dt_events` collection
- The service must NOT expose a `RecordEvent` / `ListEvents` RPC — events
  travel through the entity API

### FR-006: Pub/Sub (v1)
- After every successful `CreateEntity`, publish a Cross topic chosen by the
  resolved `TypeDefinition.StorageCollection`:
  - `dt_entities`  → `cross.dt.{agencyID}.entity.created`
  - `dt_telemetry` → `cross.dt.{agencyID}.telemetry.recorded`
  - `dt_events`    → `cross.dt.{agencyID}.event.recorded`
- Publish failures must be logged but not surfaced to the caller — the entity
  is already persisted

### FR-007: CodeValdCross Registration
- On startup, register with CodeValdCross using service name `codevalddt`
  and address `:50055`
- Repeat registration every 20 seconds as a liveness heartbeat

### FR-008: DTDL v3 Compatibility
- The data model must be exportable to Azure Digital Twins DTDL v3 format
- Entity types → DTDL `Interface`
- Properties → DTDL `Property`
- Telemetry → DTDL `Telemetry`
- Relationships → DTDL `Relationship` (stored as ArangoDB edge documents)
- Events → DTDL `Telemetry` with structured payload

---

## 4. Non-Functional Requirements

| NFR | Requirement |
|---|---|
| Language | Go 1.21+ |
| API | gRPC + protobuf |
| Storage | ArangoDB — single shared database (`DT_ARANGO_DATABASE` env var); collections scoped by `agencyID` field |
| Schema standard | DTDL v3 compatible |
| Context propagation | All exported methods take `context.Context` as first arg |
| Godoc | All exported symbols must have godoc comments |
| File size | Max 500 lines per file |
| Function size | Max 50 lines per function |
| Test coverage | All business logic covered with `-race` tests |
| No hardcoded storage | `Backend` interface injected via constructor |
| No cross-service imports | All cross-service calls go through gRPC |
| Telemetry write frequency | **Very High** — `dt_telemetry` is the hot collection; sustained ingestion must not block reads on `dt_entities` / `dt_relationships`. Implications: `Immutable: true` on every telemetry `TypeDefinition` (no UPDATE-on-write contention); composite `(properties.entityID, properties.timestamp)` index for time-range scans; retention policy is an open question (see §5) |

---

## 5. Open Questions (v1)

| Question | Decision |
|---|---|
| Property value enforcement at entity creation | Deferred — v1 trusts the caller on property values |
| Relationship name + ToType enforcement | **Resolved** — enforced by `entitygraph.ValidateCreateRelationship`; unknown name or mismatched ToType returns `ErrInvalidRelationship` |
| Live telemetry streaming (gRPC server-stream) | Deferred |
| Entity deletion cascade to relationships/telemetry/events | **Resolved — no cascade in v1.** `DeleteEntity` soft-deletes only the entity; its relationships, telemetry, and events are retained as-is. Orphan cleanup deferred to v2. |
| Soft delete vs. hard delete for entities | **Resolved — soft delete.** `DeleteEntity` sets `deleted: true` and `deletedAt` on the document. Hard delete is not exposed in v1. |
| Telemetry retention policy (TTL on `dt_telemetry`) | **Parked** (2026-04-27) — write frequency is **Very High** so `dt_telemetry` will grow fast. Options: keep all readings indefinitely, or apply an ArangoDB TTL index on `properties.timestamp` with a configurable window. Revisit when an Agency's traffic profile is in scope. Until resolved, MVP-DT-002 should bootstrap `dt_telemetry` **without** a TTL index — adding one later is non-destructive. |
| `TraverseGraph` max-depth ceiling | **Parked** (2026-04-27) — current AQL template uses an unbounded `@depth` parameter. Decision needed before deep-graph use cases land: clamp server-side (e.g. 10 hops, returning `InvalidArgument` above) or trust the caller. Until resolved, MVP-DT-004 should pass the caller's `Depth` through unchanged. |
| `EntityFilter` time-range + ordering | Tracked in CodeValdSharedLib as `SHAREDLIB-014`. Not a blocker for DT MVP scaffolding; required before FR-004 time-range queries are implementable. |
