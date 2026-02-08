package member

import (
	"context"

	domain "workshop/internal/domain/member"
)

// Store persists Member state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Member, error)
	GetByEmail(ctx context.Context, email string) (domain.Member, error)
	Save(ctx context.Context, value domain.Member) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]domain.Member, error)
	SearchByName(ctx context.Context, query string, limit int) ([]domain.Member, error)
}

// ListFilter carries filtering parameters for List operations.
type ListFilter struct {
	Limit   int
	Offset  int
	Program string
	Status  string
}
