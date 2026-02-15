package featureflag

import (
	"context"

	domain "workshop/internal/domain/featureflag"
)

// Store persists FeatureFlag state.
type Store interface {
	GetByKey(ctx context.Context, key string) (domain.FeatureFlag, error)
	List(ctx context.Context) ([]domain.FeatureFlag, error)
	Save(ctx context.Context, value domain.FeatureFlag) error
}
