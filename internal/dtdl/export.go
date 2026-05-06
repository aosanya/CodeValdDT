// Package dtdl converts a CodeValdDT [types.Schema] to DTDL v3 Interface documents.
//
// Export is a read-only, schema-level operation. No entity instances are
// involved. The result is a JSON array of DTDL v3 Interface objects, one per
// [types.TypeDefinition] in the schema.
//
// Mapping rules:
//   - TypeDefinition with StorageCollection "dt_telemetry" or "dt_events"
//     → all PropertyDefinitions exported as DTDL Telemetry items
//   - All other TypeDefinitions → PropertyDefinitions exported as DTDL Property
//     (writable: true)
//   - RelationshipDefinitions → DTDL Relationship; target is the DTMI of the
//     ToType if it is declared in the same schema.
//
// DTDL @id format: dtmi:codevald:{agencyID}:{TypeDefinition.Name};1
package dtdl

import (
	"encoding/json"
	"fmt"

	"github.com/aosanya/CodeValdSharedLib/types"
)

const (
	collTelemetry = "dt_telemetry"
	collEvents    = "dt_events"
)

// dtdlContent is one entry inside the DTDL Interface "contents" array.
// @type determines which fields are meaningful.
type dtdlContent struct {
	Type     string `json:"@type"`
	Name     string `json:"name"`
	Schema   any    `json:"schema,omitempty"`
	Writable *bool  `json:"writable,omitempty"`
	Target   string `json:"target,omitempty"`
}

// dtdlInterface is a DTDL v3 Interface document.
type dtdlInterface struct {
	Context     string        `json:"@context"`
	ID          string        `json:"@id"`
	Type        string        `json:"@type"`
	DisplayName string        `json:"displayName"`
	Contents    []dtdlContent `json:"contents,omitempty"`
}

// ExportSchema converts schema to a JSON-encoded array of DTDL v3 Interface
// documents, one per TypeDefinition. Returns an empty JSON array ([]) when
// schema.Types is empty. agencyID is used to build each Interface's DTMI @id.
func ExportSchema(agencyID string, schema types.Schema) ([]byte, error) {
	interfaces := make([]dtdlInterface, 0, len(schema.Types))

	dtmiByName := make(map[string]string, len(schema.Types))
	for _, td := range schema.Types {
		dtmiByName[td.Name] = dtmi(agencyID, td.Name)
	}

	for _, td := range schema.Types {
		interfaces = append(interfaces, toInterface(agencyID, td, dtmiByName))
	}

	return json.Marshal(interfaces)
}

func toInterface(agencyID string, td types.TypeDefinition, dtmiByName map[string]string) dtdlInterface {
	iface := dtdlInterface{
		Context:     "dtmi:dtdl:context;3",
		ID:          dtmi(agencyID, td.Name),
		Type:        "Interface",
		DisplayName: td.DisplayName,
	}
	if iface.DisplayName == "" {
		iface.DisplayName = td.Name
	}

	isTelemetry := td.StorageCollection == collTelemetry || td.StorageCollection == collEvents

	for _, prop := range td.Properties {
		iface.Contents = append(iface.Contents, toContent(prop, isTelemetry))
	}

	for _, rel := range td.Relationships {
		c := dtdlContent{Type: "Relationship", Name: rel.Name}
		if target, ok := dtmiByName[rel.ToType]; ok {
			c.Target = target
		}
		iface.Contents = append(iface.Contents, c)
	}

	return iface
}

func toContent(prop types.PropertyDefinition, asTelemetry bool) dtdlContent {
	if asTelemetry {
		return dtdlContent{
			Type:   "Telemetry",
			Name:   prop.Name,
			Schema: propSchema(prop),
		}
	}
	t := true
	return dtdlContent{
		Type:     "Property",
		Name:     prop.Name,
		Schema:   propSchema(prop),
		Writable: &t,
	}
}

// propSchema maps a PropertyDefinition to its DTDL schema value.
// Primitive types map to DTDL primitive schema strings.
// Arrays map to a DTDL Array schema object.
func propSchema(prop types.PropertyDefinition) any {
	switch prop.Type {
	case types.PropertyTypeInteger:
		return "integer"
	case types.PropertyTypeFloat, types.PropertyTypeNumber, types.PropertyTypeRating:
		return "double"
	case types.PropertyTypeBoolean:
		return "boolean"
	case types.PropertyTypeDate:
		return "date"
	case types.PropertyTypeDatetime:
		return "dateTime"
	case types.PropertyTypeArray:
		elem := elemSchema(prop.ElementType)
		return map[string]any{"@type": "Array", "elementSchema": elem}
	default:
		// string, uuid, option, select, multiselect → "string"
		return "string"
	}
}

func elemSchema(pt types.PropertyType) string {
	switch pt {
	case types.PropertyTypeInteger:
		return "integer"
	case types.PropertyTypeFloat, types.PropertyTypeNumber:
		return "double"
	case types.PropertyTypeBoolean:
		return "boolean"
	case types.PropertyTypeDate:
		return "date"
	case types.PropertyTypeDatetime:
		return "dateTime"
	default:
		return "string"
	}
}

func dtmi(agencyID, name string) string {
	return fmt.Sprintf("dtmi:codevald:%s:%s;1", agencyID, name)
}
