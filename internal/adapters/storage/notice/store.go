package notice

import (
	"context"
	"time"

	domain "workshop/internal/domain/notice"
)

// Store persists Notice state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Notice, error)
	Save(ctx context.Context, value domain.Notice) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]domain.Notice, error)
	ListPublished(ctx context.Context, noticeType string, now time.Time) ([]domain.Notice, error)
}

// ListFilter carries filtering parameters for List operations.
type ListFilter struct {
	Type   string
	Status string
	Limit  int
	Offset int
}
