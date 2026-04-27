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
- Each entity has a `typeID` (name of an `EntityTypeDefinition` on the Agency),
  a free-form `properties` map, and timestamps
- Entity creation is allowed regardless of Agency publication status in v1

### FR-002: Graph Relationships
- Relationships between entities must be stored as **ArangoDB edge collection**
  documents with `_from` and `_to` pointing to entity document handles
- The `relationships` collection must be created as `CollectionTypeEdge`
- Relationships have a `name` (e.g. `connects_to`) and optional properties

### FR-003: Graph Traversal
- The service must support traversing the graph from a starting entity with
  configurable depth and direction (inbound, outbound, any)
- Traversal is implemented using ArangoDB AQL graph queries on a named graph
  that includes the `relationships` edge collection

### FR-004: Telemetry
- The service must support recording telemetry readings against an entity
  (name, value, timestamp)
- Historical telemetry must be queryable by entity, optionally filtered by
  time range
- After every successful `RecordTelemetry`, publish
  `cross.dt.{agencyID}.telemetry.recorded` via CodeValdCross

### FR-005: Events
- The service must support appending events to an entity's event log
  (name, payload, timestamp)
- Events must be listable per entity in chronological order

### FR-006: Pub/Sub (v1)
- After every successful `CreateEntity`, publish
  `cross.dt.{agencyID}.entity.created`
- After every successful `RecordTelemetry`, publish
  `cross.dt.{agencyID}.telemetry.recorded`

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

---

## 5. Open Questions (v1)

| Question | Decision |
|---|---|
| Schema enforcement at entity creation | Deferred — v1 trusts the caller |
| Live telemetry streaming (gRPC server-stream) | Deferred |
| Entity deletion cascade to relationships/telemetry/events | **Resolved — no cascade in v1.** `DeleteEntity` soft-deletes only the entity; its relationships, telemetry, and events are retained as-is. Orphan cleanup deferred to v2. |
| Soft delete vs. hard delete for entities | **Resolved — soft delete.** `DeleteEntity` sets `deleted: true` and `deletedAt` on the document. Hard delete is not exposed in v1. |
