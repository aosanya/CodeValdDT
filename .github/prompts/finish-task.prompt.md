---
agent: agent
---

# Complete and Merge Current Task

## Completion Process (MANDATORY)

1. **Validate code quality**
   ```bash
   go build ./...           # Must succeed
   go test -v -race ./...   # Must pass
   go vet ./...             # Must show 0 issues
   golangci-lint run ./...  # Must pass
   ```

2. **Remove all debug logs before merge (MANDATORY)**
   ```bash
   grep -r "fmt.Printf\|fmt.Println" . --include="*.go"
   grep -r "log.Printf.*DT-\|log.Println.*DT-" . --include="*.go"
   ```

3. **Verify service contract compliance**
   - [ ] All new exported symbols have godoc comments
   - [ ] All new exported methods accept `context.Context` as first argument
   - [ ] `Backend` is injected — no hardcoded ArangoDB in manager
   - [ ] `cross.dt.{agencyID}.entity.created` published on every `CreateEntity`
   - [ ] `cross.dt.{agencyID}.telemetry.recorded` published on every
         `RecordTelemetry`
   - [ ] `relationships` uses edge collection — not a document collection
   - [ ] No AI/LLM logic, no frontend serving added
   - [ ] Errors are typed (`ErrEntityNotFound`, not raw strings)
   - [ ] No file exceeds 500 lines
   - [ ] gRPC handlers delegate to `DTManager` — no business logic in handlers

4. **Update documentation if architecture changed**
   - Update `documentation/2-SoftwareDesignAndArchitecture/architecture.md` if
     the implementation deviated
   - Update task status in `documentation/3-SofwareDevelopment/mvp.md`
     (🔲 → ✅)

5. **Merge to main**
   ```bash
   go build ./...
   go test -v -race ./...
   go vet ./...

   git add .
   git commit -m "DT-XXX: Implement [description]

   - Key implementation detail 1
   - Key implementation detail 2
   - Removed all debug logs
   - All tests pass with -race
   "

   git checkout main
   git merge feature/DT-XXX_description --no-ff -m "Merge DT-XXX: [description]"
   git branch -d feature/DT-XXX_description
   ```

## Success Criteria

- ✅ `go build ./...` succeeds
- ✅ `go test -race ./...` passes
- ✅ `go vet ./...` shows 0 issues
- ✅ All debug logs removed
- ✅ Service contract compliance verified
- ✅ Documentation updated if architecture changed
- ✅ Merged to `main` and feature branch deleted
