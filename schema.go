// Package codevalddt — pre-delivered digital-twin schema.
//
// This file exposes [DefaultDTSchema], which returns the fixed [types.Schema]
// for CodeValdDT. internal/app seeds this schema idempotently on startup via
// entitygraph.SeedSchema so the agency is bootstrapped and ready to accept
// runtime schema updates via [DTSchemaManager].
//
// The schema declares five TypeDefinitions:
//   - AssetLocation    — physical zone or location that contains equipment (mutable)
//   - Equipment        — any physical asset in a facility (pump, valve, motor, …) (mutable)
//   - Sensor           — measuring device attached to equipment (mutable)
//   - TelemetryReading — immutable sensor reading written via CreateEntity; stored in dt_telemetry
//   - EquipmentEvent   — immutable equipment lifecycle event; stored in dt_events
//
// Graph topology:
//
//	AssetLocation ──contains──────────► Equipment ──connects_to──► Equipment
//	              ◄──located_in────────
//
//	Equipment ──has_sensor──► Sensor
//	          ◄──attached_to──
//
// Time-series associations (property reference, not an edge):
//
//	TelemetryReading.entityID → equipment or sensor entity ID
//	EquipmentEvent.entityID   → equipment entity ID
//
// Storage:
//   - AssetLocation, Equipment, Sensor → dt_entities (mutable entity documents)
//   - TelemetryReading                 → dt_telemetry (immutable; high-write hot collection)
//   - EquipmentEvent                   → dt_events    (immutable; ordered event log)
//   - All edges                        → dt_relationships edge collection
//
// Inverse relationships auto-created by [entitygraph.DataManager.CreateRelationship]:
//
//	Equipment ──located_in──► AssetLocation (inverse of contains)
//	Sensor    ──attached_to─► Equipment     (inverse of has_sensor)
package codevalddt

import "github.com/aosanya/CodeValdSharedLib/types"

