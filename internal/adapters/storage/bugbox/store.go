package bugbox

import (
	"context"

	domain "workshop/internal/domain/bugbox"
)

// Store persists BugBox Submission state.
type Store interface {
	Save(ctx context.Context, s domain.Submission) error
	GetByID(ctx context.Context, id string) (domain.Submission, error)
}
