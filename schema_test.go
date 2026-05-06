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

// DefaultDTSchema ships two platform meta-types: TelemetryType and EventType.
// These are the only pre-wired types — all domain types are agency-defined at runtime.
func TestDefaultDTSchema_PlatformMetaTypes(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	want := []string{"TelemetryType", "EventType"}
	if len(schema.Types) != len(want) {
		t.Fatalf("Types: got %d, want %d", len(schema.Types), len(want))
	}
	for i, name := range want {
		if schema.Types[i].Name != name {
			t.Errorf("Types[%d].Name: got %q, want %q", i, schema.Types[i].Name, name)
		}
	}
}

func TestDefaultDTSchema_StorageCollections(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	cases := []struct {
		typeName string
		wantColl string
		wantImm  bool
	}{
		{"TelemetryType", "dt_telemetry_types", false},
		{"EventType", "dt_event_types", false},
	}
	byName := typesByName(schema)
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

func TestDefaultDTSchema_TelemetryTypeRequiredProperties(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	td := mustFindType(t, schema, "TelemetryType")
	propMap := propsByName(td)

	p, ok := propMap["name"]
	if !ok {
		t.Fatal("TelemetryType: missing property \"name\"")
	}
	if p.Type != types.PropertyTypeString {
		t.Errorf("TelemetryType.name: type got %q, want string", p.Type)
	}
	if !p.Required {
		t.Error("TelemetryType.name: want Required=true")
	}
}

func TestDefaultDTSchema_EventTypeRequiredProperties(t *testing.T) {
	schema := codevalddt.DefaultDTSchema()
	td := mustFindType(t, schema, "EventType")
	propMap := propsByName(td)

	p, ok := propMap["name"]
	if !ok {
		t.Fatal("EventType: missing property \"name\"")
	}
	if p.Type != types.PropertyTypeString {
		t.Errorf("EventType.name: type got %q, want string", p.Type)
	}
	if !p.Required {
		t.Error("EventType.name: want Required=true")
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

func propsByName(td types.TypeDefinition) map[string]types.PropertyDefinition {
	m := make(map[string]types.PropertyDefinition, len(td.Properties))
	for _, p := range td.Properties {
		m[p.Name] = p
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
