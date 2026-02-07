package term

import (
	"context"

	domain "workshop/internal/domain/term"
)

// Store persists Term state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Term, error)
	Save(ctx context.Context, value domain.Term) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Term, error)
}
