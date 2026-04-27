package codevalddt

import "errors"

// ErrEntityNotFound is returned when an entity does not exist for the given
// agency and entity ID, or when the entity has been soft-deleted.
var ErrEntityNotFound = errors.New("entity not found")

// ErrRelationshipNotFound is returned when a relationship does not exist for
// the given agency and relationship ID.
var ErrRelationshipNotFound = errors.New("relationship not found")

// ErrSchemaNotFound is returned when no schema (draft or published) exists
// for the agency, or when the requested version cannot be resolved.
var ErrSchemaNotFound = errors.New("schema not found")

// ErrInvalidEntity is returned when a request is missing required fields
// (typically AgencyID or TypeID).
var ErrInvalidEntity = errors.New("invalid entity: missing required fields")

// ErrInvalidRelationship is returned when a relationship request is missing
// required fields (AgencyID, Name, FromID, ToID), or when the proposed edge
// is not declared in the source TypeDefinition.
var ErrInvalidRelationship = errors.New("invalid relationship: missing required fields")

// ErrInvalidSchema is returned when a schema is missing required fields or
// fails validation before being persisted.
var ErrInvalidSchema = errors.New("invalid schema: missing required fields")

// ErrImmutableType is returned by UpdateEntity when the resolved
// TypeDefinition has Immutable set to true. Telemetry readings and event
// records are immutable by definition — only CreateEntity and DeleteEntity
// are valid for these types.
var ErrImmutableType = errors.New("entity type is immutable: update not allowed")
