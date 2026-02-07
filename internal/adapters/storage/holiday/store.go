package holiday

import (
	"context"

	domain "workshop/internal/domain/holiday"
)

// Store persists Holiday state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Holiday, error)
	Save(ctx context.Context, value domain.Holiday) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Holiday, error)
}
