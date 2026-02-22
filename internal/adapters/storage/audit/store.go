package audit

import (
	"context"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/audit"
)

// Store defines the interface for audit event persistence.
type Store interface {
	// Save persists an audit event.
	// PRE: event is valid
	// POST: Event is persisted
	Save(ctx context.Context, event domain.Event) error

	// List returns audit events with optional filtering.
	// PRE: limit > 0
	// POST: Returns events ordered by timestamp desc
	List(ctx context.Context, filter Filter, limit int) ([]domain.Event, error)

	// GetByID retrieves a specific audit event.
	// PRE: id is non-empty
	// POST: Returns the event or error if not found
	GetByID(ctx context.Context, id string) (domain.Event, error)
}

// Filter defines query parameters for listing audit events.
type Filter struct {
	Category   *domain.Category
	Action     *domain.Action
	ActorID    *string
	Severity   *domain.Severity
	ResourceID *string
	FromDate   *string
	ToDate     *string
}

// Ensure SQLiteStore implements Store interface.
var _ Store = (*SQLiteStore)(nil)

// SQLDB defines the database interface needed by the store.
type SQLDB interface {
	storage.SQLDB
}
