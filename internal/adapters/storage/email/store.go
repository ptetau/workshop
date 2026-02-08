package email

import (
	"context"

	domain "workshop/internal/domain/email"
)

// Store persists Email state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Email, error)
	Save(ctx context.Context, e domain.Email) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]domain.Email, error)
	SaveRecipients(ctx context.Context, emailID string, recipients []domain.Recipient) error
	GetRecipients(ctx context.Context, emailID string) ([]domain.Recipient, error)
	ListByRecipientMemberID(ctx context.Context, memberID string) ([]domain.Email, error)
}

// ListFilter specifies criteria for listing emails.
type ListFilter struct {
	Status   string // Filter by status (empty = all)
	SenderID string // Filter by sender (empty = all)
	Search   string // Keyword search in subject/body (empty = all)
}
