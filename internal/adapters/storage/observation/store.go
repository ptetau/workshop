package observation

import (
	"context"

	domain "workshop/internal/domain/observation"
)

// Store persists Observation state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Observation, error)
	Save(ctx context.Context, value domain.Observation) error
	Delete(ctx context.Context, id string) error
	ListByMemberID(ctx context.Context, memberID string) ([]domain.Observation, error)
}
