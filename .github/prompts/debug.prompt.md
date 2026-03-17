---
agent: agent
---

# Debug a CodeValdDT Issue

## Common Failure Scenarios

### Scenario 1: `CreateEntity` Succeeds but `entity.created` Never Published
**Symptom**: Entity appears in ArangoDB but no downstream reaction via Cross  
**Cause**: `cross.dt.{agencyID}.entity.created` publish is missing or Cross client is nil  
**Check**: Confirm `m.crossClient.Publish(...)` is called after `backend.InsertEntity`

### Scenario 2: `TraverseGraph` Returns Empty or Wrong Results
**Symptom**: Traversal from a known entity returns no results  
**Cause**: Graph not defined in ArangoDB, wrong edge collection name, or wrong direction  
**Check**: Verify the named graph `dt_graph` includes the `relationships` edge collection; check `_from`/`_to` document handles use the correct collection prefix (`entities/`)

### Scenario 3: `CreateRelationship` Panics with "collection not an edge collection"
**Symptom**: ArangoDB rejects the insert  
**Cause**: `relationships` collection was created as a document collection, not an edge collection  
**Check**: Drop and recreate as edge collection type in `storage/arangodb/storage.go`

### Scenario 4: `RecordTelemetry` Stores Data but No Event Published
**Symptom**: Telemetry readable via `QueryTelemetry` but Cross subscribers never fire  
**Cause**: Missing `m.crossClient.Publish(...)` after `backend.InsertTelemetry`  
**Check**: Trace `RecordTelemetry` in `internal/manager/manager.go`

### Scenario 5: `Register` Always Fails with `DeadlineExceeded`
**Symptom**: Heartbeat loop logs errors continuously  
**Cause**: CodeValdCross is not running, or wrong address in config  
**Check**: Verify `CROSS_GRPC_ADDR` env var; confirm Cross is up before starting DT

### Scenario 6: Nil Pointer Panic in Manager
**Symptom**: `nil pointer dereference` in `internal/manager/manager.go`  
**Cause**: `cmd/main.go` did not inject `Backend` or `crossClient`  
**Check**: Trace wiring in `cmd/main.go`

## Debug Print Guidelines

### Prefix Format
All debug prints MUST use: `[DT-XXX]`

```go
log.Printf("[DT-XXX] CreateEntity called: agencyID=%s typeID=%s", req.AgencyID, req.TypeID)
log.Printf("[DT-XXX] Entity inserted: id=%s", entity.ID)
log.Printf("[DT-XXX] Publishing topic=%s entityID=%s", topic, entity.ID)
log.Printf("[DT-XXX] Error in operation: %v", err)
```

### Strategic Placement

1. **Function entry**: `log.Printf("[DT-XXX] CreateEntity called: ...")`
2. **After storage insert**: `log.Printf("[DT-XXX] Entity inserted: id=%s", entity.ID)`
3. **Before/after publish**: `log.Printf("[DT-XXX] Publishing: topic=%s", topic)`
4. **Graph traversal**: `log.Printf("[DT-XXX] TraverseGraph: startID=%s depth=%d", req.StartEntityID, req.Depth)`
5. **Heartbeat**: `log.Printf("[DT-XXX] Register: addr=%s err=%v", addr, err)`

### What NOT to Debug
- Simple getters
- Trivial utility functions
- Already well-instrumented production logs
