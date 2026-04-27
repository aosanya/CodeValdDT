# 4 — QA

## Overview

This section covers testing strategy, acceptance criteria, and quality assurance for CodeValdDT.

---

## Index

| Document | Description |
|---|---|
| _(none yet)_ | Test plans and QA artifacts will be added as tasks are implemented |

---

## Testing Standards

All contributions must satisfy:

| Check | Command | Requirement |
|---|---|---|
| Build | `go build ./...` | Must succeed — no compilation errors |
| Unit tests | `go test -v -race ./...` | All tests green; no data races |
| Static analysis | `go vet ./...` | 0 issues |
| Linting | `golangci-lint run ./...` | Must pass |
| Coverage | `go test -coverprofile=coverage.out ./...` | Target ≥ 80% on exported functions |

---

## Test Structure Convention

Tests live alongside source files using Go's standard `_test.go` convention:

```
internal/
  manager/
    manager_test.go   ← DTDataManager unit tests (mock DTSchemaManager + Cross publisher)
  server/
    server_test.go    ← gRPC handler tests (table-driven; mock DTDataManager)
  registrar/
    registrar_test.go ← Cross heartbeat loop tests
storage/
  arangodb/
    storage_test.go   ← ArangoDB DTSchemaManager + collection/index bootstrap tests
```

Integration tests that require external services (ArangoDB, CodeValdCross) must be tagged `//go:build integration` and use `t.Skip()` when `DT_ARANGO_URL` / `CROSS_ADDR` is not set.

---

## Acceptance Criteria per Task

See the `### Tests` section of each task file in [../3-SofwareDevelopment/mvp-details/](../3-SofwareDevelopment/mvp-details/README.md) for the full test matrix per MVP task.
