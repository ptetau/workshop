package deletion

import (
	"errors"
	"time"
)

// Status constants for deletion request lifecycle.
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusProcessed = "processed"
	StatusCancelled = "cancelled"
	StatusExpired   = "expired"
)

// Domain errors.
var (
	ErrEmptyMemberID      = errors.New("member_id is required")
	ErrEmptyRequestID     = errors.New("request_id is required")
	ErrEmptyEmail         = errors.New("email is required")
	ErrInvalidStatus      = errors.New("invalid status transition")
	ErrGracePeriodExpired = errors.New("grace period has expired")
	ErrAlreadyProcessed   = errors.New("deletion request already processed")
)

// Request represents a member's data deletion request (GDPR Article 17).
type Request struct {
	ID             string
	MemberID       string
	Email          string // For verification email
	Status         string
	RequestedAt    time.Time
	GracePeriodEnd time.Time  // 14 days by default
	ConfirmedAt    *time.Time // When user confirms via email
	ProcessedAt    *time.Time // When data is actually deleted
	CancelledAt    *time.Time // If user cancels during grace period
	IPAddress      string     // Audit trail
	UserAgent      string     // Audit trail
}

// Validate checks that the Request has valid data.
// PRE: Request fields may be empty
// POST: Returns nil if valid, error otherwise
// INVARIANT: ID, MemberID, Email, RequestedAt, GracePeriodEnd must be non-empty
func (r *Request) Validate() error {
	if r.ID == "" {
		return ErrEmptyRequestID
	}
	if r.MemberID == "" {
		return ErrEmptyMemberID
	}
	if r.Email == "" {
		return ErrEmptyEmail
	}
	if r.RequestedAt.IsZero() {
		return errors.New("requested_at must be set")
	}
	if r.GracePeriodEnd.IsZero() {
		return errors.New("grace_period_end must be set")
	}
	return nil
}

// CanConfirm returns true if the request can be confirmed via email.
// PRE: Request status is known
// POST: Returns true if pending and within grace period
// INVARIANT: Status must be pending, time must be before grace period end
func (r *Request) CanConfirm() bool {
	return r.Status == StatusPending && time.Now().Before(r.GracePeriodEnd)
}

// CanCancel returns true if the user can cancel the deletion.
// PRE: Request status is known
// POST: Returns true if pending/confirmed and within grace period and not processed
// INVARIANT: Status must be pending or confirmed, time before grace period end, not processed
func (r *Request) CanCancel() bool {
	return (r.Status == StatusPending || r.Status == StatusConfirmed) &&
		time.Now().Before(r.GracePeriodEnd) &&
		r.ProcessedAt == nil
}

// CanProcess returns true if the grace period has passed and deletion can proceed.
// PRE: Request status is known
// POST: Returns true if confirmed and grace period passed and not processed
// INVARIANT: Status must be confirmed, time after grace period end, not processed
func (r *Request) CanProcess() bool {
	return r.Status == StatusConfirmed &&
		time.Now().After(r.GracePeriodEnd) &&
		r.ProcessedAt == nil
}

// IsExpired returns true if the grace period passed without confirmation.
// PRE: Request status is known
// POST: Returns true if pending and grace period passed
// INVARIANT: Status must be pending, time after grace period end
func (r *Request) IsExpired() bool {
	return r.Status == StatusPending && time.Now().After(r.GracePeriodEnd)
}

// IsTerminal returns true if the request reached a final state.
// PRE: Request status is known
// POST: Returns true if processed, cancelled, or expired
// INVARIANT: Status must be one of terminal states
func (r *Request) IsTerminal() bool {
	return r.Status == StatusProcessed || r.Status == StatusCancelled || r.Status == StatusExpired
}

// MarkConfirmed records email confirmation of deletion.
// PRE: CanConfirm returns true
// POST: Status set to confirmed, ConfirmedAt set to now
// INVARIANT: Request must be in pending status, within grace period
func (r *Request) MarkConfirmed() error {
	if !r.CanConfirm() {
		return ErrInvalidStatus
	}
	now := time.Now()
	r.Status = StatusConfirmed
	r.ConfirmedAt = &now
	return nil
}

// MarkCancelled records cancellation during grace period.
// PRE: CanCancel returns true
// POST: Status set to cancelled, CancelledAt set to now
// INVARIANT: Request must be cancellable
func (r *Request) MarkCancelled() error {
	if !r.CanCancel() {
		return ErrInvalidStatus
	}
	now := time.Now()
	r.Status = StatusCancelled
	r.CancelledAt = &now
	return nil
}

// MarkProcessed records that data has been deleted.
// PRE: CanProcess returns true
// POST: Status set to processed, ProcessedAt set to now
// INVARIANT: Request must be confirmed and grace period passed
func (r *Request) MarkProcessed() error {
	if !r.CanProcess() {
		if r.ProcessedAt != nil {
			return ErrAlreadyProcessed
		}
		if r.Status != StatusConfirmed {
			return ErrInvalidStatus
		}
		return ErrGracePeriodExpired
	}
	now := time.Now()
	r.Status = StatusProcessed
	r.ProcessedAt = &now
	return nil
}

// MarkExpired marks the request as expired if grace period passed.
// PRE: IsExpired returns true
// POST: Status set to expired
// INVARIANT: Request must be pending and grace period passed
func (r *Request) MarkExpired() error {
	if !r.IsExpired() {
		return errors.New("grace period has not expired yet")
	}
	r.Status = StatusExpired
	return nil
}

// RemainingGracePeriod returns time until grace period ends (can be negative).
// PRE: GracePeriodEnd is set
// POST: Returns duration until grace period end
// INVARIANT: None
func (r *Request) RemainingGracePeriod() time.Duration {
	return time.Until(r.GracePeriodEnd)
}

// NewRequest creates a new deletion request with standard 14-day grace period.
func NewRequest(id, memberID, email, ipAddress, userAgent string) *Request {
	now := time.Now()
	return &Request{
		ID:             id,
		MemberID:       memberID,
		Email:          email,
		Status:         StatusPending,
		RequestedAt:    now,
		GracePeriodEnd: now.Add(14 * 24 * time.Hour), // 14 days
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
	}
}
