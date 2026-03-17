# DTDL v3 — Schemas

> Part of the [DTDL v3 reference](README.md)

A schema is the **data type** of a Property, Telemetry, or Command field.
Use a primitive schema for simple values; use a complex schema for structured
or enumerated data.

**CodeValdDT**: schemas define what goes into `Entity.Properties` values and
`TelemetryReading.Value`. CodeValdDT stores them as `any` in v1 — no
runtime validation.

---

## Primitive Schemas

Use these directly as the value of any `schema` field:

| Schema | What it is |
|---|---|
| `boolean` | `true` or `false` |
| `date` | ISO 8601 date (e.g. `"2026-01-15"`) |
| `dateTime` | ISO 8601 date + time (e.g. `"2026-01-15T10:30:00Z"`) |
| `double` | 64-bit float |
| `duration` | ISO 8601 duration (e.g. `"PT1H30M"`) |
| `float` | 32-bit float |
| `integer` | 32-bit signed integer |
| `long` | 64-bit signed integer |
| `string` | UTF-8 string |
| `time` | ISO 8601 time of day |

---

## Complex Schemas

Complex schemas let you build structured types from primitives.
Max nesting depth is **5 levels**.

---

## Array

An ordered list where every element has the same type.

### Example

```json
{
  "@type": "Telemetry",
  "name": "ledState",
  "schema": {
    "@type": "Array",
    "elementSchema": "boolean"
  }
}
```

Serialized: `"ledState": [ true, true, false, true ]`

---

## Enum

A set of named labels that map to integer or string values.

### Required fields

| Field | Description |
|---|---|
| `valueSchema` | Must be `integer` or `string` |
| `enumValues` | Array of `{ "name": "...", "enumValue": ... }` pairs |

### Example

```json
{
  "@type": "Telemetry",
  "name": "state",
  "schema": {
    "@type": "Enum",
    "valueSchema": "integer",
    "enumValues": [
      { "name": "offline", "enumValue": 1 },
      { "name": "online",  "enumValue": 2 }
    ]
  }
}
```

Serialized: `"state": 2`

---

## Map

A string-keyed dictionary where all values share the same type.

### Required fields

| Field | Description |
|---|---|
| `mapKey` | `{ "name": "...", "schema": "string" }` — key must always be `string` |
| `mapValue` | `{ "name": "...", "schema": <Schema> }` |

### Example

```json
{
  "@type": "Property",
  "name": "modules",
  "writable": true,
  "schema": {
    "@type": "Map",
    "mapKey":   { "name": "moduleName",  "schema": "string" },
    "mapValue": { "name": "moduleState", "schema": "string" }
  }
}
```

Serialized: `"modules": { "moduleA": "running", "moduleB": "stopped" }`

---

## Object

A named-field struct — like a Go struct inline.

### Example

```json
{
  "@type": "Telemetry",
  "name": "accelerometer",
  "schema": {
    "@type": "Object",
    "fields": [
      { "name": "x", "schema": "double" },
      { "name": "y", "schema": "double" },
      { "name": "z", "schema": "double" }
    ]
  }
}
```

Serialized: `"accelerometer": { "x": 12.7, "y": 5.5, "z": 19.1 }`

---

## Geospatial Schemas

GeoJSON-based schemas for location data:

| Short name | GeoJSON type |
|---|---|
| `point` | Point |
| `lineString` | LineString |
| `polygon` | Polygon |
| `multiPoint` | MultiPoint |
| `multiLineString` | MultiLineString |
| `multiPolygon` | MultiPolygon |

### Example

```json
{ "@type": "Telemetry", "name": "location", "schema": "point" }
```

Serialized: `"location": { "type": "Point", "coordinates": [ 47.64, -122.13 ] }`

---

## Reusable Schemas (Interface `schemas` array)

Define a complex schema once in the Interface's `schemas` array and reference it
by `@id` in multiple places — avoids repeating the same Object or Enum definition.

> A schema in `schemas` **must** have an explicit `@id`. It cannot be reused
> by inheriting Interfaces or by Components in a containing Interface.

### Example

```json
{
  "@context": "dtmi:dtdl:context;3",
  "@id": "dtmi:com:example:ReusableTypeExample;1",
  "@type": "Interface",
  "contents": [
    { "@type": "Telemetry", "name": "accelerometer1", "schema": "dtmi:com:example:acceleration;1" },
    { "@type": "Telemetry", "name": "accelerometer2", "schema": "dtmi:com:example:acceleration;1" }
  ],
  "schemas": [
    {
      "@id": "dtmi:com:example:acceleration;1",
      "@type": "Object",
      "fields": [
        { "name": "x", "schema": "double" },
        { "name": "y", "schema": "double" },
        { "name": "z", "schema": "double" }
      ]
    }
  ]
}
```
