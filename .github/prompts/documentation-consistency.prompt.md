---
agent: agent
---

# Documentation Consistency Checker

## Purpose
Systematic consistency checks for **CodeValdDT** — one question at a time.

---

## Current Technology Stack (Reference)

```yaml
Service:
  Language: Go 1.21+
  Module: github.com/aosanya/CodeValdDT
  gRPC: google.golang.org/grpc
  Storage: ArangoDB — one DB per agency
  Port: :50055
  ServiceName: codevalddt

Key interfaces:
  - DTManager: CreateEntity, GetEntity, UpdateEntity, DeleteEntity,
               ListEntities, CreateRelationship, DeleteRelationship,
               TraverseGraph, RecordTelemetry, QueryTelemetry,
               RecordEvent, ListEvents
  - Backend: same operations

ArangoDB collections (per agency DB):
  - entities        (document)
  - relationships   (edge collection)
  - telemetry       (document)
  - events          (document)

Cross-service events (v1):
  Produces:
    - cross.dt.{agencyID}.entity.created
    - cross.dt.{agencyID}.telemetry.recorded
  Consumes: (none)

Documentation structure:
  1-SoftwareRequirements:  requirements.md
  2-SoftwareDesignAndArchitecture: architecture.md, dtdl/
  3-SofwareDevelopment: mvp.md, mvp-details/
  4-QA: README.md
```

---

## Consistency Check Areas (priority order)

1. **Interface contract** — does `architecture.md` match actual Go interfaces?
2. **Data models** — do `models.go` field names match what's documented?
3. **Error types** — does `errors.go` match the error table in `architecture.md`?
4. **Registration payload** — does the `RegisterRequest` in code match
   `architecture.md` Section 5?
5. **Edge collection** — is `relationships` created as an edge collection in
   `storage/arangodb/storage.go`?
6. **Pub/sub topics** — do topic strings in code match the constants and docs?
7. **mvp.md task status** — are completed tasks marked ✅?
8. **File size limits** — any files over 500 lines?

```bash
wc -l **/*.go
wc -l documentation/**/*.md
```

---

## Stop Conditions

- ❌ Any file in `documentation/` over 400 lines without a subfolder →
  **must refactor first**
- ❌ Architecture doc references interfaces that don't exist in code →
  **must update**
- ❌ `mvp.md` tasks marked 🔲 that are already implemented → **must update**
- ❌ `relationships` collection created as document type → **must fix**
