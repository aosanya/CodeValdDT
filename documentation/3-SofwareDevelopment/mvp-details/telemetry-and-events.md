# Telemetry & Events

Topics: Routed Entity Creates · dt_telemetry · dt_events · Immutability · Time-Range Queries · SHAREDLIB-014

---

## Tasks

| Task | Status | Depends On |
|---|---|---|
| MVP-DT-002 — `dt_telemetry` and `dt_events` collection bootstrap; `(properties.entityID, properties.timestamp)` indexes | ✅ Done (2026-04-27) | MVP-DT-001 |
| MVP-DT-006 — Integration tests: telemetry and event creation, immutability rejection | ✅ Done (2026-04-28) | MVP-DT-002 |
| SHAREDLIB-014 — `EntityFilter` time-range + default ordering for `dt_telemetry` / `dt_events` | ⏸️ Blocked (CodeValdSharedLib) | — |

Architecture ref: [architecture-flows.md §9](../../2-SoftwareDesignAndArchitecture/architecture-flows.md),
[architecture-storage.md §7](../../2-SoftwareDesignAndArchitecture/architecture-storage.md),
[architecture-interfaces.md §3](../../2-SoftwareDesignAndArchitecture/architecture-interfaces.md)

---

## Overview

Telemetry readings and events are **not** separate Go types or gRPC RPCs.
They are `Entity` documents written via the standard `CreateEntity` RPC, routed
to specialised collections by `TypeDefinition.StorageCollection`.

| Kind | `TypeDefinition.StorageCollection` | `TypeDefinition.Immutable` | Collection |
|---|---|---|---|
| Entity | `""` (empty) | false | `dt_entities` |
| Telemetry reading | `"dt_telemetry"` | **true** | `dt_telemetry` |
| Event log entry | `"dt_events"` | **true** | `dt_events` |

> There is no `RecordTelemetry`, `QueryTelemetry`, `RecordEvent`, or `ListEvents`
> RPC. Callers use `CreateEntity` to write and `ListEntities` to read — filtered
> by `TypeID` and (when SHAREDLIB-014 lands) by time range on `properties.timestamp`.

---

## Acceptance Criteria

- [x] `dt_telemetry` collection bootstrapped on startup (create-if-not-present, idempotent)
- [x] `dt_events` collection bootstrapped on startup (create-if-not-present, idempotent)
- [x] `CreateEntity` with a `TypeDefinition` whose `StorageCollection == "dt_telemetry"` writes to `dt_telemetry`
- [x] `CreateEntity` with a `TypeDefinition` whose `StorageCollection == "dt_events"` writes to `dt_events`
- [x] `UpdateEntity` returns `codes.FailedPrecondition` (`ErrImmutableType`) for any type with `Immutable: true`
- [x] `(properties.entityID, properties.timestamp)` persistent index exists on both `dt_telemetry` and `dt_events`
- [x] Cross topic `cross.dt.{agencyID}.telemetry.recorded` published after successful telemetry `CreateEntity`
- [x] Cross topic `cross.dt.{agencyID}.event.recorded` published after successful event `CreateEntity`
- [ ] `ListEntities` with time-range filter on `properties.timestamp` returns readings in `ASC` order — **blocked on SHAREDLIB-014**

---

## Design: User-Defined Telemetry and Event Types (Dynamic Collections)

Agencies define their own telemetry and event types **after deployment** by
creating `TelemetryType` and `EventType` entities. Each definition gets its
own dedicated storage collection — there is no shared `dt_telemetry` bucket.

### Step 1 — Register the type (once, at schema setup time)

Create a `TelemetryType` entity:

```json
{
  "agencyID": "agency-123",
  "typeID":   "TelemetryType",
  "properties": {
    "name":        "SensorReading",
    "description": "Raw sensor measurement from any attached sensor",
    "unit":        "varies"
  }
}
```

The DT service derives the collection name from `name`:
- `SensorReading` → `dt_telemetry_sensor_readings`
- `AccountBalance` → `dt_telemetry_account_balances`
- `GradePoint`     → `dt_telemetry_grade_points`

It then provisions that collection (if absent) and adds a corresponding
`TypeDefinition` to the agency's active schema with
`StorageCollection: "dt_telemetry_sensor_readings"` and `Immutable: true`.

### Step 2 — Write readings (at runtime)

Once registered, readings are written via the standard `CreateEntity` RPC:

```json
{
  "agencyID":   "agency-123",
  "typeID":     "SensorReading",
  "properties": {
    "entityID":  "sensor-entity-uuid",
    "value":     42.5,
    "unit":      "°C",
    "timestamp": "2026-01-01T12:00:00Z"
  }
}
```

