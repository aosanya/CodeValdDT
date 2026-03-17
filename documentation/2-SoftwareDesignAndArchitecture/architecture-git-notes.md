# CodeValdDT — Architecture: CodeValdGit Reference Notes

> Part of [architecture.md](architecture.md)
>
> **Note**: This section contains CodeValdGit architecture notes that are
> referenced from the CodeValdDT architecture context. The canonical source is
> `CodeValdGit/documentation/2-SoftwareDesignAndArchitecture/architecture.md`.

## CodeValdGit Storage Backend Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Repo granularity | 1 repo per Agency | Mirrors CodeValdCortex's database-per-agency isolation |
| Agent write policy | Always on a branch, never `main` | Prevents concurrent agent writes from corrupting shared history |
| Branch naming | `task/{task-id}` | Short-lived, traceable back to CodeValdCortex task records |
| Merge strategy | Auto-merge on task completion | No human approval gate for now; policy layer can extend this later |
| Storage backend | Pluggable via `storage.Storer` interface | go-git's open/closed design; caller injects the storer — filesystem and ArangoDB are both valid implementations |
| Worktree filesystem | Pluggable via `billy.Filesystem` interface | go-git separates object storage from the working tree; both are independently injectable |

---

## go-git Pluggable Interfaces

go-git separates storage into two injectable interfaces:

| Interface | Package | Purpose |
|---|---|---|
| `storage.Storer` | `github.com/go-git/go-git/v5/storage` | Git objects, refs, index, config |
| `billy.Filesystem` | `github.com/go-git/go-billy/v5` | Working tree (checked-out files) |

## CodeValdGit `Backend` Interface

```go
// Backend abstracts storage-specific repo lifecycle.
// Implemented by storage/filesystem and storage/arangodb.
type Backend interface {
    InitRepo(ctx context.Context, agencyID string) error
    OpenStorer(ctx context.Context, agencyID string) (storage.Storer, billy.Filesystem, error)
    DeleteRepo(ctx context.Context, agencyID string) error
    PurgeRepo(ctx context.Context, agencyID string) error
}
```

## Storage Backends

### Filesystem Backend (`storage/filesystem/`)

| Operation | Implementation |
|---|---|
| `InitRepo` | `git.PlainInit` on disk; empty commit on `main` |
| `DeleteRepo` | `os.Rename` to `{archive_path}/{agency-id}/` (non-destructive) |
| `PurgeRepo` | `os.RemoveAll` of archive directory |
| `OpenStorer` | `filesystem.NewStorage` + `osfs.New` |

### ArangoDB Backend (`storage/arangodb/`)

| Operation | Implementation |
|---|---|
| `InitRepo` | Insert seed documents into `git_objects`, `git_refs`, `git_config`, `git_index` |
| `DeleteRepo` | Set `deleted: true` flag on all agency documents (non-destructive; auditable) |
| `PurgeRepo` | Delete all documents where `agencyID == target` from all four collections |
| `OpenStorer` | `arango.NewStorage(db, agencyID)` + `memfs.New()` (or `osfs` for a durable worktree) |

| Collection | Contents |
|---|---|
| `git_objects` | Encoded Git objects (blobs, trees, commits, tags) keyed by SHA |
| `git_refs` | Branch and tag references |
| `git_index` | Staging area index |
| `git_config` | Per-repo Git config |

## CodeValdGit Package Layout

```
github.com/aosanya/CodeValdGit/
├── codevaldgit.go          # RepoManager + Repo + Backend interfaces
├── types.go                # FileEntry, Commit, FileDiff, AuthorInfo, ErrMergeConflict
├── errors.go               # Sentinel errors (ErrRepoNotFound, ErrBranchNotFound, etc.)
├── config.go               # NewRepoManager constructor
├── internal/
│   ├── manager/            # Concrete repoManager — implements RepoManager, delegates to Backend
│   ├── repo/               # Shared Repo implementation — used by both storage backends
│   └── gitutil/            # Shared go-git helper utilities
└── storage/
    ├── filesystem/         # NewFilesystemBackend() — implements Backend (filesystem lifecycle)
    └── arangodb/           # NewArangoBackend()    — implements Backend (ArangoDB lifecycle)
```

## Branching Model

```
main
 │
 ├── task/task-abc-001     ← Agent A works here
 │     commits...
 │     └── auto-merged → main on task completion
 │
 └── task/task-xyz-002     ← Agent B works here (concurrent, isolated)
       commits...
       └── auto-merged → main on task completion
```

### Branch Lifecycle

1. **Task starts** → `CreateBranch("task/{task-id}", from: "main")`
2. **Agent writes files** → `Commit(branch: "task/{task-id}", files, author, message)`
3. **Task completes** → `MergeBranch("task/{task-id}", into: "main")`
   - If fast-forward is possible → merge directly
   - If `main` has advanced → **auto-rebase** task branch onto `main`, then fast-forward merge
   - If rebase conflicts → return `ErrMergeConflict{Files: [...]}` to caller; branch left clean for retry
4. **Branch deleted** → `DeleteBranch("task/{task-id}")`

> **Implementation note**: go-git only supports `FastForwardMerge`. The rebase step must be implemented
> by cherry-picking commits from the task branch onto the latest `main` using go-git's plumbing layer.

## CodeValdCortex Integration Points

| CodeValdCortex Event | CodeValdGit Call |
|---|---|
| Agency created | `RepoManager.InitRepo(agencyID)` |
| Task started | `Repo.CreateBranch(taskID)` |
| Agent writes output | `Repo.WriteFile(taskID, path, content, ...)` |
| Task completed | `Repo.MergeBranch(taskID)` → `Repo.DeleteBranch(taskID)` |
| Agency deleted | `RepoManager.DeleteRepo(agencyID)` |
| UI file browser | `Repo.ListDirectory("main", path)` |
| UI file view | `Repo.ReadFile("main", path)` |
| UI history view | `Repo.Log("main", path)` |

## What Gets Removed from CodeValdCortex

Once CodeValdGit is integrated, the following will be deleted:

- `internal/git/ops/operations.go` — custom SHA-1 blob/tree/commit engine
- `internal/git/storage/repository.go` — ArangoDB Git object storage
- `internal/git/fileindex/service.go` — ArangoDB file index service
- `internal/git/fileindex/repository.go` — ArangoDB file index repository
- `internal/git/models/` — custom Git object models
- ArangoDB collections: `git_objects`, `git_refs`, `repositories`
