package orchestrators

import (
	"context"
	"time"

	domain "workshop/internal/domain/estimatedhours"
)

// EstimatedHoursStoreForSelfEstimate defines the store interface needed by self-estimate orchestrators.
type EstimatedHoursStoreForSelfEstimate interface {
	Save(ctx context.Context, e domain.EstimatedHours) error
	GetByID(ctx context.Context, id string) (domain.EstimatedHours, error)
}

// SubmitSelfEstimateInput carries input for the self-estimate submission orchestrator.
type SubmitSelfEstimateInput struct {
	MemberID    string
	StartDate   string
	EndDate     string
	WeeklyHours float64
	Note        string
}

// SubmitSelfEstimateDeps holds dependencies for the self-estimate submission orchestrator.
type SubmitSelfEstimateDeps struct {
	EstimatedHoursStore EstimatedHoursStoreForSelfEstimate
	GenerateID          func() string
	Now                 func() time.Time
}

// ExecuteSubmitSelfEstimate creates a new self-estimate entry with status=pending.
// PRE: input fields are populated, deps are valid
// POST: estimated hours entry is persisted with source=self_estimate, status=pending
func ExecuteSubmitSelfEstimate(ctx context.Context, input SubmitSelfEstimateInput, deps SubmitSelfEstimateDeps) (domain.EstimatedHours, error) {
	entry := domain.EstimatedHours{
		ID:          deps.GenerateID(),
		MemberID:    input.MemberID,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		WeeklyHours: input.WeeklyHours,
		Source:      domain.SourceSelfEstimate,
		Status:      domain.StatusPending,
		Note:        input.Note,
		CreatedBy:   input.MemberID, // member submits for themselves
		CreatedAt:   deps.Now(),
	}

	if err := entry.CalculateTotalHours(); err != nil {
		return domain.EstimatedHours{}, err
	}

	if err := entry.Validate(); err != nil {
		return domain.EstimatedHours{}, err
	}

	if err := deps.EstimatedHoursStore.Save(ctx, entry); err != nil {
		return domain.EstimatedHours{}, err
	}

	return entry, nil
}

// ReviewSelfEstimateInput carries input for the review orchestrator.
type ReviewSelfEstimateInput struct {
	ID            string  // estimated hours entry ID
	Action        string  // "approve" or "reject"
	AdjustedHours float64 // optional: override total hours on approve (0 = keep original)
	ReviewNote    string  // required for reject, optional for approve
	ReviewerID    string  // account ID of admin/coach
}

// ReviewSelfEstimateDeps holds dependencies for the review orchestrator.
type ReviewSelfEstimateDeps struct {
	EstimatedHoursStore EstimatedHoursStoreForSelfEstimate
	Now                 func() time.Time
}

// ExecuteReviewSelfEstimate approves or rejects a pending self-estimate.
// PRE: input.ID refers to a pending entry, input.Action is "approve" or "reject"
// POST: entry status is updated, review fields populated
func ExecuteReviewSelfEstimate(ctx context.Context, input ReviewSelfEstimateInput, deps ReviewSelfEstimateDeps) (domain.EstimatedHours, error) {
	entry, err := deps.EstimatedHoursStore.GetByID(ctx, input.ID)
	if err != nil {
		return domain.EstimatedHours{}, err
	}

	now := deps.Now()

	switch input.Action {
	case "approve":
		if err := entry.Approve(input.ReviewerID, input.AdjustedHours, input.ReviewNote, now); err != nil {
			return domain.EstimatedHours{}, err
		}
	case "reject":
		if err := entry.Reject(input.ReviewerID, input.ReviewNote, now); err != nil {
			return domain.EstimatedHours{}, err
		}
	default:
		return domain.EstimatedHours{}, domain.ErrInvalidStatus
	}

	if err := deps.EstimatedHoursStore.Save(ctx, entry); err != nil {
		return domain.EstimatedHours{}, err
	}

	return entry, nil
}
