package dtdl

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aosanya/CodeValdSharedLib/types"
)

const agencyID = "agency-123"

func TestExportDTDL_EmptySchema(t *testing.T) {
	schema := types.Schema{Types: nil}
	data, err := ExportSchema(agencyID, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "[]" {
		t.Errorf("expected [], got %s", data)
	}
}

func TestExportDTDL_PropertyMapping(t *testing.T) {
	schema := types.Schema{
		Types: []types.TypeDefinition{
			{
				Name: "Pump",
				Properties: []types.PropertyDefinition{
					{Name: "status", Type: types.PropertyTypeString},
					{Name: "pressure", Type: types.PropertyTypeNumber},
					{Name: "count", Type: types.PropertyTypeInteger},
					{Name: "active", Type: types.PropertyTypeBoolean},
				},
			},
		},
	}

	data, err := ExportSchema(agencyID, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var interfaces []map[string]any
	if err := json.Unmarshal(data, &interfaces); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(interfaces))
	}

	contents, _ := interfaces[0]["contents"].([]any)
	schemaOf := func(name string) string {
		for _, c := range contents {
			m := c.(map[string]any)
			if m["name"] == name {
				return m["schema"].(string)
			}
		}
		t.Fatalf("property %q not found in contents", name)
		return ""
	}

	if got := schemaOf("status"); got != "string" {
		t.Errorf("string property: want schema=string, got %q", got)
	}
	if got := schemaOf("pressure"); got != "double" {
		t.Errorf("number property: want schema=double, got %q", got)
	}
	if got := schemaOf("count"); got != "integer" {
		t.Errorf("integer property: want schema=integer, got %q", got)
	}
	if got := schemaOf("active"); got != "boolean" {
		t.Errorf("boolean property: want schema=boolean, got %q", got)
	}
}

func TestExportDTDL_TelemetryRouting(t *testing.T) {
	schema := types.Schema{
		Types: []types.TypeDefinition{
			{
				Name:              "TemperatureReading",
				StorageCollection: "dt_telemetry",
				Properties: []types.PropertyDefinition{
					{Name: "value", Type: types.PropertyTypeNumber},
				},
			},
			{
				Name:              "SystemEvent",
				StorageCollection: "dt_events",
				Properties: []types.PropertyDefinition{
					{Name: "payload", Type: types.PropertyTypeString},
				},
			},
			{
				Name: "Pump",
				Properties: []types.PropertyDefinition{
					{Name: "status", Type: types.PropertyTypeString},
				},
			},
		},
	}

	data, err := ExportSchema(agencyID, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var interfaces []map[string]any
	if err := json.Unmarshal(data, &interfaces); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	byName := make(map[string]map[string]any)
	for _, iface := range interfaces {
		byName[iface["displayName"].(string)] = iface
	}

	checkContentsType := func(ifaceName, propName, wantType string) {
		t.Helper()
		iface, ok := byName[ifaceName]
		if !ok {
			t.Fatalf("interface %q not found", ifaceName)
		}
		contents := iface["contents"].([]any)
		for _, c := range contents {
			m := c.(map[string]any)
			if m["name"] == propName {
				if got := m["@type"]; got != wantType {
					t.Errorf("%s.%s: want @type=%q, got %q", ifaceName, propName, wantType, got)
				}
				_, hasWritable := m["writable"]
				if wantType == "Telemetry" && hasWritable {
					t.Errorf("%s.%s: Telemetry should not have writable field", ifaceName, propName)
				}
				return
			}
		}
		t.Fatalf("%s.%s: property not found in contents", ifaceName, propName)
	}

	checkContentsType("TemperatureReading", "value", "Telemetry")
	checkContentsType("SystemEvent", "payload", "Telemetry")
	checkContentsType("Pump", "status", "Property")
}

func TestExportDTDL_RelationshipMapping(t *testing.T) {
	schema := types.Schema{
		Types: []types.TypeDefinition{
			{Name: "Pipe"},
			{
				Name: "Pump",
				Relationships: []types.RelationshipDefinition{
					{Name: "connects_to", ToType: "Pipe"},
					{Name: "monitors", ToType: "Sensor"},
				},
			},
		},
	}

	data, err := ExportSchema(agencyID, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var interfaces []map[string]any
	if err := json.Unmarshal(data, &interfaces); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	var pump map[string]any
	for _, iface := range interfaces {
		if iface["displayName"] == "Pump" {
			pump = iface
		}
	}
	if pump == nil {
		t.Fatal("Pump interface not found")
	}

	contents := pump["contents"].([]any)
	relByName := make(map[string]map[string]any)
	for _, c := range contents {
		m := c.(map[string]any)
		if m["@type"] == "Relationship" {
			relByName[m["name"].(string)] = m
		}
	}

	// connects_to → known ToType "Pipe" → target populated
	ct, ok := relByName["connects_to"]
	if !ok {
		t.Fatal("connects_to relationship not found")
	}
	if ct["name"] != "connects_to" {
		t.Errorf("expected name=connects_to, got %q", ct["name"])
	}
	if got, ok := ct["target"]; !ok || got != dtmi(agencyID, "Pipe") {
		t.Errorf("connects_to: want target=%q, got %q", dtmi(agencyID, "Pipe"), got)
	}

	// monitors → unknown ToType "Sensor" → no target field
	mon, ok := relByName["monitors"]
	if !ok {
		t.Fatal("monitors relationship not found")
	}
	if _, hasTarget := mon["target"]; hasTarget {
		t.Errorf("monitors: target should be absent for unknown ToType")
	}
}

func TestExportDTDL_DTMI_Format(t *testing.T) {
	schema := types.Schema{
		Types: []types.TypeDefinition{
			{Name: "Pump"},
			{Name: "Pipe"},
		},
	}

	data, err := ExportSchema(agencyID, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var interfaces []map[string]any
	if err := json.Unmarshal(data, &interfaces); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for _, iface := range interfaces {
		id, _ := iface["@id"].(string)
		name, _ := iface["displayName"].(string)
		want := "dtmi:codevald:" + agencyID + ":" + name + ";1"
		if id != want {
			t.Errorf("@id: want %q, got %q", want, id)
		}
		if !strings.HasPrefix(id, "dtmi:") {
			t.Errorf("@id %q does not start with dtmi:", id)
		}
		if len(id) > 128 {
			t.Errorf("@id %q exceeds 128 characters (%d)", id, len(id))
		}
	}
}

func TestExportDTDL_DisplayNameFallback(t *testing.T) {
	schema := types.Schema{
		Types: []types.TypeDefinition{
			{Name: "Pump", DisplayName: ""},
			{Name: "Pipe", DisplayName: "Water Pipe"},
		},
	}

	data, err := ExportSchema(agencyID, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var interfaces []map[string]any
	if err := json.Unmarshal(data, &interfaces); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	byID := make(map[string]map[string]any)
	for _, iface := range interfaces {
		byID[iface["@id"].(string)] = iface
	}

	pump := byID[dtmi(agencyID, "Pump")]
	if pump["displayName"] != "Pump" {
		t.Errorf("expected displayName=Pump (fallback), got %q", pump["displayName"])
	}
	pipe := byID[dtmi(agencyID, "Pipe")]
	if pipe["displayName"] != "Water Pipe" {
		t.Errorf("expected displayName=Water Pipe, got %q", pipe["displayName"])
	}
}
