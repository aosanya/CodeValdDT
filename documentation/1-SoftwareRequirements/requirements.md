# CodeValdDT — Requirements

## 1. Purpose

CodeValdDT is a **Go gRPC microservice** that manages the Digital Twin layer
of the CodeVald platform.

A Digital Twin is a live, graph-structured model of an Agency's real-world
entities — machines, people, locations, industrial assets, individuals in a
population, or any typed object the agency cares about. CodeValdDT stores
those entities, the graph of relationships between them, telemetry readings,
and events.

---

## 2. Scope

### In Scope
- Full entity lifecycle (create, read, update, delete, list)
- Graph of relationships between entities (ArangoDB edge collection)
- Graph traversal by depth and direction
- Telemetry recording and historical querying per entity
- Event log recording and reading per entity
- Pub/sub event publishing via CodeValdCross
- Registration and liveness heartbeat with CodeValdCross
- DTDL v3 compatible schema model (Azure Digital Twins migration path)

### Out of Scope
- Entity type schema definition and enforcement — schema lives in CodeValdAgency; v1 trusts the caller
- Task or work item management (CodeValdWork)
- Git artifact management (CodeValdGit)
- AI agent orchestration (CodeValdAI)
- Real-time streaming subscriptions (deferred)

---

## 3. Functional Requirements

### FR-001: Entity Management
- CodeValdDT must support creating, reading, updating, deleting, and listing
  entity instances scoped by `agencyID`
- Each entity has a `typeID` (name of an `EntityTypeDefinition` on the Agency),
  a free-form `properties` map, and timestamps
- Entity creation is allowed regardless of Agency publication status in v1

### FR-002: Graph Relationships
- Relationships between entities must be stored as **ArangoDB edge collection**
  documents with `_from` and `_to` pointing to entity document handles
- The `relationships` collection must be created as `CollectionTypeEdge`
- Relationships have a `name` (e.g. `connects_to`) and optional properties

### FR-003: Graph Traversal
- The service must support traversing the graph from a starting entity with
  configurable depth and direction (inbound, outbound, any)
- Traversal is implemented using ArangoDB AQL graph queries on a named graph
  that includes the `relationships` edge collection

### FR-004: Telemetry
- The service must support recording telemetry readings against an entity
  (name, value, timestamp)
- Historical telemetry must be queryable by entity, optionally filtered by
  time range
- After every successful `RecordTelemetry`, publish
  `cross.dt.{agencyID}.telemetry.recorded` via CodeValdCross

### FR-005: Events
- The service must support appending events to an entity's event log
  (name, payload, timestamp)
- Events must be listable per entity in chronological order

### FR-006: Pub/Sub (v1)
- After every successful `CreateEntity`, publish
  `cross.dt.{agencyID}.entity.created`
- After every successful `RecordTelemetry`, publish
  `cross.dt.{agencyID}.telemetry.recorded`

### FR-007: CodeValdCross Registration
- On startup, register with CodeValdCross using service name `codevalddt`
  and address `:50055`
- Repeat registration every 20 seconds as a liveness heartbeat

### FR-008: DTDL v3 Compatibility
- The data model must be exportable to Azure Digital Twins DTDL v3 format
- Entity types → DTDL `Interface`
- Properties → DTDL `Property`
- Telemetry → DTDL `Telemetry`
- Relationships → DTDL `Relationship` (stored as ArangoDB edge documents)
- Events → DTDL `Telemetry` with structured payload

---

## 4. Non-Functional Requirements

| NFR | Requirement |
|---|---|
| Language | Go 1.21+ |
| API | gRPC + protobuf |
| Storage | ArangoDB — single shared database (`DT_ARANGO_DATABASE` env var); collections scoped by `agencyID` field |
| Schema standard | DTDL v3 compatible |
| Context propagation | All exported methods take `context.Context` as first arg |
| Godoc | All exported symbols must have godoc comments |
| File size | Max 500 lines per file |
| Function size | Max 50 lines per function |
| Test coverage | All business logic covered with `-race` tests |
| No hardcoded storage | `Backend` interface injected via constructor |
| No cross-service imports | All cross-service calls go through gRPC |

---

## 5. Open Questions (v1)

| Question | Decision |
|---|---|
| Schema enforcement at entity creation | Deferred — v1 trusts the caller |
| Live telemetry streaming (gRPC server-stream) | Deferred |
| Entity deletion cascade to relationships/telemetry/events | **Resolved — no cascade in v1.** `DeleteEntity` soft-deletes only the entity; its relationships, telemetry, and events are retained as-is. Orphan cleanup deferred to v2. |
| Soft delete vs. hard delete for entities | **Resolved — soft delete.** `DeleteEntity` sets `deleted: true` and `deletedAt` on the document. Hard delete is not exposed in v1. |

