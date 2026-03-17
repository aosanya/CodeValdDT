---
agent: agent
---

# Start New Task

> ⚠️ **Before starting**, run `CodeValdDT/.github/prompts/finish-task.prompt.md`
> to ensure any in-progress task is properly completed and merged first.

## Task Startup Process (MANDATORY)

1. **Select the next task**
   - Check `documentation/3-SofwareDevelopment/mvp.md` for the task list and
     current status
   - Check `documentation/3-SofwareDevelopment/mvp-details/` for detailed
     specs per topic
   - Check `documentation/1-SoftwareRequirements/requirements.md` for
     unimplemented functional requirements
   - Follow the onion approach — core entity CRUD before graph traversal before
     telemetry before events

2. **Read the specification**
   - Re-read the relevant FRs in `requirements.md`
   - Re-read the corresponding section in `architecture.md`
   - Read the task spec in `documentation/3-SofwareDevelopment/mvp-details/`
   - Understand how the task fits into `DTManager`
   - Note the mandatory pub/sub requirements for `CreateEntity` and
     `RecordTelemetry`

3. **Create feature branch from `main`**
   ```bash
   cd /workspaces/CodeVald-AIProject/CodeValdDT
   git checkout main
   git pull origin main
   git checkout -b feature/DT-XXX_description
   ```
   Branch naming: `feature/DT-XXX_description` (lowercase with underscores)

4. **Read project guidelines**
   - Review `.github/instructions/rules.instructions.md`
   - Key rules: interface-first, inject Backend, publish events, no AI/frontend
     logic, context propagation, godoc on all exports, edge collection for
     relationships

5. **Create a todo list**
   - Break the task into actionable steps
   - Use the manage_todo_list tool to track progress

## Pre-Implementation Checklist

- [ ] Relevant FRs and architecture sections re-read
- [ ] Feature branch created: `feature/DT-XXX_description`
- [ ] Existing files checked — no duplicate types in `models.go` or `errors.go`
- [ ] Understood which files to modify (`internal/manager/`, `internal/server/`,
      `storage/arangodb/`, `cmd/`, `proto/`)
- [ ] Todo list created

## Development Standards

- **No AI/LLM logic, no frontend serving** — this service manages twin data only
- **`DTManager` is the only entry point** — gRPC handlers delegate to it
- **`cross.dt.{agencyID}.entity.created` is mandatory** — publish on every
  successful `CreateEntity`
- **`cross.dt.{agencyID}.telemetry.recorded` is mandatory** — publish on every
  successful `RecordTelemetry`
- **`relationships` is an edge collection** — never use a document collection
- **Backend is injected** — never hardcode ArangoDB connection in manager
- **Every exported symbol** must have a godoc comment
- **Every exported method** takes `context.Context` as the first argument
- **Registration heartbeat** — call `Register` on Cross every 20 seconds

## Git Workflow

```bash
# Create feature branch
git checkout -b feature/DT-XXX_description

# Regular commits during development
git add .
git commit -m "DT-XXX: Descriptive message"

# Build validation before merge
go build ./...           # must succeed
go test -v -race ./...   # must pass
go vet ./...             # must show 0 issues
golangci-lint run ./...  # must pass

# Merge
git checkout main
git merge feature/DT-XXX_description --no-ff
git branch -d feature/DT-XXX_description
```
