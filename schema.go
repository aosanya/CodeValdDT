// Package codevalddt — pre-delivered digital-twin schema scaffold.
//
// [DefaultDTSchema] returns the [types.Schema] seeded idempotently on startup
// via [entitygraph.SeedSchema]. It bootstraps the ArangoDB collections and
// installs the two platform meta-types that enable runtime domain configuration:
//
//   - TelemetryType — catalog entry for a user-defined telemetry stream.
//     Creating a TelemetryType entity causes the DT service to provision a
//     dedicated storage collection named dt_telemetry_{snake_case_plural(name)}.
//     Example: name="SensorReading" → collection dt_telemetry_sensor_readings.
//
//   - EventType — catalog entry for a user-defined event stream.
//     Creating an EventType entity causes the DT service to provision a
//     dedicated storage collection named dt_events_{snake_case_plural(name)}.
//     Example: name="ValveEvent" → collection dt_events_valve_events.
//
// All domain entity types (the "digital twins" the agency actually manages)
// are declared at runtime by the agency operator via [DTSchemaManager.SetSchema].
// DefaultDTSchema ships no domain types — it provides only the infrastructure
// needed to let the agency build its own schema.
//
// Storage routing is driven entirely by [types.TypeDefinition.StorageCollection]:
//
//	"" (empty)           → dt_entities    (mutable entity documents — default)
//	"dt_telemetry_*"     → user collection (immutable; high-write hot collection)
//	"dt_events_*"        → user collection (immutable; ordered event log)
//
// Immutability is per-type: any TypeDefinition with Immutable == true causes
// [DTDataManager.UpdateEntity] to return [ErrImmutableType].
package codevalddt

import "github.com/aosanya/CodeValdSharedLib/types"

// DefaultDTSchema returns the pre-delivered [types.Schema] seeded on startup.
// It installs TelemetryType and EventType — the platform meta-types that let
// agencies define their own telemetry and event streams at runtime.
// The operation is idempotent — seeding the same schema ID multiple times is safe.
func DefaultDTSchema() types.Schema {
	return types.Schema{
		ID:      "dt-schema-v1",
		Version: 1,
		Tag:     "v1",
		Types: []types.TypeDefinition{
			{
				Name:              "TelemetryType",
				DisplayName:       "Telemetry Type",
				PathSegment:       "telemetry-types",
				EntityIDParam:     "telemetryTypeId",
				StorageCollection: "dt_telemetry_types",
				Properties: []types.PropertyDefinition{
					// name is the TypeID used when writing readings of this stream.
					// Drives the derived collection name: dt_telemetry_{snake_plural(name)}.
					// Example: "SensorReading" → collection dt_telemetry_sensor_readings.
					{Name: "name", Type: types.PropertyTypeString, Required: true},
					// description is a human-readable summary of what this stream measures.
					{Name: "description", Type: types.PropertyTypeString},
					// unit is the default SI unit label for readings of this type (e.g. "°C", "bar").
					{Name: "unit", Type: types.PropertyTypeString},
					{Name: "created_at", Type: types.PropertyTypeDatetime},
					{Name: "updated_at", Type: types.PropertyTypeDatetime},
				},
			},
			{
				Name:              "EventType",
				DisplayName:       "Event Type",
				PathSegment:       "event-types",
				EntityIDParam:     "eventTypeId",
				StorageCollection: "dt_event_types",
				Properties: []types.PropertyDefinition{
					// name is the TypeID used when writing events of this stream.
					// Drives the derived collection name: dt_events_{snake_plural(name)}.
					// Example: "ValveEvent" → collection dt_events_valve_events.
					{Name: "name", Type: types.PropertyTypeString, Required: true},
					// description is a human-readable summary of what this event represents.
					{Name: "description", Type: types.PropertyTypeString},
					{Name: "created_at", Type: types.PropertyTypeDatetime},
					{Name: "updated_at", Type: types.PropertyTypeDatetime},
				},
			},
		},
	}
}
