package theme

import (
	"context"

	domain "workshop/internal/domain/theme"
)

// Store persists Theme state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Theme, error)
	Save(ctx context.Context, value domain.Theme) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Theme, error)
	ListByProgram(ctx context.Context, program string) ([]domain.Theme, error)
}
