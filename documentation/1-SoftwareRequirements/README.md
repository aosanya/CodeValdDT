# 1 — Software Requirements

## Overview

This section captures the **what** and **why** for CodeValdDT — without prescribing how.

---

## Index

| Document | Description |
|---|---|
| [requirements.md](requirements.md) | Functional requirements (FR-001–FR-008), non-functional requirements, scope, and resolved open questions |
| [introduction/problem-definition.md](introduction/problem-definition.md) | Problem statement and motivation for the service |
| [introduction/high-level-features.md](introduction/high-level-features.md) | High-level capability summary |
| [introduction/stakeholders.md](introduction/stakeholders.md) | Consumers and stakeholders of the service |

---

## Summary

CodeValdDT is a **Go gRPC microservice** that manages the Digital Twin layer of the CodeVald platform — a live, graph-structured model of an Agency's real-world entities (machines, people, locations, industrial assets, …) backed by ArangoDB. Schemas are DTDL v3 compatible for future migration to Azure Digital Twins.

### Core Requirements at a Glance

| FR | Requirement |
|---|---|
| FR-001 | Entity lifecycle (create / read / update / delete / list), scoped by `agencyID` |
| FR-002 | Graph relationships stored in an ArangoDB **edge collection** (`dt_relationships`) |
| FR-003 | Graph traversal by depth and direction via AQL on the `dt_graph` named graph |
| FR-004 | Telemetry readings stored as `Entity` instances routed to `dt_telemetry` via `TypeDefinition.StorageCollection`; queryable by source entity and time range |
| FR-005 | Events stored as `Entity` instances routed to `dt_events`; listable by source entity in chronological order |
| FR-006 | Pub/sub via CodeValdCross — `entity.created`, `telemetry.recorded`, `event.recorded` topics, chosen from the resolved `StorageCollection` |
| FR-007 | CodeValdCross registration heartbeat every 20 s as service `codevalddt` on `:50055` |
| FR-008 | DTDL v3 compatible data model — exportable to Azure Digital Twins |
