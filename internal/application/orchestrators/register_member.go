package orchestrators

import (
	"context"
	"errors"

	"workshop/internal/domain/member"

	"github.com/google/uuid"
)

// MemberStore defines the interface for member persistence.
type MemberStore interface {
	Save(ctx context.Context, m member.Member) error
	GetByID(ctx context.Context, id string) (member.Member, error)
	GetByEmail(ctx context.Context, email string) (member.Member, error)
}

// RegisterMemberInput carries input for the orchestrator.
type RegisterMemberInput struct {
	Email   string
	Name    string
	Program string
}

// RegisterMemberDeps holds dependencies for RegisterMember.
type RegisterMemberDeps struct {
	MemberStore MemberStore
}

// ExecuteRegisterMember coordinates member registration.
// PRE: Valid email, non-empty name, valid program
// POST: Member created with ID, Status=active
// INVARIANT: Email must be unique (enforced by store)
func ExecuteRegisterMember(ctx context.Context, input RegisterMemberInput, deps RegisterMemberDeps) (string, error) {
	// Validate input
	if input.Name == "" {
		return "", errors.New("name cannot be empty")
	}
	if input.Email == "" {
		return "", errors.New("email cannot be empty")
	}
	if input.Program != member.ProgramAdults && input.Program != member.ProgramKids {
		return "", errors.New("program must be 'adults' or 'kids'")
	}

	// Create member with generated ID
	m := member.Member{
		ID:      uuid.New().String(),
		Name:    input.Name,
		Email:   input.Email,
		Program: input.Program,
		Status:  member.StatusActive,
		Fee:     0, // Default fee, can be updated later
	}

	// Validate domain rules
	if err := m.Validate(); err != nil {
		return "", err
	}

	// Save to store
	if err := deps.MemberStore.Save(ctx, m); err != nil {
		return "", err
	}

	return m.ID, nil
}