### Out of Scope
- Remote Git hosting (no GitHub/GitLab push/pull — local repos only, for now)
- Authentication / access control (handled by CodeValdCortex's policy layer)
- Pull request UI (merge is programmatic, not UI-driven)

---

## 3. Replaces

CodeValdGit **fully replaces** `internal/git/` in CodeValdCortex:

| Replaced package | Reason for replacement |
|---|---|
| `internal/git/ops/` | Custom SHA-1 Git object engine over ArangoDB → replaced by go-git |
| `internal/git/storage/` | ArangoDB `git_objects`, `git_refs`, `repositories` collections → replaced by real `.git` on disk |
| `internal/git/fileindex/` | ArangoDB-backed file index → replaced by go-git tree walking |
| `internal/git/models/` | Custom GitObject, GitTree, GitCommit structs → replaced by go-git types |

> **No migration needed.** The ArangoDB Git collections (`git_objects`, `git_refs`, `repositories`) will be dropped entirely.

---

## 4. Functional Requirements

### FR-001: Repository Per Agency
- Each Agency in CodeValdCortex has exactly **one Git repository**
- Repository identity is the **Agency ID** (matches the existing database-per-agency isolation model)
- Repos must be initializable, openable, and deletable by Agency ID

### FR-002: Any File Type
- The library must store **any file type** without restriction
- Text files (`.go`, `.md`, `.yaml`, `.json`, etc.) should be stored as-is and produce meaningful diffs
- Binary files are stored as blobs

### FR-003: Branch-Per-Task Workflow
- Agents **must not commit directly to `main`**
- Every write operation happens on a **task branch**: `task/{task-id}`
- The library must support:
  - Creating a task branch from `main`
  - Committing files to a task branch
  - Auto-merging a task branch to `main` on task completion
  - Deleting the task branch after merge

### FR-004: Commit Attribution
- Every commit must record the **author** (agent ID or human user) and a **message**
- Commit messages should be structured and machine-readable (e.g., include task ID)

### FR-005: File Operations
- Read file content at HEAD or any commit SHA
- List directory contents (tree walking)
- Get commit history for a file or path
- Diff between two commits or between a branch and `main`

### FR-006: Merge Conflict Resolution
- When `MergeBranch` is called and `main` has advanced since the task branch was created (fast-forward not possible), the library **must first attempt an auto-rebase** of the task branch onto the current `main`
- If the rebase succeeds (no file-level conflicts), the fast-forward merge proceeds automatically
- If the rebase encounters a content conflict, the library **must return a structured error** to the caller (CodeValdCortex) containing:
  - The conflicting file path(s)
  - The nature of the conflict
- The caller is responsible for routing the conflict back to the agent for resolution
- The task branch must be left in a clean state (rebase aborted) on conflict so the agent can retry

> **go-git constraint**: `Repository.Merge()` only supports `FastForwardMerge` strategy (added v5.12.0). Three-way merges and rebase are not natively supported in go-git. The rebase step must be implemented manually by walking commits on the task branch and cherry-picking them onto `main`.

### FR-007: Repository Archiving
- When an Agency is deleted, its Git repository **must not be hard-deleted immediately**
- `DeleteRepo(agencyID)` must **archive** the repository by moving it to a configurable archive path: `{archive_base_path}/{agency-id}/`
- The archived repo is a complete, valid `.git` repository — it can be inspected or restored at any time
- A separate `PurgeRepo(agencyID)` call performs the actual hard delete (`os.RemoveAll`) for operators who explicitly want permanent removal
- The `RepoManager` must be configured with both a `base_path` (live repos) and an `archive_base_path` (archived repos)

### FR-008: History and Diff Read Access (UI)
- The library must support reading historical state for the CodeValdCortex UI at launch
- Required operations:
  - **File content at any ref**: `ReadFile(ctx, ref, path)` where `ref` is a branch name, tag, or commit SHA
  - **Directory listing at any ref**: `ListDirectory(ctx, ref, path)` — enables a file browser at any point in history
  - **File commit history**: `Log(ctx, ref, path)` — returns ordered list of commits that touched a given path
  - **Diff between two refs**: `Diff(ctx, fromRef, toRef)` — returns per-file changes between any two commits or branches
- All read operations must be non-mutating and safe to call concurrently
- These are already present in the draft `Repo` interface in the architecture doc

---

## 5. Non-Functional Requirements

### NFR-001: Embeddable Library
- Must be importable as a standard Go module
- No long-running daemon or sidecar process required
- Caller (CodeValdCortex) controls concurrency
- Storage backend is injected by the caller via `storage.Storer` — supports filesystem and ArangoDB out of the box

### NFR-002: No External Git Binary
- Must use go-git's pure-Go implementation
- No dependency on the `git` CLI binary at runtime

---

## 6. Open Questions (Research Gaps)

| # | Question | Impact |
|---|---|---|
| ~~OQ-001~~ | ~~Where are Git repos stored? Filesystem path, shared PVC, or in-memory?~~ | ✅ **Resolved** — pluggable via `storage.Storer`; filesystem and ArangoDB are both supported backends; caller injects the implementation |
| ~~OQ-002~~ | ~~What happens when an auto-merge fails due to a conflict?~~ | ✅ **Resolved** — see FR-006: auto-rebase then surface conflict error to caller |
| ~~OQ-003~~ | ~~What happens to the Git repo when an Agency is deleted?~~ | ✅ **Resolved** — see FR-007: `DeleteRepo` archives to `archive_base_path`; `PurgeRepo` hard-deletes |
| ~~OQ-004~~ | ~~Should the library support read access to historical commits from the CodeValdCortex UI?~~ | ✅ **Resolved** — yes, at launch; see FR-008 |
| ~~OQ-005~~ | ~~Are there any file size limits or quotas per repo?~~ | ✅ **Resolved** — no limits enforced; library imposes no file size or repo size constraints |
