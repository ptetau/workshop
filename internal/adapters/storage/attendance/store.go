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
	ListByMemberIDAndDate(ctx context.Context, memberID string, date string) ([]domain.Attendance, error)
	ListByDateRange(ctx context.Context, startDate string, endDate string) ([]domain.Attendance, error)
	ListDistinctMemberIDsByScheduleAndDate(ctx context.Context, scheduleID string, classDate string) ([]string, error)
	ListDistinctMemberIDsByScheduleIDsSince(ctx context.Context, scheduleIDs []string, since string) ([]string, error)
	ListByMemberIDAndDateRange(ctx context.Context, memberID string, startDate string, endDate string) ([]domain.Attendance, error)
	DeleteByMemberIDAndDateRange(ctx context.Context, memberID string, startDate string, endDate string) (int, error)
	SumMatHoursByMemberID(ctx context.Context, memberID string) (float64, error)
}

// ListFilter carries filtering parameters for List operations.
type ListFilter struct {
	Limit  int
	Offset int
}
