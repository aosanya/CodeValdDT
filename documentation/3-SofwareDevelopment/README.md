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
| MVP-DT-001 | Module Scaffolding | ⏸️ Blocked on SHAREDLIB-010 |
| MVP-DT-002 | ArangoDB Backend | ⏸️ Blocked |
| MVP-DT-003 | gRPC Service Proto & Codegen | ⏸️ Blocked |
| MVP-DT-004 | gRPC Server Implementation | ⏸️ Blocked |
| MVP-DT-005 | CodeValdCross Registration | ⏸️ Blocked |
| MVP-DT-006 | Unit & Integration Tests | ⏸️ Blocked |

---

## Execution Order

```
SHAREDLIB-010 (unblocks all DT work)
      ↓
MVP-DT-001  ← Module scaffolding, go.mod, errors.go, models.go
      ↓
┌─────────────┬─────────────┐
MVP-DT-002          MVP-DT-003
(ArangoDB backend)  (proto + codegen)
└─────────────┴─────────────┘
      ↓
MVP-DT-004  ← gRPC server implementation
      ↓
┌─────────────┬─────────────┐
MVP-DT-005          MVP-DT-006
(Cross registration) (tests)
```
