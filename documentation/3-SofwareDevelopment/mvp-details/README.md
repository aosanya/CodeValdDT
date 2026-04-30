# MVP Details — Per-Topic Task Specifications

## Purpose

Topic-grouped expansion of the [mvp.md](../mvp.md) task list. Each file in this
directory covers one **domain topic** (not one task ID) and contains the
acceptance criteria, design notes, and `### Tests` matrix that the QA layer
references.

> **Refactor rule**: Files in this directory must stay **≤ 500 lines**. This
> README must stay **≤ 300 lines**. If a topic file grows past 500, split by
> sub-topic; never split by task ID.

---

## Topic Index

| Topic File | Covers | Status |
|---|---|---|
| [entity-lifecycle.md](entity-lifecycle.md) | FR-001: Entity CRUD, storage routing, immutability guard, soft delete | ✅ Done |
| [relationships-and-graph.md](relationships-and-graph.md) | FR-002: Edge collection; FR-003: Graph traversal (`dt_graph`, AQL) | ✅ Done |
| [telemetry-and-events.md](telemetry-and-events.md) | FR-004: Telemetry; FR-005: Events — routed `CreateEntity`, SHAREDLIB-014 dependency | ✅ Partial (time-range blocked on SHAREDLIB-014) |
| [integration.md](integration.md) | FR-006: Pub/sub topics; FR-007: Cross registration heartbeat; SharedLib wiring | ✅ Done |
| [dtdl-export.md](dtdl-export.md) | FR-008: DTDL v3 schema export — next candidate | 📋 Not Started |

---

## MVP Task → Topic File Map

| MVP Task | Topic File |
|---|---|
| MVP-DT-001 (Module Scaffolding) | [entity-lifecycle.md](entity-lifecycle.md) |
| MVP-DT-002 (ArangoDB Backend) | [entity-lifecycle.md](entity-lifecycle.md), [relationships-and-graph.md](relationships-and-graph.md), [telemetry-and-events.md](telemetry-and-events.md) |
| ~~MVP-DT-003~~ (Withdrawn) | — |
| ~~MVP-DT-004~~ (Withdrawn) | — |
| MVP-DT-005 (Cross Registration) | [integration.md](integration.md) |
| MVP-DT-006 (Unit & Integration Tests) | All topic files |
| DT-007 (DTDL v3 Export — next) | [dtdl-export.md](dtdl-export.md) |

---

## Research Status (last pass 2026-04-27)

### Area 1 — Entity model
- ✅ Vocabulary slot — exists in [`types.TypeDefinition`](file:///workspaces/CodeVald-AIProject/CodeValdSharedLib/types/schema.go) (`Name`, `Properties`, `Relationships`, `StorageCollection`, `Immutable`, …). Concrete type names are **per-agency runtime data** populated through `DTSchemaManager.SetSchema`; CodeValdDT itself ships no built-in vocabulary.
- ✅ Active schema version — selected by `Schema.Active == true` (single-published-version-per-agency invariant). See [`types.Schema.Active`](file:///workspaces/CodeVald-AIProject/CodeValdSharedLib/types/schema.go).
- ✅ Validation policy — v1 trusts the caller (per FR; deferred to v2).

### Area 2 — Relationships
- ✅ Vocabulary slot — exists in [`types.RelationshipDefinition`](file:///workspaces/CodeVald-AIProject/CodeValdSharedLib/types/schema.go) (`Name`, `ToType`, `ToMany`, `Required`, …). Concrete relationship names live in the agency's `Schema`, not in CodeValdDT.
- ✅ `dt_relationships.name` index — added to [architecture-storage.md](../../2-SoftwareDesignAndArchitecture/architecture-storage.md) on the same pass.
- 🅿️ `TraverseGraph` max-depth ceiling — **Parked**; tracked in [requirements.md §5](../../1-SoftwareRequirements/requirements.md). MVP-DT-004 passes the caller's `Depth` through unchanged until the question is revisited.

### Area 3 — Telemetry
- ✅ Shape — telemetry writes are routed `CreateEntity` calls (`StorageCollection: "dt_telemetry"`, `Immutable: true`). No `RecordTelemetry` RPC. See [architecture-flows.md §9](../../2-SoftwareDesignAndArchitecture/architecture-flows.md).
- ✅ Vocabulary slot — `TypeDefinition` covers it; concrete telemetry type names are per-agency runtime data.
- ✅ Write frequency — **Very High** (recorded as an NFR in [requirements.md §4](../../1-SoftwareRequirements/requirements.md)).
- ✅ Default ordering — `properties.timestamp ASC` when `StorageCollection ∈ {dt_telemetry, dt_events}`. Documented in [architecture-storage.md](../../2-SoftwareDesignAndArchitecture/architecture-storage.md) and [architecture-flows.md](../../2-SoftwareDesignAndArchitecture/architecture-flows.md). Enforced by `EntityFilter` contract in `SHAREDLIB-014`.
- ✅ SharedLib gap filed — `SHAREDLIB-014` (`EntityFilter` time-range + default ordering). See [CodeValdSharedLib mvp.md](../../../../CodeValdSharedLib/documentation/3-SofwareDevelopment/mvp.md).
- 🅿️ **Retention policy** — **Parked**; tracked in [requirements.md §5](../../1-SoftwareRequirements/requirements.md). MVP-DT-002 bootstraps `dt_telemetry` without a TTL index; adding one later is non-destructive.

### Area 4 — Events
- ✅ Shape — events are routed `CreateEntity` calls (`StorageCollection: "dt_events"`, `Immutable: true`). No `RecordEvent` RPC.
- ✅ Vocabulary slot — `TypeDefinition` covers it; concrete event names are per-agency runtime data.
- ✅ Per-source-entity chronological ordering — guaranteed by the default `properties.timestamp ASC` ordering rule (above).

### Area 5 — Integration
- [ ] No declared consumer of `cross.dt.{agencyID}.telemetry.recorded` or `cross.dt.{agencyID}.event.recorded`. The topics exist on the Cross bus; subscriber list is unknown.
- [ ] No declared consumer of `TraverseGraph` results — UI? CodeValdComm? — needed to size the response payload and decide whether pagination is required.
