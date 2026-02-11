package estimatedhours

import (
	"context"

	domain "workshop/internal/domain/estimatedhours"
)

// Store persists EstimatedHours state.
type Store interface {
	Save(ctx context.Context, e domain.EstimatedHours) error
	GetByID(ctx context.Context, id string) (domain.EstimatedHours, error)
	ListByMemberID(ctx context.Context, memberID string) ([]domain.EstimatedHours, error)
	Delete(ctx context.Context, id string) error
	SumApprovedByMemberID(ctx context.Context, memberID string) (float64, error)
}
