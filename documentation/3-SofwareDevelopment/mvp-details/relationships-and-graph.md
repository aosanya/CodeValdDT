# Relationships & Graph Traversal

Topics: Edge Collection · Relationship CRUD · Named Graph · AQL Traversal · TraverseGraph

---

## Tasks

| Task | Status | Depends On |
|---|---|---|
| MVP-DT-002 — ArangoDB backend: `dt_relationships` edge collection + `dt_graph` named graph bootstrap | ✅ Done (2026-04-27) | MVP-DT-001 |
| MVP-DT-006 — Integration tests: relationship CRUD + traversal end-to-end | ✅ Done (2026-04-28) | MVP-DT-002 |

Architecture ref: [architecture-interfaces.md §2 & §3](../../2-SoftwareDesignAndArchitecture/architecture-interfaces.md),
[architecture-storage.md §7](../../2-SoftwareDesignAndArchitecture/architecture-storage.md)

---

## Overview

Relationships between entities are stored as native ArangoDB **edge collection**
documents in `dt_relationships`. Graph traversal runs AQL `FOR v, e, p IN ...`
queries on the named graph `dt_graph`.

The relationship RPCs delegate to `DTDataManager` (= `entitygraph.DataManager`):

| RPC | Go method | Notes |
|---|---|---|
| `CreateRelationship` | `dm.CreateRelationship` | Writes an edge doc with `_from` + `_to` |
| `GetRelationship` | `dm.GetRelationship` | Returns `ErrRelationshipNotFound` if not found |
| `DeleteRelationship` | `dm.DeleteRelationship` | Hard delete — edges have no soft-delete in v1 |
| `ListRelationships` | `dm.ListRelationships` | Filtered by `agencyID`, `fromID`, `toID`, and/or `name` |
| `TraverseGraph` | `dm.TraverseGraph` | Returns `TraverseGraphResult{Vertices, Edges}` |

---

## Acceptance Criteria

- [x] `dt_relationships` created as `CollectionTypeEdge` — **one-time constraint, cannot change after creation**
- [x] `dt_graph` named graph bootstrapped on startup (idempotent) with `dt_relationships` as edge collection and `dt_entities` as vertex collection
- [x] `CreateRelationship` writes `_from = "dt_entities/{fromID}"`, `_to = "dt_entities/{toID}"` as ArangoDB handles
- [x] `ListRelationships` supports filter by `agencyID`, `fromID`, `toID`, `name`; zero-value fields are ignored
- [x] `TraverseGraph` accepts `direction` ("inbound" | "outbound" | "any"), `depth` (≥ 1), and `startID`
- [x] `TraverseGraph` returns `TraverseGraphResult{Vertices []Entity, Edges []Relationship}` — both included so callers get edge properties without a second round-trip
- [x] Soft-deleted entities are excluded from `TraverseGraph` vertices
- [x] `(agencyID, name)` persistent index on `dt_relationships` for `ListRelationships` name filter
- [x] `go build ./...`, `go vet ./...`, `go test -race ./...` all pass

---

## Edge Collection Constraint

> ⚠️ `dt_relationships` **MUST** be created as an edge collection
> (`CollectionTypeEdge`). Creating it as a document collection prevents
> AQL graph traversal and breaks `TraverseGraph`. This is verified by the
> integration test suite against a real ArangoDB instance.

The SharedLib `entitygraph/arangodb` backend handles collection creation.
CodeValdDT's `storage/arangodb/storage.go` shim supplies the fixed collection names:

```go
// storage/arangodb/storage.go
func NewBackend(db driver.Database) entitygraph.Backend {
    return sharedadb.NewBackend(db, sharedadb.Config{
        EntitiesCollection:      "dt_entities",
        RelationshipsCollection: "dt_relationships",   // must be edge
        TelemetryCollection:     "dt_telemetry",
        EventsCollection:        "dt_events",
        SchemasCollection:       "dt_schemas",
        GraphName:               "dt_graph",
    })
}
```

---

## Named Graph

```
Graph: dt_graph
Edge collection:   dt_relationships
Vertex collection: dt_entities
```

The same graph is used for traversal across all agencies in the shared database.
All AQL queries filter by `agencyID` and exclude soft-deleted vertices.

### AQL Traversal Template

```aql
FOR v, e, p IN 1..@depth @direction @startVertex GRAPH 'dt_graph'
  FILTER v.agencyID == @agencyID AND v.deleted != true
  RETURN { vertex: v, edge: e }
```

The start vertex is resolved from the `startID` field:
`"dt_entities/" + req.StartID`.

---

## Data Model

### Relationship

```go
// Relationship is a directed graph edge between two entities.
// Stored in an ArangoDB edge collection — _from and _to reference entities/ documents.
type Relationship struct {
    ID         string
    AgencyID   string
    Name       string            // semantic label, e.g. "connects_to", "reports_to"
    FromID     string            // source entity ID
    ToID       string            // target entity ID
    Properties map[string]any
    CreatedAt  time.Time
}
```

### TraverseGraphResult

```go
// TraverseGraphResult is returned by TraverseGraph.
// Both visited vertices and traversed edges are returned so callers can
// inspect relationship names without a second round-trip.
// Soft-deleted entities are excluded from Vertices.
type TraverseGraphResult struct {
    Vertices []Entity
    Edges    []Relationship
}
```

---

## Document Shape

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

---

## Indexes (dt_relationships)

| Field(s) | Type | Purpose |
|---|---|---|
| `agencyID` | persistent | Scope all edge queries |
| `name` | persistent | `ListRelationships` filter by relationship type |

---

## Open Question — TraverseGraph Max-Depth Ceiling

**Parked** (2026-04-27): The current AQL template passes `@depth` through from
the caller unchanged. A server-side ceiling has not been decided:

- **Option A**: Clamp at N hops server-side (e.g. 10); return `InvalidArgument` above
- **Option B**: Trust the caller; document the risk in the API contract

Until resolved, the implementation passes `Depth` unchanged. Revisit when a
deep-graph use case or performance SLA is scoped.

---

## gRPC Error Code Mapping

| Go error | gRPC code |
|---|---|
| `ErrRelationshipNotFound` | `codes.NotFound` |
| `ErrInvalidRelationship` | `codes.InvalidArgument` |
| all others | `codes.Internal` |

---

## Tests

| Test | File | Coverage |
|---|---|---|
| `TestNewBackend_CollectionNames` | `storage/arangodb/storage_test.go` | Correct `dt_relationships` / `dt_graph` names passed to SharedLib |
| `TestEntityServer_CreateRelationship` | `internal/app/app_integration_test.go` | Edge created; `_from` / `_to` set correctly |
| `TestEntityServer_TraverseGraph` | `internal/app/app_integration_test.go` | Returns reachable vertices and edges; excludes deleted vertices |
| `TestEntityServer_ListRelationships_ByName` | `internal/app/app_integration_test.go` | Filter by `name` returns only matching edges |

Integration tests tagged `//go:build integration`; skip without `DT_ARANGO_ENDPOINT`.
