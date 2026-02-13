package calendar

import (
	"context"

	domain "workshop/internal/domain/calendar"
)

// Store persists CalendarEvent state.
type Store interface {
	Save(ctx context.Context, e domain.Event) error
	GetByID(ctx context.Context, id string) (domain.Event, error)
	ListByDateRange(ctx context.Context, from, to string) ([]domain.Event, error)
	Delete(ctx context.Context, id string) error
}
