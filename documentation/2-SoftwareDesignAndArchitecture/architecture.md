# CodeValdDT — Architecture

This document is split into focused files — each under 300 lines.

| File | Contents |
|---|---|
| [architecture-interfaces.md](architecture-interfaces.md) | §1 Design decisions · §2 `DTManager` + `Backend` interfaces · §3 Data models |
| [architecture-service.md](architecture-service.md) | §4 Package structure · §5 gRPC service definition · §6 Cross registration |
| [architecture-storage.md](architecture-storage.md) | §7 ArangoDB schema — collections, document shapes, indexes |
| [architecture-flows.md](architecture-flows.md) | §8 Error types · §9 CreateEntity flow · §10 UpdateEntity (immutability) · §10a DeleteEntity (soft delete) · §11 SharedLib dependency |

---

> **Rule**: No documentation file in this directory may exceed 300 lines.
> Split into a new focused file when the limit is reached and add it to the table above.

