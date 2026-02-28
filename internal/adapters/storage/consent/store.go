package consent

import (
	"context"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/consent"
)

// Store defines the interface for consent persistence.
type Store interface {
	// Save persists a consent record.
	// PRE: consent is valid
	// POST: Consent is persisted (insert or update)
	Save(ctx context.Context, c domain.Consent) error

	// GetByMemberID retrieves all consent records for a member.
	// PRE: memberID is non-empty
	// POST: Returns all consent records for the member
	GetByMemberID(ctx context.Context, memberID string) ([]domain.Consent, error)

	// GetByType retrieves a specific consent type for a member.
	// PRE: memberID and consentType are non-empty
	// POST: Returns the consent or error if not found
	GetByType(ctx context.Context, memberID string, consentType domain.Type) (domain.Consent, error)

	// HasValidConsent checks if member has valid consent of a specific type.
	// PRE: memberID and consentType are non-empty
	// POST: Returns true if consent exists, granted, and not revoked
	HasValidConsent(ctx context.Context, memberID string, consentType domain.Type) (bool, error)
}

// Ensure SQLiteStore implements Store interface.
var _ Store = (*SQLiteStore)(nil)

// SQLDB defines the database interface needed by the store.
type SQLDB interface {
	storage.SQLDB
}
