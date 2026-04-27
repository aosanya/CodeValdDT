# MVP Details — Per-Topic Task Specifications

## Purpose

Topic-grouped expansion of the [mvp.md](../mvp.md) task list. Each file in this
directory covers one **domain topic** (not one task ID) and contains the
acceptance criteria, design notes, and `### Tests` matrix that the QA layer
references.

> **Refactor rule (from `research.prompt.md`)**: Files in this directory must
> stay **≤ 500 lines**. This README must stay **≤ 300 lines**. If a topic file
> grows past 500, split by sub-topic; never split by task ID.

---

## Topic Index

| Topic File | Covers MVP Tasks | Status |
|---|---|---|
| _(none yet)_ | — | All tasks blocked on `SHAREDLIB-010`; topic files added as research closes the gaps below |

---

## Outstanding Research Gaps (drives future topic files)

These come from a `research.prompt.md` gap-analysis pass on 2026-04-27. Each
gap below must be resolved before the matching topic file can be written.

### Area 1 — Entity model
- [ ] No concrete entity types listed (only illustrative `Pump` / `Thermostat`)
- [ ] No write-time validation policy beyond "v1 trusts the caller"
- [ ] No definition of how the **active** schema version is selected at read
  time (highest version int? a separate pointer document?)

### Area 2 — Relationships
- [ ] No concrete relationship-name vocabulary
- [ ] `TraverseGraph` has no documented max-depth bound (AQL template uses an
  unbounded `@depth`)
- [ ] No index on `dt_relationships.name`, but `RelationshipFilter.Name` is a
  filterable field

### Area 3 — Telemetry
- ✅ Resolved 2026-04-27: telemetry writes are routed `CreateEntity` calls
  (`StorageCollection: "dt_telemetry"`, `Immutable: true`) — never a
  dedicated RPC. `CreateEntity` flow now branches on `StorageCollection` to
  pick the Cross topic. See [architecture-flows.md §9](../../2-SoftwareDesignAndArchitecture/architecture-flows.md)
- [ ] No telemetry-`typeID` vocabulary slot defined (e.g. `TemperatureReading`,
  `PressureReading`) — populate when an Agency domain is in scope
- [ ] Expected write frequency and retention policy unspecified — needed to
  size `dt_telemetry` and decide whether a TTL index is required
- [ ] `EntityFilter` in `CodeValdSharedLib/entitygraph` does not yet carry
  `TimeRangeFrom` / `TimeRangeTo` fields against `properties.timestamp`. Open
  a SharedLib gap (`SHAREDLIB-XXX_entityfilter_time_range`) before FR-004's
  time-range query becomes implementable

### Area 4 — Events
- ✅ Resolved 2026-04-27: events are routed `CreateEntity` calls
  (`StorageCollection: "dt_events"`, `Immutable: true`) — never a dedicated
  RPC; topic is `cross.dt.{agencyID}.event.recorded`
- [ ] No event-`typeID` vocabulary slot defined (e.g. `ValveOpened`,
  `AlarmRaised`) — populate when an Agency domain is in scope
- [ ] Per-source-entity ordering guarantee not contracted (the
  `properties.entityID, properties.timestamp` index makes timestamp-ordered
  reads cheap, but the API doesn't promise it)

### Area 5 — Integration
- [ ] No declared consumer of `cross.dt.{agencyID}.telemetry.recorded`
- [ ] No declared consumer of `TraverseGraph` results (UI? CodeValdComm?)
