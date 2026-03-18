# CodeValdDT — Architecture: ArangoDB Storage

> Part of [architecture.md](architecture.md)

## 7. ArangoDB Schema

### Shared Database

The ArangoDB database is **pre-existing** — CodeValdDT connects to it on
startup using `arangoutil.Connect` from SharedLib. It does **not** create the
database. The database name is supplied via the `DT_ARANGO_DATABASE`
environment variable (e.g. `codevald_demo`).

CodeValdDT is responsible for ensuring its collections and indexes exist on
startup (create-if-not-present on every boot — idempotent). It is **not**
responsible for database lifecycle.

All collection documents carry an `agencyID` field that scopes all reads and writes.

| Collection | Collection Type | Contents |
|---|---|---|
| `dt_schemas` | document | Versioned, immutable schema definitions (`TypeDefinition` list) |
| `dt_entities` | document | Entity instances; soft-deleted via `deleted` + `deletedAt` fields |
| `dt_relationships` | **edge** | Graph edges between entities — `_from` + `_to` ref `dt_entities/{id}` |
| `dt_telemetry` | document | Time-series telemetry readings |
| `dt_events` | document | Event log entries |

> ⚠️ `dt_relationships` **MUST** be created as an edge collection
> (`CollectionTypeEdge`). Creating it as a document collection prevents AQL
> graph traversal and breaks `TraverseGraph`. This is a one-time constraint —
> collection type cannot be changed after creation.

### Named Graph

A single named graph `dt_graph` is created in the shared database.
Because all agencies share the same vertex and edge collections, every
traversal query filters by `agencyID` and excludes soft-deleted vertices:

```
Graph: dt_graph
Edge collection:   relationships
Vertex collection: dt_entities
```

AQL traversal template (returns vertex + edge for each hop):

```
FOR v, e, p IN 1..@depth @direction @startVertex GRAPH 'dt_graph'
  FILTER v.agencyID == @agencyID AND v.deleted != true
  RETURN { vertex: v, edge: e }
```

### Document Shapes

**`dt_entities/{id}`**
```json
{
  "_key":       "entity-uuid",
  "agencyID":   "agency-123",
  "typeID":     "Pump",
  "properties": { "pressure": 4.2, "status": "running" },
  "createdAt":  "2026-01-01T00:00:00Z",
  "updatedAt":  "2026-01-01T00:00:00Z",
  "deleted":    false,
  "deletedAt":  null
}
```

**`dt_relationships/{id}`** (edge document)
```json
{
  "_key":       "rel-uuid",
  "_from":      "dt_entities/entity-uuid-1",
  "_to":        "dt_entities/entity-uuid-2",
  "agencyID":   "agency-123",
  "name":       "connects_to",
  "properties": {},
  "createdAt":  "2026-01-01T00:00:00Z"
}
```

**`dt_telemetry/{id}`**
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

**`dt_events/{id}`**
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
| `dt_entities` | `agencyID` | persistent | All entity queries scope by agency |
| `dt_entities` | `typeID` | persistent | ListEntities by type |
| `dt_entities` | `agencyID, deleted` | persistent | Efficiently exclude soft-deleted entities from list/traversal |
| `dt_relationships` | `agencyID` | persistent | Scope edge queries |
| `dt_telemetry` | `entityID, timestamp` | persistent | Time-range queries per entity |
| `dt_events` | `entityID, timestamp` | persistent | Chronological event log per entity |
