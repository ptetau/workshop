package traininggoal

import (
	"context"

	domain "workshop/internal/domain/traininggoal"
)

// Store persists TrainingGoal state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.TrainingGoal, error)
	Save(ctx context.Context, value domain.TrainingGoal) error
	Delete(ctx context.Context, id string) error
	GetActiveByMemberID(ctx context.Context, memberID string) (domain.TrainingGoal, error)
	ListByMemberID(ctx context.Context, memberID string) ([]domain.TrainingGoal, error)
}
