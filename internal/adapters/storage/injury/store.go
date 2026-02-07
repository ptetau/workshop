package injury

import (
	"context"

	domain "workshop/internal/domain/injury"
)

// Store persists Injury state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Injury, error)
	Save(ctx context.Context, value domain.Injury) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]domain.Injury, error)
}

// ListFilter carries filtering parameters for List operations.
type ListFilter struct {
	Limit  int
	Offset int
}
