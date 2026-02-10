package milestone

import (
	"context"

	domain "workshop/internal/domain/milestone"
)

// Store persists Milestone state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Milestone, error)
	Save(ctx context.Context, value domain.Milestone) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Milestone, error)
}

// MemberMilestoneStore persists earned milestone state for members.
type MemberMilestoneStore interface {
	Save(ctx context.Context, value domain.MemberMilestone) error
	ListByMemberID(ctx context.Context, memberID string) ([]domain.MemberMilestone, error)
	MarkNotified(ctx context.Context, id string) error
	ListUnnotifiedByMemberID(ctx context.Context, memberID string) ([]domain.MemberMilestone, error)
}
