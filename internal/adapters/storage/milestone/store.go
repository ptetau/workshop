package milestone

import (
	"context"

	domain "workshop/internal/domain/milestone"
)

// Store persists Milestone state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Milestone, error)
	Save(ctx context.Context, value domain.Milestone) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Milestone, error)
}
