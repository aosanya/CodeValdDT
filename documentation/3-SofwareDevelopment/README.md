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

---

## MVP Status

| Task ID | Title | Status |
|---|---|---|
| MVP-DT-001 | Module Scaffolding | 📋 Not Started (unblocked — SHAREDLIB-010/011 done) |
| MVP-DT-002 | ArangoDB Backend | ⏸️ Blocked on MVP-DT-001 |
| MVP-DT-003 | gRPC Service Proto & Codegen | ⏸️ Blocked on MVP-DT-001 |
| MVP-DT-004 | gRPC Server Implementation | ⏸️ Blocked on MVP-DT-001/002/003 |
| MVP-DT-005 | CodeValdCross Registration | ⏸️ Blocked on MVP-DT-004 |
| MVP-DT-006 | Unit & Integration Tests | ⏸️ Blocked on MVP-DT-001/002/004 |

---

## Execution Order

```
SHAREDLIB-010 ✅ + SHAREDLIB-011 ✅ (already done — DT work unblocked)
      ↓
MVP-DT-001  ← Module scaffolding, go.mod, errors.go, models.go (DTDataManager + DTSchemaManager aliases)
      ↓
┌─────────────┬─────────────┐
MVP-DT-002          MVP-DT-003
(ArangoDB backend)  (proto + codegen)
└─────────────┴─────────────┘
      ↓
MVP-DT-004  ← gRPC server implementation; CreateEntity branches on StorageCollection for Cross topic
      ↓
┌─────────────┬─────────────┐
MVP-DT-005          MVP-DT-006
(Cross registration) (tests)
```

> **Parallel SharedLib track**: `SHAREDLIB-014` (`EntityFilter` time-range +
> default ordering for `dt_telemetry` / `dt_events`) is open in parallel; it
> is **not** a blocker for MVP-DT-001..MVP-DT-005, but FR-004 time-range
> telemetry queries cannot be exercised end-to-end until it lands.
