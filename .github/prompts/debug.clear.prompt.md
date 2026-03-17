---
agent: agent
---

# Debug Print Removal

## Task Identification

Get the current task ID from:
1. Git branch name (e.g. `feature/DT-003_traverse-graph` → `DT-003`)
2. Active file context or user mention

## What to Remove

```go
// Remove lines like:
log.Printf("[DT-003] ...")
fmt.Printf("[DT-003] ...")
// TODO: Remove debug prints for DT-003 after issue is resolved
```

## Search Commands

```bash
# Find all debug prints for task
grep -rn "\[DT-003\]" . --include="*.go"

# Find all TODO comments for task
grep -rn "TODO.*DT-003" . --include="*.go"

# Verify nothing remains
grep -rn "fmt\.Printf\|fmt\.Println" . --include="*.go"
grep -rn "log\.Printf.*DT-\|log\.Println.*DT-" . --include="*.go"
```

## What to Keep

**DO NOT** remove:
- Production logging (no task ID prefix)
- Error handling logs (e.g. `log.Printf("codevalddt: register error: %v", err)`)
- Standard startup/shutdown logs

## Post-Removal Validation

```bash
go build ./...      # catch unused imports after removal
go vet ./...        # must show 0 issues
go test -v ./...    # must still pass
```
