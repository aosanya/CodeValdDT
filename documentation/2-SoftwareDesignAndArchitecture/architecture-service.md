# CodeValdDT — Architecture: Service, gRPC & Registration

> Part of [architecture.md](architecture.md)

## 4. Package Structure

```
CodeValdDT/
├── cmd/
│   └── main.go                   # Wires dependencies only — no business logic
├── go.mod
├── errors.go                     # ErrEntityNotFound, ErrRelationshipNotFound, ErrImmutableType, etc.
├── models.go                     # DTDataManager = entitygraph.DataManager alias; DTSchemaManager alias
├── codevalddt.go                 # Package-level godoc; imports entitygraph aliases
├── internal/
│   ├── config/
│   │   └── config.go             # Config struct + loader (env / YAML)
│   ├── manager/
│   │   └── manager.go            # Concrete DTDataManager — holds DTSchemaManager + CrossClient
│   ├── server/
│   │   └── server.go             # Inbound gRPC server — DTService handlers
│   └── registrar/
│       └── registrar.go          # Cross registration heartbeat loop
├── storage/
│   └── arangodb/
│       └── storage.go            # ArangoDB DTSchemaManager implementation
├── proto/
│   └── codevalddt/
│       └── dt.proto              # DTService gRPC definition
├── gen/
│   └── go/                       # Generated protobuf code (buf generate — do not hand-edit)
└── bin/
    └── codevalddt                # Compiled binary
```

---

## 5. gRPC Service Definition

```protobuf
syntax = "proto3";
package codevalddt.v1;

service DTService {
    // Entity management
    rpc CreateEntity         (CreateEntityRequest)         returns (Entity);
    rpc GetEntity            (GetEntityRequest)            returns (Entity);
    rpc UpdateEntity         (UpdateEntityRequest)         returns (Entity);
    rpc DeleteEntity         (DeleteEntityRequest)         returns (google.protobuf.Empty);
    rpc ListEntities         (ListEntitiesRequest)         returns (ListEntitiesResponse);

    // Graph operations
    rpc CreateRelationship   (CreateRelationshipRequest)   returns (Relationship);
    rpc GetRelationship      (GetRelationshipRequest)      returns (Relationship);
    rpc DeleteRelationship   (DeleteRelationshipRequest)   returns (google.protobuf.Empty);
    rpc ListRelationships    (ListRelationshipsRequest)    returns (ListRelationshipsResponse);
    // TraverseGraphResponse contains both vertices (repeated Entity) and
    // edges (repeated Relationship) — matches TraverseGraphResult in models.go.
    rpc TraverseGraph        (TraverseGraphRequest)        returns (TraverseGraphResponse);
}
```

Generated Go code lives in `gen/go/`. **Never hand-edit generated files.**

---

## 6. CodeValdCross Registration

On startup, `cmd/main.go` starts a registration heartbeat. The loop calls
`OrchestratorService.Register` on CodeValdCross every **20 seconds**.

```go
RegisterRequest{
    ServiceName: "codevalddt",
    Addr:        ":50055",
    Produces: []string{
        "cross.dt.{agencyID}.entity.created",
        "cross.dt.{agencyID}.telemetry.recorded",
        "cross.dt.{agencyID}.event.recorded",
    },
    Consumes: []string{},
    Routes: []Route{
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "PUT",    Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "DELETE", Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/relationships"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/relationships/{relationshipId}"},
        {Method: "DELETE", Pattern: "/{agencyId}/dt/relationships/{relationshipId}"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/relationships"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities/{entityId}/traverse"},
    },
}
```

The repeat call is the **liveness signal** — Cross expires services that stop
registering. If Cross is not yet up, the loop retries silently.
