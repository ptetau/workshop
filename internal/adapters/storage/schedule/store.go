package schedule

import (
	"context"

	domain "workshop/internal/domain/schedule"
)

// Store persists Schedule state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Schedule, error)
	Save(ctx context.Context, value domain.Schedule) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Schedule, error)
	ListByDay(ctx context.Context, day string) ([]domain.Schedule, error)
	ListByClassTypeID(ctx context.Context, classTypeID string) ([]domain.Schedule, error)
}
