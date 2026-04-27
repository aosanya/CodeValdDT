# 2 — Software Design & Architecture

## Overview

This section captures the **how** — design decisions, data models, component architecture, and technical constraints for CodeValdDT.

---

## Index

| Document | Description |
|---|---|
| [architecture.md](architecture.md) | Top-level index over the four focused architecture files below |
| [architecture-interfaces.md](architecture-interfaces.md) | Core design decisions; `DTDataManager` + `DTSchemaManager` interfaces; `Entity` / `Relationship` / `DTSchema` / `TypeDefinition` data models |
| [architecture-service.md](architecture-service.md) | Package structure; `DTService` gRPC definition; CodeValdCross registration heartbeat |
| [architecture-storage.md](architecture-storage.md) | ArangoDB schema — collections (`dt_schemas`, `dt_entities`, `dt_relationships`, `dt_telemetry`, `dt_events`), document shapes, indexes, named graph |
| [architecture-flows.md](architecture-flows.md) | Error-to-gRPC mapping; CreateEntity flow; UpdateEntity immutability guard; DeleteEntity soft-delete flow; SharedLib dependency |
| [dtdl/](dtdl/README.md) | DTDL v3 reference — Interface, contents, schemas, additional concerns |

---

## Key Design Decisions at a Glance

| Decision | Choice | Rationale |
|---|---|---|
| Business-logic entry point | `DTDataManager = entitygraph.DataManager` (SharedLib) | Shared with CodeValdComm; gRPC handlers stay thin |
| Schema management | `DTSchemaManager = entitygraph.SchemaManager` | `dt_schemas` is owned by the schema manager; data manager has no schema methods |
| Storage | ArangoDB — single shared database (`DT_ARANGO_DATABASE`); collections scoped by `agencyID` field | Matches platform env-var convention; one DB to operate |
| Graph storage | ArangoDB **edge collection** `dt_relationships` + named graph `dt_graph` | Native AQL graph traversal; no separate graph engine |
| Entity deletion | Soft delete (`deleted` + `deletedAt`) — no cascade | Preserves telemetry/event history; orphan cleanup deferred to v2 |
| Type immutability | Driven by `TypeDefinition.Immutable` — `UpdateEntity` returns `ErrImmutableType` | Telemetry/event records can't be edited after write |
| Storage routing | Driven by `TypeDefinition.StorageCollection` (`dt_entities` default; `dt_telemetry` / `dt_events` for routed types) | No hard-coded type checks in the manager |
| Cross integration | Registration heartbeat every 20 s; pub/sub via SharedLib `gen/go/codevaldcross/v1` `Publish` stub | Platform standard; agencyID-scoped topics |
| No cross-service Go imports | All cross-service calls go through gRPC | Stable versioned contracts; independent deployment |
