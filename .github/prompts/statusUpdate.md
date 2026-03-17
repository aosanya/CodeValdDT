---
agent: agent
---

# CodeValdDT — Status Update

## Purpose
Record status updates and progress notes into topic files under:

```
CodeValdDT/documentation/3-SofwareDevelopment/status/
```

---

## CodeValdDT — Current Capabilities

### Role in the Platform
CodeValdDT is the **digital twin service** — it stores and exposes a
graph-structured model of an agency's real-world entities. Every entity,
relationship, telemetry reading, and event is scoped by `agencyID`.

### gRPC Endpoints (Inbound)

| Service | Method | Description |
|---|---|---|
| `DTService` | `CreateEntity` | Create an entity instance; publishes `cross.dt.{agencyID}.entity.created` |
| `DTService` | `GetEntity` | Get an entity by ID |
| `DTService` | `UpdateEntity` | Update entity property values |
| `DTService` | `DeleteEntity` | Delete an entity |
| `DTService` | `ListEntities` | List entities, optionally filtered by type |
| `DTService` | `CreateRelationship` | Create a graph edge between two entities |
| `DTService` | `DeleteRelationship` | Remove a graph edge |
| `DTService` | `TraverseGraph` | Walk the graph from a starting entity |
| `DTService` | `RecordTelemetry` | Record a telemetry reading; publishes `cross.dt.{agencyID}.telemetry.recorded` |
| `DTService` | `QueryTelemetry` | Query historical telemetry for an entity |
| `DTService` | `RecordEvent` | Append an event to an entity's log |
| `DTService` | `ListEvents` | Read an entity's event history |

### Pub/Sub (v1)

| Topic | Direction | Description |
|---|---|---|
| `cross.dt.{agencyID}.entity.created` | **produces** | After every successful `CreateEntity` |
| `cross.dt.{agencyID}.telemetry.recorded` | **produces** | After every successful `RecordTelemetry` |

### Key Design Properties
- **One ArangoDB database per agency** — consistent with platform convention
- **`relationships` is an edge collection** — ArangoDB native graph traversal
- **No schema enforcement in v1** — trust the caller
- **DTDL v3 compatible** — Azure Digital Twins migration path
- **Heartbeat** — `Register` called every 20 s via Cross

---

## Status File Rules

### Target directory
```
CodeValdDT/documentation/3-SofwareDevelopment/status/
```

### File size limit
- **≤ 400 lines** → write/append to a single topic file: `status/{topic}.md`
- **> 400 lines** → escalate to a subfolder with a `README.md` index

### Workflow

```bash
# Step 1 — Check existing file size
wc -l documentation/3-SofwareDevelopment/status/{topic}.md

# Step 2 — Choose write target
# ≤ 400 lines → append to existing file
# > 400 lines → create subfolder + README.md
```
