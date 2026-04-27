# High-Level Features

## Feature Overview

CodeValdDT provides the following top-level capabilities to the rest of the CodeVald platform:

---

### 1. Entity Lifecycle
- **Create / Read / Update / Delete / List** entity instances scoped by `agencyID`
- Each entity has a `typeID` (matches a `TypeDefinition.Name` in the agency's current `DTSchema`), a `properties` map, and timestamps
- **Soft delete** — `DeleteEntity` sets `deleted: true` + `deletedAt`; soft-deleted entities are excluded from `ListEntities` and `TraverseGraph`
- **Immutability guard** — `UpdateEntity` returns `ErrImmutableType` when the resolved `TypeDefinition.Immutable` is true

### 2. Graph Relationships
- **Directed edges** stored in an ArangoDB **edge collection** (`dt_relationships`), with `_from` / `_to` referencing `dt_entities/{id}`
- **Create / Get / Delete / List** relationships; filter by `FromID`, `ToID`, or relationship `Name`
- **Graph traversal** from a starting entity by `Direction` (`outbound` / `inbound` / `any`) and `Depth`, returning both vertices and edges in one round-trip

### 3. Telemetry
- Record telemetry readings (`name`, `value`, `timestamp`) against an entity
- Query historical telemetry per entity, optionally filtered by time range
- Every successful record fires `cross.dt.{agencyID}.telemetry.recorded` on CodeValdCross

### 4. Events
- Append events (`name`, `payload`, `timestamp`) to an entity's event log
- List events per entity in chronological order

### 5. Schema Management (DTDL v3 Compatible)
- `DTSchemaManager` owns `SetSchema`, `GetSchema`, `ListSchemaVersions` against the `dt_schemas` collection
- Versioned, immutable schema documents — one document per agency per version
- Each `TypeDefinition` carries `Properties`, optional `StorageCollection` (route writes to `dt_telemetry` / `dt_events`), and optional `Immutable` flag
- Schema model exports cleanly to Azure Digital Twins DTDL v3

### 6. Cross Integration
- Registers as service `codevalddt` on `:50055` with CodeValdCross every 20 seconds (liveness signal)
- Publishes `cross.dt.{agencyID}.entity.created` after every successful entity create
- Publishes `cross.dt.{agencyID}.telemetry.recorded` after every successful telemetry write

---

## What CodeValdDT Does NOT Do

| Out of Scope | Reason |
|---|---|
| Entity-type schema **enforcement** at write time | v1 trusts the caller; schema validation deferred |
| Live telemetry streaming (gRPC server-stream) | Deferred — pull-based queries only in v1 |
| Cascade delete of relationships / telemetry / events | Soft-delete only; orphan cleanup deferred to v2 |
| Hard delete | Not exposed in v1 |
| Task or work-item management | Lives in CodeValdWork |
| Git artifact management | Lives in CodeValdGit |
| AI agent orchestration | Lives in CodeValdAI |
| Authentication / access control | Handled by CodeValdCortex's policy layer |
