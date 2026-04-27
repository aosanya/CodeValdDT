# DTDL v3 — Interface

> Part of the [DTDL v3 reference](README.md)

An **Interface** is the type definition for a digital twin — like a class in Go.
It lists the twin's properties, telemetry streams, commands, relationships, and
embedded components.

**In CodeValdDT**: each `EntityTypeDefinition` on the Agency corresponds to one
DTDL Interface. DTDL gives us a portable schema format; CodeValdDT stores the
actual entity instances.

---

## Key Fields

| Field | Required | Description |
|---|---|---|
| `@context` | yes | Always `"dtmi:dtdl:context;3"` |
| `@type` | yes | Always `"Interface"` |
| `@id` | yes | Unique DTMI identifier (max 128 chars) |
| `displayName` | no | Human-readable name |
| `contents` | no | Array of Telemetry, Property, Command, Relationship, Component elements |
| `extends` | no | Inherit from one or more other Interfaces |
| `schemas` | no | Reusable complex schema definitions for this Interface |

> The full text of a single Interface (excluding inherited content) must not exceed **1 MiByte**.
> The combined `contents` hierarchy (all nested elements) must not exceed **100,000 elements**.

---

## Examples

### Thermostat — Telemetry + Property

```json
{
  "@context": "dtmi:dtdl:context;3",
  "@id": "dtmi:com:example:Thermostat;1",
  "@type": "Interface",
  "displayName": "Thermostat",
  "contents": [
    {
      "@type": "Telemetry",
      "name": "temp",
      "schema": "double"
    },
    {
      "@type": "Property",
      "name": "setPointTemp",
      "writable": true,
      "schema": "double"
    }
  ]
}
```

**CodeValdDT mapping**: `setPointTemp` is a key in the `Properties` map of a
`Thermostat` entity. `temp` is realised as its own `TypeDefinition` (e.g.
`ThermostatTemp`) with `StorageCollection: "dt_telemetry"` and `Immutable:
true`; each reading is written as an `Entity` via `CreateEntity` and lands in
the `dt_telemetry` collection.

---

### Phone — Components

```json
{
  "@context": "dtmi:dtdl:context;3",
  "@id": "dtmi:com:example:Phone;2",
  "@type": "Interface",
  "displayName": "Phone",
  "contents": [
    {
      "@type": "Component",
      "name": "frontCamera",
      "schema": "dtmi:com:example:Camera;3"
    },
    {
      "@type": "Component",
      "name": "backCamera",
      "schema": "dtmi:com:example:Camera;3"
    }
  ]
}
```

**CodeValdDT note**: Components are not implemented in v1. This is a reference
example only.

---

### Building — Property + Relationship

```json
{
  "@context": "dtmi:dtdl:context;3",
  "@id": "dtmi:com:example:Building;1",
  "@type": "Interface",
  "displayName": "Building",
  "contents": [
    {
      "@type": "Property",
      "name": "name",
      "schema": "string",
      "writable": true
    },
    {
      "@type": "Relationship",
      "name": "contains",
      "target": "dtmi:com:example:Room;1"
    }
  ]
}
```

**CodeValdDT mapping**: The `contains` relationship becomes a `CreateRelationship`
call — stored as an ArangoDB edge in the `relationships` edge collection.
Traversal via `TraverseGraph` walks edges between entity documents.

---

### Inheritance — `extends`

`ConferenceRoom` inherits `occupied` from `Room` and adds `capacity`.

```json
[
  {
    "@context": "dtmi:dtdl:context;3",
    "@id": "dtmi:com:example:Room;1",
    "@type": "Interface",
    "contents": [
      { "@type": "Property", "name": "occupied", "schema": "boolean" }
    ]
  },
  {
    "@context": "dtmi:dtdl:context;3",
    "@id": "dtmi:com:example:ConferenceRoom;1",
    "@type": "Interface",
    "extends": "dtmi:com:example:Room;1",
    "contents": [
      { "@type": "Property", "name": "capacity", "schema": "integer" }
    ]
  }
]
```

`ConferenceRoom` is a subtype of `Room`. DTDL validators will check that
`ConferenceRoom` instances conform to `Room`'s schema too.
