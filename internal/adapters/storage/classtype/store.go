package classtype

import (
	"context"

	domain "workshop/internal/domain/classtype"
)

// Store persists ClassType state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.ClassType, error)
	Save(ctx context.Context, value domain.ClassType) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.ClassType, error)
	ListByProgramID(ctx context.Context, programID string) ([]domain.ClassType, error)
}
