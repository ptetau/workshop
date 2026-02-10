package waiver

import (
	"context"

	domain "workshop/internal/domain/waiver"
)

// Store persists Waiver state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Waiver, error)
	GetByMemberID(ctx context.Context, memberID string) (domain.Waiver, error)
	Save(ctx context.Context, value domain.Waiver) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]domain.Waiver, error)
}

// ListFilter carries filtering parameters for List operations.
type ListFilter struct {
	Limit  int
	Offset int
}
