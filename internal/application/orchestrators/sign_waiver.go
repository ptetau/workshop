package orchestrators

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"workshop/internal/domain/member"
	"workshop/internal/domain/waiver"

	"github.com/google/uuid"
)

// WaiverStore defines the interface for waiver persistence.
type WaiverStore interface {
	Save(ctx context.Context, w waiver.Waiver) error
}

// SignWaiverInput carries input for the orchestrator.
type SignWaiverInput struct {
	AcceptedTerms bool
	Email         string
	MemberName    string
	IPAddress     string // Passed from HTTP context
}

// SignWaiverDeps holds dependencies for SignWaiver.
type SignWaiverDeps struct {
	MemberStore MemberStore
	WaiverStore WaiverStore
}

// ExecuteSignWaiver coordinates digital waiver signing.
// PRE: Valid email, AcceptedTerms=true
// POST: Waiver created with timestamp and IP
// INVARIANT: One active waiver per member
func ExecuteSignWaiver(ctx context.Context, input SignWaiverInput, deps SignWaiverDeps) error {
	// Validate input
	if input.MemberName == "" {
		return errors.New("member name cannot be empty")
	}
	if input.Email == "" {
		return errors.New("email cannot be empty")
	}
	if !input.AcceptedTerms {
		return errors.New("terms must be accepted")
	}

	// Find or create member by email
	existing, err := deps.MemberStore.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	memberID := existing.ID
	if errors.Is(err, sql.ErrNoRows) {
		memberID = uuid.New().String()
		m := member.Member{
			ID:      memberID,
			Name:    input.MemberName,
			Email:   input.Email,
			Program: member.ProgramAdults, // Default
			Status:  member.StatusActive,
		}

		if err := m.Validate(); err != nil {
			return err
		}

		if err := deps.MemberStore.Save(ctx, m); err != nil {
			return err
		}
	}

	// Create waiver
	w := waiver.Waiver{
		ID:            uuid.New().String(),
		MemberID:      memberID,
		AcceptedTerms: input.AcceptedTerms,
		IPAddress:     input.IPAddress,
		SignedAt:      time.Now(),
	}

	// Validate domain rules
	if err := w.Validate(); err != nil {
		return err
	}

	// Save waiver
	if err := deps.WaiverStore.Save(ctx, w); err != nil {
		return err
	}

	return nil
}
