package codevalddt_test

import (
	"testing"

	codevalddt "github.com/aosanya/CodeValdDT"
)

// TestDefaultDTSchema_IsEmptyScaffold pins the load-bearing invariant that DT
// ships no built-in TypeDefinitions. Agencies declare their own entity types,
// telemetry channels, and event channels at runtime via DTSchemaManager —
// adding a pre-baked TypeDefinition here would silently change agency
// onboarding behaviour. See codevalddt.go and architecture-flows.md §9.
func TestDefaultDTSchema_IsEmptyScaffold(t *testing.T) {
	got := codevalddt.DefaultDTSchema()

	if got.ID != "dt-schema-v1" {
		t.Errorf("DefaultDTSchema.ID: got %q, want %q", got.ID, "dt-schema-v1")
	}
	if got.Version != 1 {
		t.Errorf("DefaultDTSchema.Version: got %d, want 1", got.Version)
	}
	if got.Tag != "v1" {
		t.Errorf("DefaultDTSchema.Tag: got %q, want %q", got.Tag, "v1")
	}
	if got.Types != nil {
		t.Errorf("DefaultDTSchema.Types: got %d entries, want nil", len(got.Types))
	}
}
