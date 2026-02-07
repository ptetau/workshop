package message

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrEmptySenderID   = errors.New("sender ID is required")
	ErrEmptyReceiverID = errors.New("receiver ID (member) is required")
	ErrEmptyContent    = errors.New("message content cannot be empty")
)

// Message represents a direct in-app message from Admin to a member.
type Message struct {
	ID         string
	SenderID   string // Admin AccountID
	ReceiverID string // Member ID
	Subject    string
	Content    string
	ReadAt     time.Time
	CreatedAt  time.Time
}

// Validate checks if the Message has valid data.
// PRE: Message struct is populated
// POST: Returns nil if valid, error otherwise
func (m *Message) Validate() error {
	if m.SenderID == "" {
		return ErrEmptySenderID
	}
	if m.ReceiverID == "" {
		return ErrEmptyReceiverID
	}
	if m.Content == "" {
		return ErrEmptyContent
	}
	if m.CreatedAt.IsZero() {
		return errors.New("created_at must be set")
	}
	return nil
}

// IsRead returns true if the message has been read.
// INVARIANT: ReadAt field is not mutated
func (m *Message) IsRead() bool {
	return !m.ReadAt.IsZero()
}

// MarkRead records when the message was read.
// PRE: Message exists
// POST: ReadAt is set to current time if previously zero
func (m *Message) MarkRead() {
	if m.ReadAt.IsZero() {
		m.ReadAt = time.Now()
	}
}
