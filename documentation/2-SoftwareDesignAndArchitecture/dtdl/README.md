# DTDL v3 — Digital Twins Definition Language

> **Source**: [Azure opendigitaltwins-dtdl](https://azure.github.io/opendigitaltwins-dtdl/DTDL/v3/DTDL.v3.html)

CodeValdDT follows DTDL v3 so that an Agency's schema can be exported to Azure
Digital Twins without rewriting.

---

## Document Index

| File | Contents |
|---|---|
| [interface.md](interface.md) | `Interface` — structure, inheritance, examples |
| [contents.md](contents.md) | `Telemetry`, `Property`, `Command`, `Relationship`, `Component` |
| [schemas.md](schemas.md) | Primitive and complex schemas (Array, Enum, Map, Object) |
| [additional.md](additional.md) | Versioning, DTMI format, localization, language extensions |

---

## What DTDL Is

DTDL is a JSON-based language that describes **what a digital twin looks like** —
its properties, the data it emits, the commands it accepts, and how it
connects to other twins.

The six building blocks:

| Concept | What it is |
|---|---|
| **Interface** | The type definition — like a class |
| **Property** | A stored value on the twin (e.g. `status`, `name`) |
| **Telemetry** | A streaming measurement (e.g. `temperature`, `pressure`) |
| **Command** | An operation you can call on the twin |
| **Relationship** | A named link to another twin — **graph edge in CodeValdDT** |
| **Component** | Embed another Interface inside this one (composition) |

---

## How CodeValdDT Uses DTDL

| DTDL concept | CodeValdDT equivalent |
|---|---|
| `Interface` | `EntityTypeDefinition` on the Agency |
| `Property` | `Entity.Properties` map field |
| `Telemetry` | `CreateEntity` with a `typeID` whose `TypeDefinition` has `StorageCollection: "dt_telemetry"` and `Immutable: true` — reading is stored as an `Entity` document in `dt_telemetry` |
| `Relationship` | `CreateRelationship` gRPC call → stored in ArangoDB **edge collection** (`dt_relationships`) |
| `Command` | Not implemented in v1 |
| `Component` | Not implemented in v1 |

---

## Context Declaration

Every DTDL document starts with this line — it tells parsers which version to use:

```json
{
  "@context": "dtmi:dtdl:context;3"
}
```

---

## Minimal Example

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

In CodeValdDT terms: `Thermostat` is an `EntityTypeDefinition`. The
`setPointTemp` field becomes an entry in `Entity.Properties` on a `Thermostat`
entity. The `temp` field becomes a separate telemetry `TypeDefinition` (e.g.
`ThermostatTemp`) with `StorageCollection: "dt_telemetry"` and `Immutable:
true` — readings are written as `Entity` instances via `CreateEntity`, with
the source thermostat's `entityID`, the `value`, and the `timestamp` carried
in `properties`.
