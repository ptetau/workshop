package deletion

import (
	"context"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/deletion"
)

// Store defines the interface for deletion request persistence.
type Store interface {
	// GetByID retrieves a deletion request by its ID.
	// PRE: id is non-empty
	// POST: Returns the request or an error if not found
	GetByID(ctx context.Context, id string) (domain.Request, error)

	// GetByMemberID retrieves the most recent deletion request for a member.
	// PRE: memberID is non-empty
	// POST: Returns the request or an error if not found
	GetByMemberID(ctx context.Context, memberID string) (domain.Request, error)

	// Save persists a deletion request to the database.
	// PRE: entity has been validated
	// POST: Entity is persisted (insert or update)
	Save(ctx context.Context, r domain.Request) error

	// ListPending returns deletion requests that need processing (confirmed, grace period ended).
	// PRE: limit > 0
	// POST: Returns up to limit entries ordered by grace_period_end
	ListPending(ctx context.Context, limit int) ([]domain.Request, error)

	// ListByStatus returns deletion requests filtered by status.
	// PRE: status is non-empty, limit > 0
	// POST: Returns matching entries ordered by requested_at desc
	ListByStatus(ctx context.Context, status string, limit int) ([]domain.Request, error)
}

// Ensure SQLiteStore implements Store interface.
var _ Store = (*SQLiteStore)(nil)

// SQLDB defines the database interface needed by the store.
type SQLDB interface {
	storage.SQLDB
}
