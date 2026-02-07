package program

import (
	"context"

	domain "workshop/internal/domain/program"
)

// Store persists Program state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Program, error)
	Save(ctx context.Context, value domain.Program) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Program, error)
}
