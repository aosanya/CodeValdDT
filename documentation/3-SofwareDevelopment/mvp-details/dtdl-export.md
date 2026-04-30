# DTDL v3 Export (FR-008)

Topics: DTDL v3 · Schema Export · Azure Digital Twins Migration Path

---

## Task

| Task | Status | Depends On |
|---|---|---|
| DT-007 — DTDL v3 export: `ExportDTDL(agencyID)` on `DTSchemaManager` + HTTP route | 📋 Not Started | — |

Architecture ref: [dtdl/README.md](../../2-SoftwareDesignAndArchitecture/dtdl/README.md),
[dtdl/interface.md](../../2-SoftwareDesignAndArchitecture/dtdl/interface.md),
[dtdl/contents.md](../../2-SoftwareDesignAndArchitecture/dtdl/contents.md)

---

## Overview

FR-008 requires the agency's CodeValdDT schema to be exportable to **DTDL v3**
so it can be ingested directly by Azure Digital Twins without rewriting.

Export is a **read-only, schema-level operation**. It reads the active
`types.Schema` for the agency from `DTSchemaManager` and translates each
`TypeDefinition` into a DTDL `Interface` document. No entity instances are
involved.

---

## DTDL Concept Mapping

| DTDL concept | CodeValdDT source | Notes |
|---|---|---|
| `Interface` (`@type: "Interface"`) | `TypeDefinition` | One DTDL Interface per TypeDefinition |
| `@id` | `dtmi:codevald:{agencyID}:{TypeDefinition.Name};1` | DTMI format; version suffix `;1` by default |
| `displayName` | `TypeDefinition.DisplayName` | Falls back to `TypeDefinition.Name` if empty |
| `Property` | Each `PropertyDefinition` where `TypeDefinition.StorageCollection ∈ {"", "dt_entities"}` | Writable properties on the entity |
| `Telemetry` | Each `PropertyDefinition` where `TypeDefinition.StorageCollection == "dt_telemetry"` | Emitted measurements — not stored on the twin |
| `Relationship` | `RelationshipDefinition.Name` entries on the TypeDefinition | Exported as DTDL `Relationship`; `target` set if `ToType` is known |
| `Command` | Not implemented in v1 | — |
| `Component` | Not implemented in v1 | — |

---

## DTDL Property Type Mapping

| `types.PropertyType` | DTDL schema |
|---|---|
| `PropertyTypeString` | `"string"` |
| `PropertyTypeNumber` | `"double"` |
| `PropertyTypeBoolean` | `"boolean"` |
| `PropertyTypeObject` | `{ "@type": "Object", "fields": [...] }` |
| `PropertyTypeArray` | `{ "@type": "Array", "elementSchema": "..." }` |

---

## Example Export

For a `TypeDefinition` named `Pump` with two properties (`pressure: number`,
`status: string`) and a relationship `connects_to`:

```json
{
  "@context": "dtmi:dtdl:context;3",
  "@id": "dtmi:codevald:agency-123:Pump;1",
  "@type": "Interface",
  "displayName": "Pump",
  "contents": [
    {
      "@type": "Property",
      "name": "pressure",
      "writable": true,
      "schema": "double"
    },
    {
      "@type": "Property",
      "name": "status",
      "writable": true,
      "schema": "string"
    },
    {
      "@type": "Relationship",
      "name": "connects_to"
    }
  ]
}
```

For a `TypeDefinition` named `TemperatureReading` with
`StorageCollection: "dt_telemetry"`:

```json
{
  "@context": "dtmi:dtdl:context;3",
  "@id": "dtmi:codevald:agency-123:TemperatureReading;1",
  "@type": "Interface",
  "displayName": "TemperatureReading",
  "contents": [
    {
      "@type": "Telemetry",
      "name": "value",
      "schema": "double"
    }
  ]
}
```

---

## Acceptance Criteria

