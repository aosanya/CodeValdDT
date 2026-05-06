// Package codevalddt — pre-delivered digital-twin schema.
//
// [DefaultDTSchema] defines the canonical starter types seeded on startup via
// [entitygraph.SeedSchema]. Agencies begin with this schema and extend it by
// calling [DTSchemaManager.SetSchema] / Publish / Activate.
//
// # Storage routing
//
// Each [types.TypeDefinition.StorageCollection] determines the ArangoDB
// collection that entity writes land in:
//
//   - ""              → dt_entities  (mutable state — equipment, locations, sensors)
//   - "dt_telemetry"  → dt_telemetry (immutable time-series — TelemetryReading)
//   - "dt_events"     → dt_events    (immutable event log   — EquipmentEvent)
//
// [types.TypeDefinition.Immutable] == true causes [DTDataManager.UpdateEntity]
// to return [ErrImmutableType]; only CreateEntity and DeleteEntity are valid
// for those types.
//
// All edges live in the dt_relationships edge collection.
// The named graph is dt_graph.
package codevalddt

import "github.com/aosanya/CodeValdSharedLib/types"

// DefaultDTSchema returns the canonical [types.Schema] seeded on startup.
// It defines five TypeDefinitions covering the three storage tiers:
//
//   - AssetLocation, Equipment, Sensor → dt_entities (mutable state)
//   - TelemetryReading                 → dt_telemetry (immutable readings)
//   - EquipmentEvent                   → dt_events    (immutable event log)
//
// Agencies may publish a new schema version via [DTSchemaManager] to add,
// rename, or extend types; the seed call on startup is idempotent.
func DefaultDTSchema() types.Schema {
	return types.Schema{
		ID:      "dt-schema-v1",
		Version: 1,
		Tag:     "v1",
		Types: []types.TypeDefinition{
			assetLocationTypeDef(),
			equipmentTypeDef(),
			sensorTypeDef(),
			telemetryReadingTypeDef(),
			equipmentEventTypeDef(),
		},
	}
}

// assetLocationTypeDef describes a physical location or zone that contains
// equipment (e.g. a floor, room, outdoor pad, or pipeline segment).
func assetLocationTypeDef() types.TypeDefinition {
	return types.TypeDefinition{
		Name:          "AssetLocation",
		DisplayName:   "Asset Location",
		PathSegment:   "asset-locations",
		EntityIDParam: "assetLocationId",
		Properties: []types.PropertyDefinition{
			{Name: "name", Type: types.PropertyTypeString, Required: true},
			{Name: "zone", Type: types.PropertyTypeString},
			{Name: "description", Type: types.PropertyTypeString},
		},
		Relationships: []types.RelationshipDefinition{
			{
				Name:        "contains",
				Label:       "Contains",
				ToType:      "Equipment",
				ToMany:      true,
				PathSegment: "equipment",
			},
		},
	}
}

// equipmentTypeDef describes any physical piece of equipment in a facility
// (pump, valve, motor, turbine, …). Status is an option-typed property so
// callers get validation against a closed set of values at schema time.
func equipmentTypeDef() types.TypeDefinition {
	return types.TypeDefinition{
		Name:          "Equipment",
		DisplayName:   "Equipment",
		PathSegment:   "equipment",
		EntityIDParam: "equipmentId",
		Properties: []types.PropertyDefinition{
			{Name: "name", Type: types.PropertyTypeString, Required: true},
			{Name: "serial_number", Type: types.PropertyTypeString},
			{
				Name:    "status",
				Type:    types.PropertyTypeOption,
				Options: []string{"running", "stopped", "fault", "maintenance"},
			},
		},
		Relationships: []types.RelationshipDefinition{
			{
				Name:        "connects_to",
				Label:       "Connects To",
				ToType:      "Equipment",
				ToMany:      true,
				PathSegment: "connected-equipment",
			},
		},
	}
}

// sensorTypeDef describes a measuring device attached to a piece of equipment.
// Sensor readings are written separately as TelemetryReading entities.
func sensorTypeDef() types.TypeDefinition {
	return types.TypeDefinition{
		Name:          "Sensor",
		DisplayName:   "Sensor",
		PathSegment:   "sensors",
		EntityIDParam: "sensorId",
		Properties: []types.PropertyDefinition{
			{Name: "name", Type: types.PropertyTypeString, Required: true},
			{Name: "sensor_type", Type: types.PropertyTypeString},
			{Name: "unit", Type: types.PropertyTypeString},
		},
		Relationships: []types.RelationshipDefinition{
			{
				Name:   "attached_to",
				Label:  "Attached To",
				ToType: "Equipment",
			},
		},
	}
}

// telemetryReadingTypeDef is the immutable telemetry type. Each reading is
// written via CreateEntity and lands in the dt_telemetry collection.
// UpdateEntity returns ErrImmutableType for this type.
//
// Canonical properties:
//   - entityID  — the equipment or sensor that produced this reading
//   - sensorID  — the Sensor entity (optional; omit for equipment-level readings)
//   - value     — the numeric measurement value
//   - unit      — SI unit label (e.g. "bar", "°C", "m³/h")
//   - timestamp — ISO 8601 datetime of the measurement
func telemetryReadingTypeDef() types.TypeDefinition {
	return types.TypeDefinition{
		Name:              "TelemetryReading",
		DisplayName:       "Telemetry Reading",
		PathSegment:       "telemetry-readings",
		EntityIDParam:     "telemetryReadingId",
		StorageCollection: "dt_telemetry",
		Immutable:         true,
		Properties: []types.PropertyDefinition{
			{Name: "entityID", Type: types.PropertyTypeString, Required: true},
			{Name: "sensorID", Type: types.PropertyTypeString},
			{Name: "value", Type: types.PropertyTypeNumber, Required: true},
			{Name: "unit", Type: types.PropertyTypeString},
			{Name: "timestamp", Type: types.PropertyTypeDatetime, Required: true},
		},
	}
}

// equipmentEventTypeDef is the immutable event type. Each event is written via
// CreateEntity and lands in the dt_events collection. UpdateEntity returns
// ErrImmutableType for this type.
//
// Canonical properties:
//   - entityID    — the equipment that emitted the event
//   - event_type  — machine-readable label (e.g. "valve.opened", "fault.detected")
//   - operator_id — the agent or user that triggered the event (optional)
//   - payload     — JSON-encoded structured payload (optional)
//   - timestamp   — ISO 8601 datetime of the event
func equipmentEventTypeDef() types.TypeDefinition {
	return types.TypeDefinition{
		Name:              "EquipmentEvent",
		DisplayName:       "Equipment Event",
		PathSegment:       "equipment-events",
		EntityIDParam:     "equipmentEventId",
		StorageCollection: "dt_events",
		Immutable:         true,
		Properties: []types.PropertyDefinition{
			{Name: "entityID", Type: types.PropertyTypeString, Required: true},
			{Name: "event_type", Type: types.PropertyTypeString, Required: true},
			{Name: "operator_id", Type: types.PropertyTypeString},
			{Name: "payload", Type: types.PropertyTypeString},
			{Name: "timestamp", Type: types.PropertyTypeDatetime, Required: true},
		},
	}
}
