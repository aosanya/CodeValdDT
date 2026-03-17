# CodeValdDT — Architecture: ArangoDB Storage

> Part of [architecture.md](architecture.md)

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
