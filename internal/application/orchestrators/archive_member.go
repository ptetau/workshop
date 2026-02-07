package orchestrators

import (
	"context"
	"errors"
	"log/slog"

	"workshop/internal/domain/member"
)

// MemberStoreForArchive defines the store interface needed by Archive/Restore.
type MemberStoreForArchive interface {
	GetByID(ctx context.Context, id string) (member.Member, error)
	Save(ctx context.Context, m member.Member) error
}

// ArchiveMemberInput carries input for the archive orchestrator.
type ArchiveMemberInput struct {
	MemberID string
}

// ArchiveMemberDeps holds dependencies for ArchiveMember.
type ArchiveMemberDeps struct {
	MemberStore MemberStoreForArchive
}

// ExecuteArchiveMember archives a member.
// PRE: MemberID must be non-empty; member must exist and not be archived
// POST: Member status set to archived
func ExecuteArchiveMember(ctx context.Context, input ArchiveMemberInput, deps ArchiveMemberDeps) error {
	if input.MemberID == "" {
		return errors.New("member ID is required")
	}

	m, err := deps.MemberStore.GetByID(ctx, input.MemberID)
	if err != nil {
		return err
	}

	if err := m.Archive(); err != nil {
		return err
	}

	if err := deps.MemberStore.Save(ctx, m); err != nil {
		return err
	}

	slog.Info("member_event", "event", "member_archived", "member_id", input.MemberID)
	return nil
}

// RestoreMemberInput carries input for the restore orchestrator.
type RestoreMemberInput struct {
	MemberID string
}

// RestoreMemberDeps holds dependencies for RestoreMember.
type RestoreMemberDeps struct {
	MemberStore MemberStoreForArchive
}

// ExecuteRestoreMember restores an archived member to active status.
// PRE: MemberID must be non-empty; member must exist and be archived
// POST: Member status set to active
func ExecuteRestoreMember(ctx context.Context, input RestoreMemberInput, deps RestoreMemberDeps) error {
	if input.MemberID == "" {
		return errors.New("member ID is required")
	}

	m, err := deps.MemberStore.GetByID(ctx, input.MemberID)
	if err != nil {
		return err
	}

	if err := m.Restore(); err != nil {
		return err
	}

	if err := deps.MemberStore.Save(ctx, m); err != nil {
		return err
	}

	slog.Info("member_event", "event", "member_restored", "member_id", input.MemberID)
	return nil
}
