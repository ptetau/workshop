package personalgoal

import (
	"context"

	domain "workshop/internal/domain/personalgoal"
)

// Store persists PersonalGoal state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.PersonalGoal, error)
	Save(ctx context.Context, goal domain.PersonalGoal) error
	Delete(ctx context.Context, id string) error
	ListByMemberID(ctx context.Context, memberID string) ([]domain.PersonalGoal, error)
	ListByDateRange(ctx context.Context, memberID string, from, to string) ([]domain.PersonalGoal, error)
}
