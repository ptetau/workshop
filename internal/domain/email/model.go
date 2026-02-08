package email

import (
	"errors"
	"time"
)

// Status constants for email lifecycle.
const (
	StatusDraft     = "draft"
	StatusQueued    = "queued"
	StatusSent      = "sent"
	StatusScheduled = "scheduled"
	StatusCancelled = "cancelled"
	StatusFailed    = "failed"
)

// Domain errors
var (
	ErrEmptySubject   = errors.New("email subject is required")
	ErrEmptyBody      = errors.New("email body is required")
	ErrEmptySenderID  = errors.New("sender ID is required")
	ErrNoRecipients   = errors.New("at least one recipient is required")
	ErrAlreadySent    = errors.New("email has already been sent")
	ErrNotScheduled   = errors.New("email is not in scheduled status")
	ErrNotDraft       = errors.New("email is not in draft status")
	ErrNotCancellable = errors.New("email cannot be cancelled in its current status")
)

// Email represents a composed email that can be sent via Resend.
type Email struct {
	ID                string
	Subject           string
	Body              string
	SenderID          string // AccountID of the sender
	Status            string
	ScheduledAt       time.Time
	SentAt            time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ResendMessageID   string // Resend API message ID for tracking
	TemplateVersionID string // Snapshot of template used at send time
}

// Recipient links an email to a member.
type Recipient struct {
	EmailID        string
	MemberID       string
	MemberName     string // Denormalized for display
	MemberEmail    string // The actual email address for delivery
	DeliveryStatus string // sent, delivered, bounced, opened (from Resend webhooks)
}

// Validate checks that the Email has valid data.
// PRE: Email struct is populated
// POST: Returns nil if valid, error otherwise
func (e *Email) Validate() error {
	if e.Subject == "" {
		return ErrEmptySubject
	}
	if e.Body == "" {
		return ErrEmptyBody
	}
	if e.SenderID == "" {
		return ErrEmptySenderID
	}
	if e.CreatedAt.IsZero() {
		return errors.New("created_at must be set")
	}
	return nil
}

// IsDraft returns true if the email is in draft status.
// INVARIANT: Status field is not mutated
func (e *Email) IsDraft() bool {
	return e.Status == StatusDraft
}

// IsSent returns true if the email has been sent.
// INVARIANT: Status field is not mutated
func (e *Email) IsSent() bool {
	return e.Status == StatusSent
}

// IsScheduled returns true if the email is scheduled for future delivery.
// INVARIANT: Status field is not mutated
func (e *Email) IsScheduled() bool {
	return e.Status == StatusScheduled
}

// MarkQueued transitions the email to queued status for immediate sending.
// PRE: Email is in draft status
// POST: Status is set to queued
func (e *Email) MarkQueued() error {
	if e.Status != StatusDraft {
		return ErrNotDraft
	}
	e.Status = StatusQueued
	return nil
}

// MarkSent records that the email was successfully sent.
// PRE: Email is in queued status
// POST: Status is sent, SentAt is set
func (e *Email) MarkSent(sentAt time.Time, resendID string) {
	e.Status = StatusSent
	e.SentAt = sentAt
	e.ResendMessageID = resendID
}

// MarkFailed records that sending failed.
// PRE: Email is in queued status
// POST: Status is failed
func (e *Email) MarkFailed() {
	e.Status = StatusFailed
}

// Schedule sets the email for future delivery.
// PRE: Email is in draft status; scheduledAt is in the future
// POST: Status is scheduled, ScheduledAt is set
func (e *Email) Schedule(scheduledAt time.Time) error {
	if e.Status != StatusDraft {
		return ErrNotDraft
	}
	e.Status = StatusScheduled
	e.ScheduledAt = scheduledAt
	return nil
}

// Reschedule changes the scheduled delivery time.
// PRE: Email is in scheduled status
// POST: ScheduledAt is updated
func (e *Email) Reschedule(scheduledAt time.Time) error {
	if e.Status != StatusScheduled {
		return ErrNotScheduled
	}
	e.ScheduledAt = scheduledAt
	return nil
}

// Cancel cancels a scheduled or draft email.
// PRE: Email is in scheduled or draft status
// POST: Status is cancelled
func (e *Email) Cancel() error {
	if e.Status != StatusScheduled && e.Status != StatusDraft {
		return ErrNotCancellable
	}
	e.Status = StatusCancelled
	return nil
}
