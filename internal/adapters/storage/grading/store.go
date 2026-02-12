package grading

import (
	"context"

	domain "workshop/internal/domain/grading"
)

// RecordStore persists GradingRecord state.
type RecordStore interface {
	GetByID(ctx context.Context, id string) (domain.Record, error)
	Save(ctx context.Context, value domain.Record) error
	ListByMemberID(ctx context.Context, memberID string) ([]domain.Record, error)
}

// ConfigStore persists GradingConfig state.
type ConfigStore interface {
	GetByID(ctx context.Context, id string) (domain.Config, error)
	Save(ctx context.Context, value domain.Config) error
	GetByProgramAndBelt(ctx context.Context, program, belt string) (domain.Config, error)
	List(ctx context.Context) ([]domain.Config, error)
}

// NoteStore persists GradingNote state.
type NoteStore interface {
	Save(ctx context.Context, value domain.Note) error
	ListByMemberID(ctx context.Context, memberID string) ([]domain.Note, error)
}

// ProposalStore persists GradingProposal state.
type ProposalStore interface {
	GetByID(ctx context.Context, id string) (domain.Proposal, error)
	Save(ctx context.Context, value domain.Proposal) error
	ListPending(ctx context.Context) ([]domain.Proposal, error)
	ListByMemberID(ctx context.Context, memberID string) ([]domain.Proposal, error)
}