The storage routing chain:
1. `DTSchemaManager.GetSchema` resolves `TypeDefinition` for `"SensorReading"`
2. `StorageCollection == "dt_telemetry_sensor_readings"` → write to that collection
3. `Immutable == true` → `UpdateEntity` returns `ErrImmutableType`
4. Cross topic `cross.dt.{agencyID}.telemetry.recorded` published

---

## How an Event Type Is Registered and Written

`EventType` follows the identical pattern:

```json
{
  "agencyID": "agency-123",
  "typeID":   "EventType",
  "properties": {
    "name":        "ValveEvent",
    "description": "Lifecycle events emitted by valve equipment"
  }
}
```

Derived collection: `ValveEvent` → `dt_events_valve_events`

Subsequent event writes use `typeID: "ValveEvent"` in `CreateEntity`.

---

## Collection Naming Convention

| TypeID registered | Derived collection |
|---|---|
| `SensorReading` | `dt_telemetry_sensor_readings` |
| `AccountBalance` | `dt_telemetry_account_balances` |
| `ValveEvent` | `dt_events_valve_events` |
| `StatusChange` | `dt_events_status_changes` |

Rule: `dt_telemetry_` + `snake_case(plural(name))` for TelemetryTypes;
`dt_events_` + `snake_case(plural(name))` for EventTypes.

---

## Default Ordering (Time-Series Collections)

When `ListEntities` resolves to a type whose `TypeDefinition.StorageCollection`
is `"dt_telemetry"` or `"dt_events"`, results MUST be sorted by
`properties.timestamp ASC`.

The `(properties.entityID, properties.timestamp)` composite index supports
time-range scans at no extra sort cost when filtered by `properties.entityID`.

This ordering contract is owned by `entitygraph.EntityFilter` in SharedLib —
tracked as **`SHAREDLIB-014`**.

---

## SHAREDLIB-014 Dependency

`SHAREDLIB-014` adds:
1. Time-range fields (`timestampFrom`, `timestampTo`) to `EntityFilter`
2. Automatic `properties.timestamp ASC` default ordering when `StorageCollection ∈ {dt_telemetry, dt_events}`

Until SHAREDLIB-014 lands:
- Telemetry and event **writes** work end-to-end
- `ListEntities` returns all readings for a `typeID` without time-range filtering
- Chronological order is not guaranteed by the current `EntityFilter`

---

## Indexes

| Collection | Field(s) | Type | Purpose |
|---|---|---|---|
| `dt_telemetry` | `agencyID, typeID` | persistent | `ListEntities` by telemetry type within an agency |
| `dt_telemetry` | `agencyID, deleted` | persistent | Exclude soft-deleted readings |
| `dt_telemetry` | `properties.entityID, properties.timestamp` | persistent | Time-range queries per producing entity |
| `dt_events` | `agencyID, typeID` | persistent | `ListEntities` by event type within an agency |
| `dt_events` | `agencyID, deleted` | persistent | Exclude soft-deleted events |
| `dt_events` | `properties.entityID, properties.timestamp` | persistent | Chronological event log per producing entity |

---

## Open Question — Telemetry Retention Policy

**Parked** (2026-04-27): `dt_telemetry` will grow at **Very High** write frequency
(see NFR §4 in requirements.md). Two options:

- **Option A**: Keep all readings indefinitely (simplest; may require archive strategy long-term)
- **Option B**: Apply an ArangoDB TTL index on `properties.timestamp` with a configurable window

Until a traffic profile is in scope, `dt_telemetry` is bootstrapped **without** a
TTL index — adding one later is non-destructive.

---

## gRPC Error Code Mapping

| Go error | gRPC code | When |
|---|---|---|
| `ErrImmutableType` | `codes.FailedPrecondition` | `UpdateEntity` on telemetry/event type |
| `ErrEntityNotFound` | `codes.NotFound` | `GetEntity` / `DeleteEntity` on unknown ID |
| `ErrInvalidEntity` | `codes.InvalidArgument` | Missing `agencyID` or `typeID` |

---

## Tests

| Test | File | Coverage |
|---|---|---|
| `TestEntityServer_CreateTelemetry_RoutesToCollection` | `internal/app/app_integration_test.go` | Write lands in `dt_telemetry`, not `dt_entities` |
| `TestEntityServer_UpdateTelemetry_Immutable` | `internal/app/app_integration_test.go` | `FailedPrecondition` returned |
| `TestEntityServer_CreateEvent_RoutesToCollection` | `internal/app/app_integration_test.go` | Write lands in `dt_events` |
| `TestEntityServer_UpdateEvent_Immutable` | `internal/app/app_integration_test.go` | `FailedPrecondition` returned |

Integration tests tagged `//go:build integration`; skip without `DT_ARANGO_ENDPOINT`.
