package calendar

import (
	"context"

	domain "workshop/internal/domain/calendar"
)

// InterestStore persists CompetitionInterest state.
type InterestStore interface {
	Save(ctx context.Context, ci domain.CompetitionInterest) error
	Delete(ctx context.Context, eventID, memberID string) error
	GetByEvent(ctx context.Context, eventID string) ([]domain.CompetitionInterest, error)
	GetByMember(ctx context.Context, memberID string) ([]domain.CompetitionInterest, error)
	CountByEvent(ctx context.Context, eventID string) (int, error)
	IsInterested(ctx context.Context, eventID, memberID string) (bool, error)
}
