package arangodb

import (
	"strings"
	"testing"

	codevalddt "github.com/aosanya/CodeValdDT"
	"github.com/aosanya/CodeValdSharedLib/types"
)

// TestToSharedConfig_PinsDTNames pins the CodeValdDT-specific collection and
// graph names on every Backend constructed through this package. The
// architecture-storage.md table treats these names as a one-time, hard
// invariant — `dt_relationships` in particular is an *edge* collection whose
// type cannot be changed post-creation. If any of these strings drift, the
// whole DT corpus becomes unreachable.
func TestToSharedConfig_PinsDTNames(t *testing.T) {
	in := Config{
		Endpoint: "http://arangodb:8529",
		Username: "root",
		Password: "secret",
		Database: "codevald_demo",
		Schema:   types.Schema{ID: "schema-1", Version: 1},
	}

	got := toSharedConfig(in)

	if got.Endpoint != in.Endpoint || got.Username != in.Username ||
		got.Password != in.Password || got.Database != in.Database {
		t.Errorf("toSharedConfig: connection fields not propagated; got %+v want %+v", got, in)
	}
	if got.Schema.ID != in.Schema.ID || got.Schema.Version != in.Schema.Version {
		t.Errorf("toSharedConfig: schema not propagated; got %+v want %+v", got.Schema, in.Schema)
	}

	cases := []struct {
		field string
		got   string
		want  string
	}{
		{"EntityCollection", got.EntityCollection, "dt_entities"},
		{"RelCollection", got.RelCollection, "dt_relationships"},
		{"SchemasDraftCol", got.SchemasDraftCol, "dt_schemas_draft"},
		{"SchemasPublishedCol", got.SchemasPublishedCol, "dt_schemas_published"},
		{"GraphName", got.GraphName, "dt_graph"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("toSharedConfig.%s: got %q, want %q", c.field, c.got, c.want)
		}
	}
}

// TestNewBackend_RejectsEmptyDatabase verifies the explicit precondition that
// Database is required — without it we would silently connect to ArangoDB's
// `_system` database and pollute it with DT collections.
func TestNewBackend_RejectsEmptyDatabase(t *testing.T) {
	_, err := NewBackend(Config{Endpoint: "http://localhost:8529", Database: ""})
	if err == nil {
		t.Fatal("NewBackend with empty Database: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Database") {
		t.Errorf("NewBackend error should mention Database; got %q", err.Error())
	}
}

// TestNew_RejectsNilDB verifies the input guard on the in-process constructor.
func TestNew_RejectsNilDB(t *testing.T) {
	_, _, err := New(nil, codevalddt.DefaultDTSchema())
	if err == nil {
		t.Fatal("New(nil, ...): expected error, got nil")
	}
	if !strings.Contains(err.Error(), "database") {
		t.Errorf("New(nil) error should mention database; got %q", err.Error())
	}
}

// TestNewBackendFromDB_RejectsNilDB mirrors TestNew_RejectsNilDB for the
// test-friendly constructor.
func TestNewBackendFromDB_RejectsNilDB(t *testing.T) {
	_, err := NewBackendFromDB(nil, codevalddt.DefaultDTSchema())
	if err == nil {
		t.Fatal("NewBackendFromDB(nil, ...): expected error, got nil")
	}
	if !strings.Contains(err.Error(), "database") {
		t.Errorf("NewBackendFromDB(nil) error should mention database; got %q", err.Error())
	}
}
