package estimatedhours

import (
	"errors"
	"math"
	"time"
)

// Max length constants for user-editable fields.
const (
	MaxNoteLength = 500
)

// Source constants.
const (
	SourceEstimate     = "estimate"      // added by coach/admin
	SourceSelfEstimate = "self_estimate" // submitted by member (§3.5)
	SourceCredit       = "credit"        // direct mat hours credit from admin (§4.7)
)

// Status constants.
const (
	StatusApproved = "approved"
	StatusPending  = "pending"
	StatusRejected = "rejected"
)

// Domain errors.
var (
	ErrEmptyMemberID      = errors.New("member ID cannot be empty")
	ErrEmptyStartDate     = errors.New("start date cannot be empty")
	ErrEmptyEndDate       = errors.New("end date cannot be empty")
	ErrStartAfterEnd      = errors.New("start date cannot be after end date")
	ErrInvalidWeekly      = errors.New("weekly hours must be greater than zero")
	ErrInvalidSource      = errors.New("source must be 'estimate' or 'self_estimate'")
	ErrInvalidStatus      = errors.New("status must be 'approved', 'pending', or 'rejected'")
	ErrNoteTooLong        = errors.New("note cannot exceed 500 characters")
	ErrEmptyCreatedBy     = errors.New("created by cannot be empty")
	ErrWeeklyHoursTooHigh = errors.New("weekly hours cannot exceed 40")
	ErrNotPending         = errors.New("only pending estimates can be reviewed")
	ErrReviewNoteTooLong  = errors.New("review note cannot exceed 500 characters")
)

// EstimatedHours represents a bulk-estimated mat hours entry for a member.
// PRE: MemberID and dates are non-empty, WeeklyHours > 0.
// INVARIANT: TotalHours = ceil(weeks between dates) × WeeklyHours.
type EstimatedHours struct {
	ID          string
	MemberID    string
	StartDate   string  // YYYY-MM-DD
	EndDate     string  // YYYY-MM-DD
	WeeklyHours float64 // hours per week
	TotalHours  float64 // computed total
	Source      string  // estimate or self_estimate
	Status      string  // approved, pending, or rejected
	Note        string
	CreatedBy   string // account ID
	CreatedAt   time.Time
	ReviewedBy  string // account ID of reviewer (admin/coach)
	ReviewedAt  time.Time
	ReviewNote  string // reason for rejection or adjustment note
}

// Validate checks the estimated hours invariants.
// PRE: none
// POST: returns nil if valid, error describing the first violation otherwise
func (e *EstimatedHours) Validate() error {
	if e.MemberID == "" {
		return ErrEmptyMemberID
	}
	if e.StartDate == "" {
		return ErrEmptyStartDate
	}
	if e.EndDate == "" {
		return ErrEmptyEndDate
	}
	start, err := time.Parse("2006-01-02", e.StartDate)
	if err != nil {
		return errors.New("invalid start date format (use YYYY-MM-DD)")
	}
	end, err := time.Parse("2006-01-02", e.EndDate)
	if err != nil {
		return errors.New("invalid end date format (use YYYY-MM-DD)")
	}
	if start.After(end) {
		return ErrStartAfterEnd
	}
	if e.WeeklyHours <= 0 {
		return ErrInvalidWeekly
	}
	if e.WeeklyHours > 40 {
		return ErrWeeklyHoursTooHigh
	}
	if e.Source != SourceEstimate && e.Source != SourceSelfEstimate && e.Source != SourceCredit {
		return ErrInvalidSource
	}
	if e.Status != StatusApproved && e.Status != StatusPending && e.Status != StatusRejected {
		return ErrInvalidStatus
	}
	if len(e.Note) > MaxNoteLength {
		return ErrNoteTooLong
	}
	if e.CreatedBy == "" {
		return ErrEmptyCreatedBy
	}
	return nil
}

// Approve marks the estimate as approved, optionally adjusting the total hours.
// PRE: Status must be pending. reviewerID is non-empty. adjustedHours > 0 if provided, else keeps original.
// POST: Status becomes approved, ReviewedBy/ReviewedAt set.
func (e *EstimatedHours) Approve(reviewerID string, adjustedHours float64, reviewNote string, now time.Time) error {
	if e.Status != StatusPending {
		return ErrNotPending
	}
	if len(reviewNote) > MaxNoteLength {
		return ErrReviewNoteTooLong
	}
	if adjustedHours > 0 {
		e.TotalHours = adjustedHours
	}
	e.Status = StatusApproved
	e.ReviewedBy = reviewerID
	e.ReviewedAt = now
	e.ReviewNote = reviewNote
	return nil
}

// Reject marks the estimate as rejected with a reason.
// PRE: Status must be pending. reviewerID is non-empty.
// POST: Status becomes rejected, ReviewedBy/ReviewedAt/ReviewNote set.
func (e *EstimatedHours) Reject(reviewerID string, reason string, now time.Time) error {
	if e.Status != StatusPending {
		return ErrNotPending
	}
	if len(reason) > MaxNoteLength {
		return ErrReviewNoteTooLong
	}
	e.Status = StatusRejected
	e.ReviewedBy = reviewerID
	e.ReviewedAt = now
	e.ReviewNote = reason
	return nil
}

// CalculateTotalHours computes the total hours from the date range and weekly hours.
// PRE: StartDate and EndDate are valid YYYY-MM-DD dates, WeeklyHours > 0
// POST: sets TotalHours = ceil(weeks) × WeeklyHours
func (e *EstimatedHours) CalculateTotalHours() error {
	start, err := time.Parse("2006-01-02", e.StartDate)
	if err != nil {
		return err
	}
	end, err := time.Parse("2006-01-02", e.EndDate)
	if err != nil {
		return err
	}
	days := end.Sub(start).Hours() / 24
	weeks := math.Ceil(days / 7)
	if weeks < 1 {
		weeks = 1
	}
	e.TotalHours = weeks * e.WeeklyHours
	return nil
}
