package attendance

import (
	"context"

	domain "workshop/internal/domain/attendance"
)

// Store persists Attendance state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Attendance, error)
	Save(ctx context.Context, value domain.Attendance) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]domain.Attendance, error)
	ListByMemberID(ctx context.Context, memberID string) ([]domain.Attendance, error)
}

// ListFilter carries filtering parameters for List operations.
type ListFilter struct {
	Limit  int
	Offset int
}
