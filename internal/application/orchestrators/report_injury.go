package orchestrators

import (
	"context"
	"errors"
	"time"

	"workshop/internal/domain/injury"

	"github.com/google/uuid"
)

// InjuryStore defines the interface for injury persistence.
type InjuryStore interface {
	Save(ctx context.Context, i injury.Injury) error
}

// ReportInjuryInput carries input for the orchestrator.
type ReportInjuryInput struct {
	BodyPart    string
	Description string
	MemberID    string
}

// ReportInjuryDeps holds dependencies for ReportInjury.
type ReportInjuryDeps struct {
	MemberStore MemberStore
	InjuryStore InjuryStore
}

// ExecuteReportInjury coordinates injury reporting.
// PRE: Member exists, BodyPart specified
// POST: Injury flag created
// INVARIANT: Injury visible for 7 days
func ExecuteReportInjury(ctx context.Context, input ReportInjuryInput, deps ReportInjuryDeps) error {
	// Validate input
	if input.MemberID == "" {
		return errors.New("member ID cannot be empty")
	}
	if input.BodyPart == "" {
		return errors.New("body part cannot be empty")
	}

	// Verify member exists
	_, err := deps.MemberStore.GetByID(ctx, input.MemberID)
	if err != nil {
		return errors.New("member not found")
	}

	// Create injury record
	inj := injury.Injury{
		ID:          uuid.New().String(),
		MemberID:    input.MemberID,
		BodyPart:    input.BodyPart,
		Description: input.Description,
		ReportedAt:  time.Now(),
	}

	// Validate domain rules
	if err := inj.Validate(); err != nil {
		return err
	}

	// Save injury
	if err := deps.InjuryStore.Save(ctx, inj); err != nil {
		return err
	}

	return nil
}
