package message

import (
	"context"

	domain "workshop/internal/domain/message"
)

// Store persists Message state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Message, error)
	Save(ctx context.Context, value domain.Message) error
	Delete(ctx context.Context, id string) error
	ListByReceiverID(ctx context.Context, receiverID string) ([]domain.Message, error)
	CountUnread(ctx context.Context, receiverID string) (int, error)
}
