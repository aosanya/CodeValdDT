# Stakeholders

## Primary Consumers

| Stakeholder | Role | How They Use CodeValdDT |
|---|---|---|
| **CodeValdCross** | Service registry + pub/sub bus | Receives `Register` heartbeat every 20 s; routes `cross.dt.{agencyID}.entity.created` and `cross.dt.{agencyID}.telemetry.recorded` events to subscribers |
| **CodeValdCortex** (and any service that proxies through Cross) | Calls `DTService` over gRPC for entity / relationship / telemetry / event operations |
| **CodeValdAgency** | Owner of the entity-type schema (DTDL v3 `Interface` definitions). CodeValdDT reads the schema via `DTSchemaManager.GetSchema` to resolve `TypeDefinition.StorageCollection` and `TypeDefinition.Immutable` |

---

## DTService Integration Points

CodeValdDT is invoked at these points by upstream services:

| Event | DTService RPC |
|---|---|
| Domain object created in an Agency | `CreateEntity` |
| Read latest entity state | `GetEntity` |
| Patch entity properties | `UpdateEntity` (rejected if `TypeDefinition.Immutable`) |
| Remove an entity | `DeleteEntity` (soft delete) |
| Browse entities of a given type | `ListEntities` |
| Add a graph edge between two entities | `CreateRelationship` |
| Read / remove a single edge | `GetRelationship`, `DeleteRelationship` |
| List edges by `FromID` / `ToID` / `Name` | `ListRelationships` |
| Walk the entity graph from a starting node | `TraverseGraph` (returns vertices + edges in one round-trip) |

Telemetry and event writes route through the schema-driven `CreateEntity` flow when `TypeDefinition.StorageCollection` is set to `dt_telemetry` or `dt_events` (see [architecture-flows.md §9](../../2-SoftwareDesignAndArchitecture/architecture-flows.md)).

---

## Secondary Stakeholders

| Stakeholder | Interest |
|---|---|
| **CodeValdComm** | Uses the same `entitygraph.DataManager` and `entitygraph.SchemaManager` interfaces from `CodeValdSharedLib` for its own entity-graph store — so the contract must stay stable across both services |
| **Platform operators** | Need to provision the shared ArangoDB database (`DT_ARANGO_DATABASE`); CodeValdDT creates collections and indexes idempotently on startup |
| **AI agents (indirect)** | Read and update twin state via CodeValdCross-routed `DTService` calls; subscribe to `entity.created` / `telemetry.recorded` to react to world changes |
| **End users (indirect)** | View twins, graphs, telemetry, and event history through the platform UI — powered by the read-side RPCs (`GetEntity`, `ListEntities`, `ListRelationships`, `TraverseGraph`) |

---

## Service Maintainers

CodeValdDT is maintained as part of the **CodeVald** platform alongside CodeValdCross, CodeValdAgency, CodeValdComm, CodeValdWork, and CodeValdGit. Development follows:
- Trunk-based development with short-lived feature branches (`feature/DT-XXX_description`)
- All shared infrastructure (registration, pub/sub stubs, ArangoDB connect, server bootstrap, entity-graph interfaces) lives in [`CodeValdSharedLib`](../../../../CodeValdSharedLib/) — CodeValdDT retains only domain logic, gRPC handlers, and storage collection bootstrap
- DTDL v3 as the schema standard — keeps an Azure Digital Twins migration path open
