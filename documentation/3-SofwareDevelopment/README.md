# 3 — Software Development

## Overview

This section tracks the development plan, MVP task breakdown, and implementation
details for CodeValdDT.

---

## Index

| Document | Description |
|---|---|
| [mvp.md](mvp.md) | Active MVP scope, task list, and completion status |
| [mvp_done.md](mvp_done.md) | Completed tasks with completion dates and branches |
| [mvp-details/](mvp-details/README.md) | Per-topic task specifications grouped by domain |

---

## MVP Status

| Task ID | Title | Status |
|---|---|---|
| MVP-DT-001 | Module Scaffolding | ✅ Done (2026-04-27) |
| MVP-DT-002 | ArangoDB Backend | ✅ Done (2026-04-27) |
| MVP-DT-003 | gRPC Service Proto & Codegen | ❌ Withdrawn (2026-04-27) |
| MVP-DT-004 | gRPC Server Implementation | ❌ Withdrawn (2026-04-27) |
| MVP-DT-005 | CodeValdCross Registration | ✅ Done (2026-04-27) |
| MVP-DT-006 | Unit & Integration Tests | ✅ Done (2026-04-28) |

All MVP rows are closed out. See [mvp_done.md](mvp_done.md) for landing
details and [mvp.md](mvp.md) for the withdrawal rationale on DT-003 / DT-004.

---

## Next Candidates

The MVP scope is closed. Open spec items that could seed a new task batch:

- **FR-008** — DTDL v3 export (no implementation yet)
- **`SHAREDLIB-014`** — `EntityFilter` time-range + default ordering for
  `dt_telemetry` / `dt_events`; required before FR-004 time-range telemetry
  queries are implementable end-to-end
- **Parked open questions** in [requirements.md §5](../1-SoftwareRequirements/requirements.md) —
  `dt_telemetry` retention TTL, `TraverseGraph` max-depth ceiling
