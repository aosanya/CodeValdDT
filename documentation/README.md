# CodeValdDT — Documentation

## Overview

**CodeValdDT** is a Go gRPC microservice that manages the **Digital Twin** layer
of the CodeVald platform.

A Digital Twin is a live, graph-structured model of an Agency's real-world
entities — machines, people, locations, networks, or any typed object the agency
cares about. CodeValdDT stores entities, the graph of relationships between them
(as ArangoDB edge collections), telemetry readings, and events. It is scoped
by `agencyID` like all other CodeVald services.

The entity type schema (DTDL v3 Interface) is defined on the Agency in
CodeValdAgency and is compatible with Azure Digital Twins for future migration.

---

## Documentation Index

| Document | Description |
|---|---|
| [1-SoftwareRequirements/requirements.md](1-SoftwareRequirements/requirements.md) | Functional requirements, scope, NFR |
| [2-SoftwareDesignAndArchitecture/architecture.md](2-SoftwareDesignAndArchitecture/architecture.md) | Design decisions, data model, gRPC service, ArangoDB schema |
| [2-SoftwareDesignAndArchitecture/dtdl/](2-SoftwareDesignAndArchitecture/dtdl/README.md) | DTDL v3 reference — Interface, contents, schemas, additional concerns |
| [3-SofwareDevelopment/mvp.md](3-SofwareDevelopment/mvp.md) | MVP task list and status |
| [3-SofwareDevelopment/mvp-details/](3-SofwareDevelopment/mvp-details/) | Per-topic task specifications |
| [4-QA/README.md](4-QA/README.md) | Testing strategy and acceptance criteria |

---

## Quick Summary

- **Language**: Go 1.21+
- **API**: gRPC + protobuf (`DTService`), port `:50055`
- **Storage**: ArangoDB — one database per agency
- **Collections**: `entities` (document), `relationships` (edge), `telemetry` (document), `events` (document)
- **Graph traversal**: ArangoDB native graph via edge collection
- **Schema standard**: DTDL v3 (Azure Digital Twins compatible)
- **Pub/sub (v1)**: produces `cross.dt.{agencyID}.entity.created` and `cross.dt.{agencyID}.telemetry.recorded`
- **Registration**: CodeValdCross heartbeat every 20 s
