package orchestrators

import (
	"context"
	"fmt"
	"time"

	attendanceDomain "workshop/internal/domain/attendance"
	domain "workshop/internal/domain/estimatedhours"
)

// EstimatedHoursStoreForBulkAdd defines the store interface needed by the bulk-add orchestrator.
type EstimatedHoursStoreForBulkAdd interface {
	Save(ctx context.Context, e domain.EstimatedHours) error
}

// OverlapMode constants.
const (
	OverlapModeNone    = ""        // default: no overlap handling
	OverlapModeReplace = "replace" // delete overlapping attendance, save estimate
)

// BulkAddEstimatedHoursInput carries input for the bulk-add orchestrator.
type BulkAddEstimatedHoursInput struct {
	MemberID    string
	StartDate   string
	EndDate     string
	WeeklyHours float64
	Note        string
	CreatedBy   string // account ID of the coach/admin
	OverlapMode string // "" or "replace"
}

// AttendanceStoreForOverlap defines the attendance store methods needed for overlap detection.
type AttendanceStoreForOverlap interface {
	ListByMemberIDAndDateRange(ctx context.Context, memberID string, startDate string, endDate string) ([]attendanceDomain.Attendance, error)
	DeleteByMemberIDAndDateRange(ctx context.Context, memberID string, startDate string, endDate string) (int, error)
}

// BulkAddEstimatedHoursDeps holds dependencies for the bulk-add orchestrator.
type BulkAddEstimatedHoursDeps struct {
	EstimatedHoursStore EstimatedHoursStoreForBulkAdd
	AttendanceStore     AttendanceStoreForOverlap // optional: nil skips overlap check
	GenerateID          func() string
	Now                 func() time.Time
}

// OverlapCheckResult describes overlapping attendance records found in the date range.
type OverlapCheckResult struct {
	HasOverlap     bool    `json:"HasOverlap"`
	OverlapCount   int     `json:"OverlapCount"`
	OverlapHours   float64 `json:"OverlapHours"`
	OverlapSummary string  `json:"OverlapSummary"`
}

// CheckEstimatedHoursOverlap checks for overlapping attendance records in the date range.
// PRE: memberID, startDate, endDate are non-empty; attendanceStore is non-nil
// POST: returns overlap info without modifying any data
func CheckEstimatedHoursOverlap(ctx context.Context, memberID, startDate, endDate string, attendanceStore AttendanceStoreForOverlap) (OverlapCheckResult, error) {
	records, err := attendanceStore.ListByMemberIDAndDateRange(ctx, memberID, startDate, endDate)
	if err != nil {
		return OverlapCheckResult{}, err
	}
	if len(records) == 0 {
		return OverlapCheckResult{}, nil
	}
	var totalHours float64
	for _, r := range records {
		if r.MatHours > 0 {
			totalHours += r.MatHours
		} else {
			totalHours += 1.5 // default estimate for sessions without checkout
		}
	}
	return OverlapCheckResult{
		HasOverlap:   true,
		OverlapCount: len(records),
		OverlapHours: totalHours,
		OverlapSummary: fmt.Sprintf("%d attendance record(s) totalling %.1fh exist in %s to %s",
			len(records), totalHours, startDate, endDate),
	}, nil
}

// ExecuteBulkAddEstimatedHours creates a new estimated hours entry for a member.
// PRE: input fields are populated, deps are valid
// POST: estimated hours entry is persisted with calculated total, source=estimate, status=approved
// If OverlapMode is "replace" and AttendanceStore is provided, overlapping records are deleted first.
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

	// Handle overlap if AttendanceStore is provided
	if deps.AttendanceStore != nil && input.OverlapMode == OverlapModeReplace {
		_, err := deps.AttendanceStore.DeleteByMemberIDAndDateRange(ctx, input.MemberID, input.StartDate, input.EndDate)
		if err != nil {
			return domain.EstimatedHours{}, fmt.Errorf("failed to remove overlapping attendance: %w", err)
		}
	}

	if err := deps.EstimatedHoursStore.Save(ctx, entry); err != nil {
		return domain.EstimatedHours{}, err
	}

	return entry, nil
}