// DefaultDTSchema returns the pre-delivered [types.Schema] seeded on startup.
// The operation is idempotent — seeding the same schema ID multiple times is safe.
//
// All edges are stored in the dt_relationships edge collection.
// TelemetryReading and EquipmentEvent have Immutable == true;
// [DTDataManager.UpdateEntity] returns [ErrImmutableType] for those types.
func DefaultDTSchema() types.Schema {
	return types.Schema{
		ID:      "dt-schema-v1",
		Version: 1,
		Tag:     "v1",
		Types: []types.TypeDefinition{
			{
				Name:              "AssetLocation",
				DisplayName:       "Asset Location",
				PathSegment:       "asset-locations",
				EntityIDParam:     "assetLocationId",
				StorageCollection: "dt_entities",
				Properties: []types.PropertyDefinition{
					// name is the human-readable label for the location (e.g. "Boiler Room A").
					{Name: "name", Type: types.PropertyTypeString, Required: true},
					// zone is an optional grouping label (e.g. "north-wing", "outdoor").
					{Name: "zone", Type: types.PropertyTypeString},
					// description is a free-text summary of the location.
					{Name: "description", Type: types.PropertyTypeString},
					{Name: "created_at", Type: types.PropertyTypeDatetime},
					{Name: "updated_at", Type: types.PropertyTypeDatetime},
				},
				Relationships: []types.RelationshipDefinition{
					// contains links the location to all equipment installed within it.
					// An AssetLocation may contain zero or more Equipment instances.
					{
						Name:        "contains",
						Label:       "Equipment",
						PathSegment: "equipment",
						ToType:      "Equipment",
						ToMany:      true,
						Inverse:     "located_in",
					},
				},
			},
			{
				Name:              "Equipment",
				DisplayName:       "Equipment",
				PathSegment:       "equipment",
				EntityIDParam:     "equipmentId",
				StorageCollection: "dt_entities",
				Properties: []types.PropertyDefinition{
					// name is the human-readable label (e.g. "Feed Pump 1", "Pressure Valve 3").
					{Name: "name", Type: types.PropertyTypeString, Required: true},
					// serial_number is the manufacturer's unique hardware identifier.
					{Name: "serial_number", Type: types.PropertyTypeString},
					// model is the equipment model designation.
					{Name: "model", Type: types.PropertyTypeString},
					// status reflects the current operational state.
					// Valid values: "running" | "stopped" | "fault" | "maintenance".
					{
						Name:    "status",
						Type:    types.PropertyTypeOption,
						Options: []string{"running", "stopped", "fault", "maintenance"},
					},
					{Name: "created_at", Type: types.PropertyTypeDatetime},
					{Name: "updated_at", Type: types.PropertyTypeDatetime},
				},
				Relationships: []types.RelationshipDefinition{
					// located_in is the physical location where this equipment is installed.
					// Inverse of AssetLocation.contains.
					{
						Name:        "located_in",
						Label:       "Location",
						PathSegment: "location",
						ToType:      "AssetLocation",
						ToMany:      false,
						Inverse:     "contains",
					},
					// connects_to models a directed process flow between two pieces of
					// equipment (e.g. pump outlet → pipe inlet). ToMany allows fan-out.
					{
						Name:        "connects_to",
						Label:       "Connected Equipment",
						PathSegment: "connected-equipment",
						ToType:      "Equipment",
						ToMany:      true,
					},
					// has_sensor links the equipment to all sensors attached to it.
					// One equipment may carry zero or more sensors.
					{
						Name:        "has_sensor",
						Label:       "Sensors",
						PathSegment: "sensors",
						ToType:      "Sensor",
						ToMany:      true,
						Inverse:     "attached_to",
					},
				},
			},
			{
				Name:              "Sensor",
				DisplayName:       "Sensor",
				PathSegment:       "sensors",
				EntityIDParam:     "sensorId",
				StorageCollection: "dt_entities",
				Properties: []types.PropertyDefinition{
					// name is the human-readable label (e.g. "PT-101", "Flow Meter A").
					{Name: "name", Type: types.PropertyTypeString, Required: true},
					// sensor_type classifies the physical quantity measured
					// (e.g. "pressure", "temperature", "flow", "vibration").
					{Name: "sensor_type", Type: types.PropertyTypeString},
					// unit is the SI unit label for readings from this sensor
					// (e.g. "bar", "°C", "m³/h", "mm/s").
					{Name: "unit", Type: types.PropertyTypeString},
					{Name: "created_at", Type: types.PropertyTypeDatetime},
					{Name: "updated_at", Type: types.PropertyTypeDatetime},
				},
				Relationships: []types.RelationshipDefinition{
					// attached_to is the equipment this sensor is physically mounted on.
					// Inverse of Equipment.has_sensor.
					{
						Name:        "attached_to",
						Label:       "Equipment",
						PathSegment: "equipment",
						ToType:      "Equipment",
						ToMany:      false,
						Required:    true,
						Inverse:     "has_sensor",
					},
				},
			},
			{
				Name:              "TelemetryReading",
				DisplayName:       "Telemetry Reading",
				PathSegment:       "telemetry-readings",
				EntityIDParam:     "telemetryReadingId",
				StorageCollection: "dt_telemetry",
				// TelemetryReading is immutable — readings are facts; they cannot be
				// corrected in place. Use a new reading to supersede an erroneous one.
				// UpdateEntity returns ErrImmutableType for this type.
				Immutable: true,
				Properties: []types.PropertyDefinition{
					// entityID is the ID of the Equipment or Sensor entity that produced
					// this reading. Used as the primary filter key for time-range queries.
					{Name: "entityID", Type: types.PropertyTypeString, Required: true},
					// sensorID is the Sensor entity ID, when the reading originates from
					// a specific sensor rather than directly from the equipment.
					// Optional — omit for equipment-level readings.
					{Name: "sensorID", Type: types.PropertyTypeString},
					// value is the numeric measurement result.
					{Name: "value", Type: types.PropertyTypeNumber, Required: true},
					// unit is the SI unit label for this reading (e.g. "bar", "°C").
					// Should match the Sensor.unit of the producing sensor.
					{Name: "unit", Type: types.PropertyTypeString},
					// timestamp is the ISO 8601 datetime of the measurement.
					// Used as the sort key for time-range queries (SHAREDLIB-014).
					{Name: "timestamp", Type: types.PropertyTypeDatetime, Required: true},
				},
			},
			{
				Name:              "EquipmentEvent",
				DisplayName:       "Equipment Event",
				PathSegment:       "equipment-events",
				EntityIDParam:     "equipmentEventId",
				StorageCollection: "dt_events",
				// EquipmentEvent is immutable — events are an audit log; they cannot be
				// altered after recording. UpdateEntity returns ErrImmutableType.
				Immutable: true,
				Properties: []types.PropertyDefinition{
					// entityID is the ID of the Equipment entity that emitted this event.
					{Name: "entityID", Type: types.PropertyTypeString, Required: true},
					// event_type is a machine-readable label for the event category.
					// Use dot-namespaced identifiers (e.g. "valve.opened", "fault.detected",
					// "maintenance.started", "status.changed").
					{Name: "event_type", Type: types.PropertyTypeString, Required: true},
					// operator_id is the agent or user that triggered the event.
					// Optional — omit for automated/system-generated events.
					{Name: "operator_id", Type: types.PropertyTypeString},
					// payload is a JSON-encoded string carrying structured event detail.
					// Schema is event_type-specific and not enforced by the platform.
					{Name: "payload", Type: types.PropertyTypeString},
					// timestamp is the ISO 8601 datetime at which the event occurred.
					// Used as the sort key for chronological event-log queries (SHAREDLIB-014).
					{Name: "timestamp", Type: types.PropertyTypeDatetime, Required: true},
				},
			},
		},
	}
}
