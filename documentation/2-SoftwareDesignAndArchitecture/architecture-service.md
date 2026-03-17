# CodeValdDT — Architecture: Service, gRPC & Registration

> Part of [architecture.md](architecture.md)

## 4. Package Structure

```
CodeValdDT/
├── cmd/
│   └── main.go                   # Wires dependencies only — no business logic
├── go.mod
├── errors.go                     # ErrEntityNotFound, ErrRelationshipNotFound, etc.
├── models.go                     # Entity, Relationship, TelemetryReading, Event, filter/request types
├── codevalddt.go                 # DTManager + Backend interfaces
├── internal/
│   ├── config/
│   │   └── config.go             # Config struct + loader (env / YAML)
│   ├── manager/
│   │   └── manager.go            # Concrete DTManager — holds Backend + CrossClient
│   ├── server/
│   │   └── server.go             # Inbound gRPC server — DTService handlers
│   └── registrar/
│       └── registrar.go          # Cross registration heartbeat loop
├── storage/
│   └── arangodb/
│       └── storage.go            # ArangoDB Backend implementation
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
    rpc DeleteRelationship   (DeleteRelationshipRequest)   returns (google.protobuf.Empty);
    rpc TraverseGraph        (TraverseGraphRequest)        returns (TraverseGraphResponse);

    // Telemetry
    rpc RecordTelemetry      (RecordTelemetryRequest)      returns (TelemetryReading);
    rpc QueryTelemetry       (QueryTelemetryRequest)       returns (QueryTelemetryResponse);

    // Events
    rpc RecordEvent          (RecordEventRequest)          returns (Event);
    rpc ListEvents           (ListEventsRequest)           returns (ListEventsResponse);

    // Schema management
    rpc PublishSchema        (PublishSchemaRequest)        returns (DTSchema);
    rpc GetSchema            (GetSchemaRequest)            returns (DTSchema);
    rpc ListSchemaVersions   (ListSchemaVersionsRequest)   returns (ListSchemaVersionsResponse);
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
    },
    Consumes: []string{},
    Routes: []Route{
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "PUT",    Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "DELETE", Pattern: "/{agencyId}/dt/entities/{entityId}"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/relationships"},
        {Method: "DELETE", Pattern: "/{agencyId}/dt/relationships/{relationshipId}"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities/{entityId}/traverse"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities/{entityId}/telemetry"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities/{entityId}/telemetry"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/entities/{entityId}/events"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/entities/{entityId}/events"},
        {Method: "POST",   Pattern: "/{agencyId}/dt/schema"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/schema"},
        {Method: "GET",    Pattern: "/{agencyId}/dt/schema/versions"},
    },
}
```

The repeat call is the **liveness signal** — Cross expires services that stop
registering. If Cross is not yet up, the loop retries silently.
