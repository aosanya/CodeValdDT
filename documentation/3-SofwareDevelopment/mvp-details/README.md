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
- [ ] Shape is unresolved: are telemetry writes a first-class
  `RecordTelemetry` RPC (per `requirements.md` FR-004) or routed
  `CreateEntity` calls with `TypeDefinition.StorageCollection: "dt_telemetry"`
  (per `architecture-flows.md` §9)? The proto in `architecture-service.md`
  currently has neither dedicated RPC
- [ ] No telemetry-name vocabulary, expected write frequency, or retention
  policy
- [ ] `cross.dt.{agencyID}.telemetry.recorded` publish trigger is undefined —
  if telemetry is a routed entity create, the `CreateEntity` flow must branch
  on `StorageCollection` to pick the right Cross topic

### Area 4 — Events
- [ ] No concrete event-name vocabulary
- [ ] Per-entity ordering guarantee not stated (timestamp-based is implied by
  the `(entityID, timestamp)` index but not contracted)

### Area 5 — Integration
- [ ] No declared consumer of `cross.dt.{agencyID}.telemetry.recorded`
- [ ] No declared consumer of `TraverseGraph` results (UI? CodeValdComm?)
