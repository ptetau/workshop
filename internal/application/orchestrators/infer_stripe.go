package orchestrators

import (
	"context"
	"log/slog"
	"time"

	"workshop/internal/domain/grading"
	"workshop/internal/domain/member"

	"github.com/google/uuid"
)

// InferStripeMemberStore defines the member store interface needed for stripe inference.
type InferStripeMemberStore interface {
	GetByID(ctx context.Context, id string) (member.Member, error)
}

// InferStripeAttendanceStore defines the attendance store interface needed for stripe inference.
type InferStripeAttendanceStore interface {
	SumMatHoursByMemberID(ctx context.Context, memberID string) (float64, error)
}

// InferStripeEstimatedHoursStore defines the estimated hours store interface needed for stripe inference.
type InferStripeEstimatedHoursStore interface {
	SumApprovedByMemberID(ctx context.Context, memberID string) (float64, error)
}

// InferStripeGradingRecordStore defines the grading record store interface needed for stripe inference.
type InferStripeGradingRecordStore interface {
	ListByMemberID(ctx context.Context, memberID string) ([]grading.Record, error)
	Save(ctx context.Context, r grading.Record) error
}

// InferStripeGradingConfigStore defines the grading config store interface needed for stripe inference.
type InferStripeGradingConfigStore interface {
	GetByProgramAndBelt(ctx context.Context, program, belt string) (grading.Config, error)
}

// InferStripeDeps holds dependencies for stripe inference.
type InferStripeDeps struct {
	MemberStore         InferStripeMemberStore
	AttendanceStore     InferStripeAttendanceStore
	EstimatedHoursStore InferStripeEstimatedHoursStore // optional: nil skips bulk estimates
	GradingRecordStore  InferStripeGradingRecordStore
	GradingConfigStore  InferStripeGradingConfigStore
}

// ExecuteInferStripe checks whether a member's stripe count should increase
// based on accumulated mat hours and auto-creates an inferred grading record if so.
// PRE: memberID is non-empty, attendance record already saved
// POST: If inferred stripe > current stripe (same belt), a new grading record is saved with MethodInferred
func ExecuteInferStripe(ctx context.Context, memberID string, deps InferStripeDeps) error {
	// Look up member to get their program
	m, err := deps.MemberStore.GetByID(ctx, memberID)
	if err != nil {
		return nil // member not found — skip silently (best-effort)
	}

	// Get current belt and stripe from latest grading record
	currentBelt := grading.BeltWhite
	currentStripe := 0
	gradingRecords, err := deps.GradingRecordStore.ListByMemberID(ctx, memberID)
	if err != nil {
		return nil // skip silently
	}
	if len(gradingRecords) > 0 {
		latest := gradingRecords[0]
		for _, r := range gradingRecords[1:] {
			if r.PromotedAt.After(latest.PromotedAt) {
				latest = r
			}
		}
		currentBelt = latest.Belt
		currentStripe = latest.Stripe
	}

	// Determine next belt in progression to get the config
	nextBelt := nextBeltForProgram(currentBelt, m.Program)
	if nextBelt == "" {
		return nil // at highest belt, no stripe inference needed
	}

	// Get config for the next belt (contains FlightTimeHours and StripeCount)
	config, err := deps.GradingConfigStore.GetByProgramAndBelt(ctx, m.Program, nextBelt)
	if err != nil {
		return nil // no config — skip silently
	}
	if config.FlightTimeHours <= 0 || config.StripeCount <= 0 {
		return nil // not applicable (e.g. kids using attendance-based grading)
	}

	// Compute total mat hours
	attendanceHours, err := deps.AttendanceStore.SumMatHoursByMemberID(ctx, memberID)
	if err != nil {
		return nil
	}
	totalHours := attendanceHours

	// Add bulk-estimated hours if available
	if deps.EstimatedHoursStore != nil {
		bulkHours, err := deps.EstimatedHoursStore.SumApprovedByMemberID(ctx, memberID)
		if err == nil {
			totalHours += bulkHours
		}
	}

	// Infer what stripe the member should have
	inferredStripe := grading.InferStripe(totalHours, config)

	if inferredStripe <= currentStripe {
		return nil // no change needed
	}

	// Create a new inferred grading record
	record := grading.Record{
		ID:         uuid.New().String(),
		MemberID:   memberID,
		Belt:       currentBelt,
		Stripe:     inferredStripe,
		PromotedAt: time.Now(),
		Method:     grading.MethodInferred,
	}
	if err := record.Validate(); err != nil {
		return nil // invalid record — skip
	}
	if err := deps.GradingRecordStore.Save(ctx, record); err != nil {
		slog.Error("infer_stripe_error", "error", err, "member_id", memberID)
		return nil // best-effort, don't fail the check-in
	}

	slog.Info("grading_event", "event", "stripe_inferred",
		"member_id", memberID,
		"belt", currentBelt,
		"old_stripe", currentStripe,
		"new_stripe", inferredStripe,
		"total_hours", totalHours,
	)
	return nil
}

// nextBeltForProgram returns the next belt in the progression for the given program.
// Returns "" if the member is at the highest belt.
func nextBeltForProgram(currentBelt, program string) string {
	var belts []string
	switch program {
	case "kids":
		belts = grading.KidsBelts
	default:
		belts = grading.AdultBelts
	}
	for i, b := range belts {
		if b == currentBelt && i+1 < len(belts) {
			return belts[i+1]
		}
	}
	return ""
}
