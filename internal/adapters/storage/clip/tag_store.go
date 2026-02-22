package clip

import (
	"context"

	domain "workshop/internal/domain/clip"
)

// TagStore persists Tag state and clip-tag associations.
type TagStore interface {
	// Tag management
	SaveTag(ctx context.Context, tag domain.Tag) error
	GetTagByID(ctx context.Context, id string) (domain.Tag, error)
	GetTagByName(ctx context.Context, name string) (domain.Tag, error)
	ListTags(ctx context.Context) ([]domain.Tag, error)
	DeleteTag(ctx context.Context, id string) error

	// Clip-Tag associations
	AddTagToClip(ctx context.Context, clipTag domain.ClipTag) error
	RemoveTagFromClip(ctx context.Context, clipID, tagID string) error
	GetTagsForClip(ctx context.Context, clipID string) ([]domain.Tag, error)
	GetClipsForTag(ctx context.Context, tagID string) ([]string, error)
	SearchClipsByTags(ctx context.Context, tagIDs []string) ([]domain.Clip, error)
}
