package orchestrators

import (
	"context"
	"time"

	domain "workshop/internal/domain/estimatedhours"
)

// EstimatedHoursStoreForBulkAdd defines the store interface needed by the bulk-add orchestrator.
type EstimatedHoursStoreForBulkAdd interface {
	Save(ctx context.Context, e domain.EstimatedHours) error
}

// BulkAddEstimatedHoursInput carries input for the bulk-add orchestrator.
type BulkAddEstimatedHoursInput struct {
	MemberID    string
	StartDate   string
	EndDate     string
	WeeklyHours float64
	Note        string
	CreatedBy   string // account ID of the coach/admin
}

// BulkAddEstimatedHoursDeps holds dependencies for the bulk-add orchestrator.
type BulkAddEstimatedHoursDeps struct {
	EstimatedHoursStore EstimatedHoursStoreForBulkAdd
	GenerateID          func() string
	Now                 func() time.Time
}

// ExecuteBulkAddEstimatedHours creates a new estimated hours entry for a member.
// PRE: input fields are populated, deps are valid
// POST: estimated hours entry is persisted with calculated total, source=estimate, status=approved
func ExecuteBulkAddEstimatedHours(ctx context.Context, input BulkAddEstimatedHoursInput, deps BulkAddEstimatedHoursDeps) (domain.EstimatedHours, error) {
	entry := domain.EstimatedHours{
		ID:          deps.GenerateID(),
		MemberID:    input.MemberID,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		WeeklyHours: input.WeeklyHours,
		Source:      domain.SourceEstimate,
		Status:      domain.StatusApproved,
		Note:        input.Note,
		CreatedBy:   input.CreatedBy,
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
