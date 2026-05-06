package codevalddt_test

import (
	"testing"

	codevalddt "github.com/aosanya/CodeValdDT"
	"github.com/aosanya/CodeValdSharedLib/types"
)

func TestDefaultDTSchema_Metadata(t *testing.T) {
	got := codevalddt.DefaultDTSchema()
	if got.ID != "dt-schema-v1" {
		t.Errorf("ID: got %q, want %q", got.ID, "dt-schema-v1")
	}
	if got.Version != 1 {
		t.Errorf("Version: got %d, want 1", got.Version)
	}
	if got.Tag != "v1" {
		t.Errorf("Tag: got %q, want %q", got.Tag, "v1")
	}
}

func TestDefaultDTSchema_TypeCount(t *testing.T) {
	got := codevalddt.DefaultDTSchema()
	if len(got.Types) != 5 {
		t.Errorf("Types: got %d, want 5", len(got.Types))
	}
}

func TestDefaultDTSchema_TypeNames(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	want := []string{"AssetLocation", "Equipment", "Sensor", "TelemetryReading", "EquipmentEvent"}
	for i, name := range want {
		if i >= len(schema.Types) {
			t.Fatalf("schema has only %d types; missing %q", len(schema.Types), name)
		}
		if schema.Types[i].Name != name {
			t.Errorf("Types[%d].Name: got %q, want %q", i, schema.Types[i].Name, name)
		}
	}
}

func TestDefaultDTSchema_StorageCollections(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	byName := typesByName(schema)

	cases := []struct {
		typeName string
		wantColl string
		wantImm  bool
	}{
		{"AssetLocation", "", false},
		{"Equipment", "", false},
		{"Sensor", "", false},
		{"TelemetryReading", "dt_telemetry", true},
		{"EquipmentEvent", "dt_events", true},
	}
	for _, tc := range cases {
		td, ok := byName[tc.typeName]
		if !ok {
			t.Errorf("type %q not found in schema", tc.typeName)
			continue
		}
		if td.StorageCollection != tc.wantColl {
			t.Errorf("%s.StorageCollection: got %q, want %q", tc.typeName, td.StorageCollection, tc.wantColl)
		}
		if td.Immutable != tc.wantImm {
			t.Errorf("%s.Immutable: got %v, want %v", tc.typeName, td.Immutable, tc.wantImm)
		}
	}
}

func TestDefaultDTSchema_TelemetryReadingProperties(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	td := mustFindType(t, schema, "TelemetryReading")

	requiredProps := map[string]types.PropertyType{
		"entityID":  types.PropertyTypeString,
		"value":     types.PropertyTypeNumber,
		"timestamp": types.PropertyTypeDatetime,
	}
	propMap := make(map[string]types.PropertyDefinition)
	for _, p := range td.Properties {
		propMap[p.Name] = p
	}
	for name, wantType := range requiredProps {
		p, ok := propMap[name]
		if !ok {
			t.Errorf("TelemetryReading: missing required property %q", name)
			continue
		}
		if p.Type != wantType {
			t.Errorf("TelemetryReading.%s: type got %q, want %q", name, p.Type, wantType)
		}
		if !p.Required {
			t.Errorf("TelemetryReading.%s: want Required=true", name)
		}
	}
}

func TestDefaultDTSchema_EquipmentEventProperties(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	td := mustFindType(t, schema, "EquipmentEvent")

	requiredProps := map[string]types.PropertyType{
		"entityID":   types.PropertyTypeString,
		"event_type": types.PropertyTypeString,
		"timestamp":  types.PropertyTypeDatetime,
	}
	propMap := make(map[string]types.PropertyDefinition)
	for _, p := range td.Properties {
		propMap[p.Name] = p
	}
	for name, wantType := range requiredProps {
		p, ok := propMap[name]
		if !ok {
			t.Errorf("EquipmentEvent: missing required property %q", name)
			continue
		}
		if p.Type != wantType {
			t.Errorf("EquipmentEvent.%s: type got %q, want %q", name, p.Type, wantType)
		}
		if !p.Required {
			t.Errorf("EquipmentEvent.%s: want Required=true", name)
		}
	}
}

func TestDefaultDTSchema_EquipmentStatusOptions(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	td := mustFindType(t, schema, "Equipment")

	for _, p := range td.Properties {
		if p.Name == "status" {
			if p.Type != types.PropertyTypeOption {
				t.Errorf("Equipment.status type: got %q, want option", p.Type)
			}
			wantOpts := []string{"running", "stopped", "fault", "maintenance"}
			if len(p.Options) != len(wantOpts) {
				t.Errorf("Equipment.status options: got %v, want %v", p.Options, wantOpts)
			}
			return
		}
	}
	t.Error("Equipment: status property not found")
}

func TestDefaultDTSchema_Relationships(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	byName := typesByName(schema)

	cases := []struct {
		typeName string
		relName  string
		toType   string
		toMany   bool
	}{
		{"AssetLocation", "contains", "Equipment", true},
		{"Equipment", "connects_to", "Equipment", true},
		{"Sensor", "attached_to", "Equipment", false},
	}
	for _, tc := range cases {
		td, ok := byName[tc.typeName]
		if !ok {
			t.Errorf("type %q not found", tc.typeName)
			continue
		}
		found := false
		for _, rel := range td.Relationships {
			if rel.Name == tc.relName {
				found = true
				if rel.ToType != tc.toType {
					t.Errorf("%s.%s.ToType: got %q, want %q", tc.typeName, tc.relName, rel.ToType, tc.toType)
				}
				if rel.ToMany != tc.toMany {
					t.Errorf("%s.%s.ToMany: got %v, want %v", tc.typeName, tc.relName, rel.ToMany, tc.toMany)
				}
				if rel.Inverse != "" {
					t.Errorf("%s.%s.Inverse: must be empty (no bidirectional declaration)", tc.typeName, tc.relName)
				}
			}
		}
		if !found {
			t.Errorf("%s: relationship %q not found", tc.typeName, tc.relName)
		}
	}
}

func TestDefaultDTSchema_PathSegmentsUnique(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	seen := make(map[string]string)
	for _, td := range schema.Types {
		if td.PathSegment == "" {
			continue
		}
		if prev, ok := seen[td.PathSegment]; ok {
			t.Errorf("duplicate PathSegment %q on types %q and %q", td.PathSegment, prev, td.Name)
		}
		seen[td.PathSegment] = td.Name
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

func typesByName(schema types.Schema) map[string]types.TypeDefinition {
	m := make(map[string]types.TypeDefinition, len(schema.Types))
	for _, td := range schema.Types {
		m[td.Name] = td
	}
	return m
}

func mustFindType(t *testing.T, schema types.Schema, name string) types.TypeDefinition {
	t.Helper()
	for _, td := range schema.Types {
		if td.Name == name {
			return td
		}
	}
	t.Fatalf("type %q not found in schema", name)
	return types.TypeDefinition{}
}
