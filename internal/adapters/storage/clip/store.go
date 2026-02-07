package clip

import (
	"context"

	domain "workshop/internal/domain/clip"
)

// Store persists Clip state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Clip, error)
	Save(ctx context.Context, value domain.Clip) error
	Delete(ctx context.Context, id string) error
	ListByThemeID(ctx context.Context, themeID string) ([]domain.Clip, error)
	ListPromoted(ctx context.Context) ([]domain.Clip, error)
}
