# Problem Definition

## The Problem

[CodeValdCortex](../../../CodeValdCortex/README.md) is an enterprise multi-agent AI orchestration platform. Each Agency the platform serves operates over a domain of real-world entities — machines on a factory floor, vehicles in a fleet, people in a population, sensors on a network — and the **state, relationships, and history** of those entities is the source of truth that agents reason over.

Before CodeValdDT, the platform had no first-class place to store this state:

| Pain | Consequence |
|---|---|
| Entity state lived in ad-hoc per-service tables | Every consumer reinvented its own schema; nothing was shareable |
| No graph between entities | Questions like "which assets does pump P-12 connect to?" required custom SQL/AQL each time |
| No telemetry log | Time-series readings (temperature, pressure, status changes) had no agreed home |
| No event log per entity | Lifecycle history (`valve_opened`, `alarm_raised`) was scattered across logs and DBs |
| No standard schema language | Agency domains couldn't be exported to or migrated from external twin platforms |

---

## The Solution

CodeValdDT is a **Go gRPC microservice** that owns the Digital Twin layer:

- **Entities** — typed, versioned, scoped by `agencyID`, soft-deleted on removal
- **Relationships** — directed graph edges stored as an ArangoDB **edge collection**, traversable via AQL
- **Telemetry** — time-series readings, queryable by entity and time range
- **Events** — append-only per-entity event log
- **Schema** — DTDL v3 compatible (Azure Digital Twins migration path) — schema documents live in `dt_schemas` and are managed by `DTSchemaManager`
- **Pub/sub** — emits `cross.dt.{agencyID}.entity.created` and `cross.dt.{agencyID}.telemetry.recorded` via CodeValdCross
- **Cross registration** — registers as service `codevalddt` on `:50055` with a 20-second liveness heartbeat

---

## Why a Separate Service

| Concern | Reason it lives in CodeValdDT, not in another service |
|---|---|
| Graph traversal | Needs ArangoDB edge collections + named graphs — wrong fit for relational stores |
| Cross-Agency isolation | Every collection document carries `agencyID`; one shared database, scoped reads/writes |
| Schema portability | DTDL v3 is the lingua franca of digital-twin platforms — CodeValdDT keeps the platform exportable |
| Re-use | The same `entitygraph.DataManager` interface is used by CodeValdComm — defined in `CodeValdSharedLib` and aliased locally as `DTDataManager` |