- [ ] `DTSchemaManager` gains an `ExportDTDL(ctx, agencyID string) ([]byte, error)` method returning a JSON array of DTDL Interface documents
- [ ] Each `TypeDefinition` in the active schema maps to exactly one DTDL Interface
- [ ] `@id` follows the `dtmi:codevald:{agencyID}:{Name};1` pattern and is ≤ 128 characters
- [ ] `TypeDefinition.StorageCollection == "dt_telemetry"` → `PropertyDefinition` entries exported as DTDL `Telemetry` items, not `Property`
- [ ] `TypeDefinition.StorageCollection == "dt_events"` → exported as a separate Interface with `Telemetry` contents (structured payload)
- [ ] `RelationshipDefinition` entries exported as DTDL `Relationship` with `target` populated when `ToType` is a known Interface `@id`
- [ ] HTTP route `GET /{agencyId}/dt/schema/dtdl` registered with CodeValdCross
- [ ] Returned JSON validates against the DTDL v3 parser (Azure `dtdl-validator` tool)
- [ ] `ErrSchemaNotFound` returned as `codes.NotFound` when no active schema exists
- [ ] `go build ./...`, `go vet ./...`, `go test -race ./...` all pass

---

## Implementation Sketch

```go
// On DTSchemaManager (storage/arangodb or SharedLib)
func (m *schemaManager) ExportDTDL(ctx context.Context, agencyID string) ([]byte, error) {
    schema, err := m.GetSchema(ctx, agencyID, 0 /* active */)
    if err != nil {
        return nil, err
    }

    var interfaces []dtdlInterface
    for _, td := range schema.Types {
        iface := toDTDLInterface(agencyID, td)
        interfaces = append(interfaces, iface)
    }

    return json.Marshal(interfaces)
}

func toDTDLInterface(agencyID string, td types.TypeDefinition) dtdlInterface {
    iface := dtdlInterface{
        Context:     "dtmi:dtdl:context;3",
        ID:          fmt.Sprintf("dtmi:codevald:%s:%s;1", agencyID, td.Name),
        Type:        "Interface",
        DisplayName: td.DisplayName,
    }
    if iface.DisplayName == "" {
        iface.DisplayName = td.Name
    }

    for _, prop := range td.Properties {
        if td.StorageCollection == "dt_telemetry" {
            iface.Contents = append(iface.Contents, dtdlTelemetry(prop))
        } else {
            iface.Contents = append(iface.Contents, dtdlProperty(prop))
        }
    }
    for _, rel := range td.Relationships {
        iface.Contents = append(iface.Contents, dtdlRelationship(rel, agencyID))
    }
    return iface
}
```

---

## HTTP Route

```
GET /{agencyId}/dt/schema/dtdl
```

Response: `200 OK` with `Content-Type: application/json` body containing the
DTDL v3 Interface array.

Register in `internal/registrar/registrar.go` alongside the existing entity and
relationship routes:

```go
{Method: "GET", Pattern: "/{agencyId}/dt/schema/dtdl"},
```

---

## Design Constraints

- Export is **read-only** — no writes to `dt_schemas` or entity collections
- Export always uses the **active** schema version (`Schema.Active == true`)
- DTDL v3 `Interface` `@id` max length is **128 characters** — `agencyID` must be short enough; truncate or hash if needed
- `Command` and `Component` DTDL constructs are **not exported in v1**; document this in the API response or a warning field

---

## Tests

| Test | Coverage |
|---|---|
| `TestExportDTDL_EmptySchema` | Returns `[]` when active schema has no TypeDefinitions |
| `TestExportDTDL_PropertyMapping` | `PropertyTypeString` → `"string"`, `PropertyTypeNumber` → `"double"` |
| `TestExportDTDL_TelemetryRouting` | `StorageCollection: "dt_telemetry"` → DTDL `Telemetry`, not `Property` |
| `TestExportDTDL_RelationshipMapping` | `RelationshipDefinition` → DTDL `Relationship` with correct `name` |
| `TestExportDTDL_SchemaNotFound` | `ErrSchemaNotFound` when no active schema |
| `TestExportDTDL_DTMI_Format` | `@id` matches `dtmi:codevald:{agencyID}:{Name};1` pattern |
