# DTDL v3 — Interface Contents

> Part of the [DTDL v3 reference](README.md)

These are the five element types that go inside an Interface's `contents` array.

---

## Telemetry

Telemetry is data the twin **streams out** — sensor readings, computed values,
alerts. It is not stored in the twin itself; it flows through.

**CodeValdDT**: a DTDL `Telemetry` declaration on a parent `Interface` is
realised as its own `TypeDefinition` whose `StorageCollection` is
`"dt_telemetry"` and `Immutable` is `true`. Readings are written by calling
`CreateEntity` with that `typeID`; the source entity's `entityID`, the
reading `value`, and the `timestamp` are carried inside `properties`. The
record lands as an `Entity` document in `dt_telemetry`, and
`cross.dt.{agencyID}.telemetry.recorded` is published. There is no separate
`RecordTelemetry` RPC.

### Required fields

| Field | Description |
|---|---|
| `@type` | `"Telemetry"` |
| `name` | Programming name (alphanumeric + underscore, starts with letter) |
| `schema` | Data type — see [schemas.md](schemas.md) |

### Example

```json
{ "@type": "Telemetry", "name": "temp", "schema": "double" }
```

---

## Property

A Property is a **stored state value** on the twin — it persists between reads.

**CodeValdDT**: stored as an entry in `Entity.Properties`. Read with `GetEntity`;
update with `UpdateEntity`.

`writable: true` means callers can change it. `writable: false` (default) means
it is set by the twin itself (read-only from the outside).

### Required fields

| Field | Description |
|---|---|
| `@type` | `"Property"` |
| `name` | Programming name (unique within the Interface and within any Relationship) |
| `schema` | Data type |

### Optional fields

| Field | Description |
|---|---|
| `writable` | `true` = read-write; `false` (default) = read-only |

### Example

```json
{ "@type": "Property", "name": "setPointTemp", "schema": "double", "writable": true }
```

---

## Command

A Command is an **operation** you can call on the twin — like an RPC.

**CodeValdDT**: not implemented in v1. Listed here for DTDL completeness.

### Required fields

| Field | Description |
|---|---|
| `@type` | `"Command"` |
| `name` | Programming name |

### Optional fields

| Field | Description |
|---|---|
| `request` | Describes the input (`CommandRequest` — name + schema) |
| `response` | Describes the output (`CommandResponse` — name + schema) |

### Example

```json
{
  "@type": "Command",
  "name": "reboot",
  "request":  { "name": "rebootTime",    "schema": "dateTime" },
  "response": { "name": "scheduledTime", "schema": "dateTime" }
}
```

---

## Relationship

A Relationship is a **named directed link** from one twin to another.

**CodeValdDT**: call `CreateRelationship(fromID, toID, name, properties)`.
The link is stored as an ArangoDB **edge document** in the `relationships`
edge collection. Walk the graph with `TraverseGraph`.

> ⚠️ `relationships` **must** be an edge collection — not a document collection.
> AQL graph traversal only works with edge collections.

### Required fields

| Field | Description |
|---|---|
| `@type` | `"Relationship"` |
| `name` | Programming name (unique within the Interface) |

### Optional fields

| Field | Description |
|---|---|
| `target` | DTMI of the permitted target Interface; omit to allow any |
| `maxMultiplicity` | Maximum number of instances (default = unlimited) |
| `minMultiplicity` | Minimum number of instances (must be 0 if set) |
| `properties` | State properties on the relationship itself (e.g. `lastCleaned`) |
| `writable` | Whether callers can modify the relationship |

### Examples

**Simple — one Building contains many Rooms:**
```json
{ "@type": "Relationship", "name": "contains", "target": "dtmi:com:example:Room;1" }
```

**Bounded — a twin connects to at most one other twin:**
```json
{
  "@type": "Relationship",
  "name": "connectedTo",
  "maxMultiplicity": 1
}
```

**With state — cleaner relationship tracks last visit:**
```json
{
  "@type": "Relationship",
  "name": "cleanedBy",
  "target": "dtmi:com:example:Cleaner;1",
  "properties": [
    { "@type": "Property", "name": "lastCleaned", "schema": "dateTime" }
  ]
}
```

---

## Component

A Component **embeds another Interface by value** — its contents become part of
this Interface directly (as opposed to Relationship, which is by reference).

**CodeValdDT**: not implemented in v1. Relationships (graph edges) cover the
inter-entity connection use-case.

**Constraints:**
- No cycles allowed in Components.
- A Component cannot contain another Component (no nesting).

### Required fields

| Field | Description |
|---|---|
| `@type` | `"Component"` |
| `name` | Programming name |
| `schema` | The Interface DTMI to embed (must not contain another Component) |

### Example

```json
{ "@type": "Component", "name": "frontCamera", "schema": "dtmi:com:example:Camera;3" }
```
