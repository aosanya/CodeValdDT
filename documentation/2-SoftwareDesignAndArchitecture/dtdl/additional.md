# DTDL v3 — Additional Concerns

> Part of the [DTDL v3 reference](README.md)

---

## Model Versioning

Interfaces are versioned by the last segment of their DTMI:

| Form | Example |
|---|---|
| Integer version | `dtmi:com:example:Pump;1` |
| Major.minor version | `dtmi:com:example:Pump;1.2` |
| Unversioned | `dtmi:com:example:Pump` |

**Rule of thumb**: increment only the **minor** version for additive,
backward-compatible changes. Bump the **major** version for breaking changes.

---

## Digital Twin Model Identifier (DTMI)

A DTMI uniquely identifies every element in a DTDL model.

**Format**: `dtmi:<path>;<version>`

```
dtmi:com:fabrikam:industrialProducts:airQualitySensor;1
```

### Rules

- Path segments contain only letters, digits, and underscores.
- First character must not be a digit; last must not be an underscore.
- Max total length: 4096 chars. Interface DTMIs are capped at 128 chars.
- Prefixes `dtmi:dtdl:` and `dtmi:standard:` are reserved — do not use in custom models.

### Reusable schemas inside an Interface

Use the same prefix as the Interface, then add a unique path segment:

```
dtmi:com:fabrikam:industrialProducts:airQualitySensor:doubleArray;1
```

---

## Automatic Identifier Assignment

If you omit `@id` from an element, DTDL assigns one automatically:

1. **Single-value property** (e.g. `schema`) → parent DTMI + `_<propertyName>` before the version.
2. **Multi-value property** (e.g. `contents`) → parent DTMI + `_<propertyName>:__<elementName>` before the version.

These auto-IDs are built recursively from the parent outward. You rarely need
to know this — it matters mostly when tools export models.

---

## Display String Localization

`displayName` and `description` accept either a plain string (defaults to
English) or a JSON object keyed by BCP 47 language tags.

```json
"displayName": "Thermostat"
```

```json
"displayName": { "en": "Thermostat", "it": "Termostato" }
```

---

## Context

Every DTDL document must declare:

```json
"@context": "dtmi:dtdl:context;3"
```

When using language extensions, list their context **after** the core one:

```json
"@context": [
  "dtmi:dtdl:context;3",
  "dtmi:dtdl:extension:quantitativeTypes;1"
]
```

---

## Language Extensions

Optional extensions add functionality on top of the core language:

| Extension | What it adds |
|---|---|
| `QuantitativeTypes v1` | Standard units and semantic types (e.g. `Temperature`, `Pressure`) |
| `Historization v1/v2` | Record historical sequences of Property/Telemetry values |
| `Annotation v1/v2` | Add custom metadata to a Property or Telemetry |
| `Overriding v1/v2` | Override a model property with a per-instance value |
| `MQTT v1/v2/v3` | Pub/sub communication properties for MQTT |
| `Requirement v1` | Mark Object fields as required |

---

## What Changed from DTDL v2

- **Arrays in Properties** — now allowed.
- **Interface size limit** — 1 MiByte per Interface.
- **Element count** — one 100,000-element cap on the full `contents` hierarchy,
  replacing the old per-collection limits.
- **`extends` limit** — 1024 Interfaces across the hierarchy (max depth 10).
- **`CommandPayload` removed** — replaced by `CommandRequest` / `CommandResponse`.
- **Semantic Types** — moved out of core; use the `QuantitativeTypes` extension.
- **DTMI versioning** — now supports unversioned and `major.minor` forms.
