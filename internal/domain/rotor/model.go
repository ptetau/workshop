package rotor

import (
	"errors"
	"strings"
	"time"
)

// Rotor status constants.
const (
	StatusDraft    = "draft"
	StatusActive   = "active"
	StatusArchived = "archived"
)

// Max length constants for user-editable fields.
const (
	MaxRotorNameLength        = 100
	MaxThemeNameLength        = 100
	MaxTopicNameLength        = 100
	MaxTopicDescriptionLength = 500
)

// Domain errors.
var (
	ErrEmptyName        = errors.New("rotor name cannot be empty")
	ErrEmptyClassTypeID = errors.New("class type ID cannot be empty")
	ErrEmptyCreatedBy   = errors.New("created by cannot be empty")
	ErrInvalidStatus    = errors.New("invalid rotor status")
	ErrNotDraft         = errors.New("rotor must be in draft status to modify")
	ErrAlreadyActive    = errors.New("rotor is already active")
	ErrCannotArchive    = errors.New("only active rotors can be archived")

	ErrEmptyThemeName    = errors.New("theme name cannot be empty")
	ErrEmptyRotorID      = errors.New("rotor ID cannot be empty")
	ErrEmptyTopicName    = errors.New("topic name cannot be empty")
	ErrEmptyRotorThemeID = errors.New("rotor theme ID cannot be empty")
	ErrInvalidDuration   = errors.New("duration must be at least 1 week")
	ErrTopicNotScheduled = errors.New("topic is not currently scheduled")
	ErrAlreadyVoted      = errors.New("already voted for this topic in current cycle")

	ErrRotorNameTooLong        = errors.New("rotor name cannot exceed 100 characters")
	ErrThemeNameTooLong        = errors.New("theme name cannot exceed 100 characters")
	ErrTopicNameTooLong        = errors.New("topic name cannot exceed 100 characters")
	ErrTopicDescriptionTooLong = errors.New("topic description cannot exceed 500 characters")
)

// Rotor represents a versioned curriculum structure for a class type.
// PRE: ClassTypeID and Name are non-empty.
// INVARIANT: Only one rotor per ClassTypeID can be active at a time.
type Rotor struct {
	ID          string
	ClassTypeID string
	Name        string
	Version     int
	Status      string // draft, active, archived
	PreviewOn   bool   // whether members can see upcoming topics
	CreatedBy   string // account ID
	CreatedAt   time.Time
	ActivatedAt time.Time
}

// Validate checks the rotor's invariants.
// PRE: none
// POST: returns nil if valid, error describing the first violation otherwise
func (r *Rotor) Validate() error {
	if r.Name == "" {
		return ErrEmptyName
	}
	if len(r.Name) > MaxRotorNameLength {
		return ErrRotorNameTooLong
	}
	if r.ClassTypeID == "" {
		return ErrEmptyClassTypeID
	}
	if r.CreatedBy == "" {
		return ErrEmptyCreatedBy
	}
	if r.Status != StatusDraft && r.Status != StatusActive && r.Status != StatusArchived {
		return ErrInvalidStatus
	}
	return nil
}

// IsDraft returns true if the rotor is in draft status.
// PRE: none
// POST: returns true if Status == StatusDraft
func (r *Rotor) IsDraft() bool {
	return r.Status == StatusDraft
}

// IsActive returns true if the rotor is in active status.
// PRE: none
// POST: returns true if Status == StatusActive
func (r *Rotor) IsActive() bool {
	return r.Status == StatusActive
}

// Activate transitions the rotor from draft to active.
// PRE: Rotor must be in draft status.
// POST: Status becomes active, ActivatedAt is set.
func (r *Rotor) Activate(now time.Time) error {
	if r.Status == StatusActive {
		return ErrAlreadyActive
	}
	if r.Status != StatusDraft {
		return ErrNotDraft
	}
	r.Status = StatusActive
	r.ActivatedAt = now
	return nil
}

// Rename changes the rotor's name.
// PRE: name is non-empty and within MaxRotorNameLength.
// POST: Name is updated.
func (r *Rotor) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > MaxRotorNameLength {
		return ErrRotorNameTooLong
	}
	r.Name = name
	return nil
}

// Archive transitions the rotor from active to archived.
// PRE: Rotor must be in active status.
// POST: Status becomes archived.
func (r *Rotor) Archive() error {
	if r.Status != StatusActive {
		return ErrCannotArchive
	}
	r.Status = StatusArchived
	return nil
}

// RotorTheme represents a concurrent theme category within a rotor.
// PRE: Name and RotorID are non-empty.
// INVARIANT: Position determines display order within the rotor.
type RotorTheme struct {
	ID       string
	RotorID  string
	Name     string // e.g. "Standing", "Guard", "Pinning"
	Position int    // display order (0-indexed)
}

// Validate checks the theme's invariants.
// PRE: none
// POST: returns nil if valid, error describing the first violation otherwise
func (t *RotorTheme) Validate() error {
	if t.Name == "" {
		return ErrEmptyThemeName
	}
	if len(t.Name) > MaxThemeNameLength {
		return ErrThemeNameTooLong
	}
	if t.RotorID == "" {
		return ErrEmptyRotorID
	}
	return nil
}

// Topic represents a single technique/drill in a theme's queue.
// PRE: Name and RotorThemeID are non-empty.
// INVARIANT: Position determines order in the queue.
type Topic struct {
	ID            string
	RotorThemeID  string
	Name          string // e.g. "Single Leg Takedown"
	Description   string // optional coach notes
	DurationWeeks int    // how many weeks this topic runs (default 1)
	Position      int    // order in the queue (0-indexed)
	LastCovered   time.Time
}

// Validate checks the topic's invariants.
// PRE: none
// POST: returns nil if valid, error describing the first violation otherwise
func (t *Topic) Validate() error {
	if t.Name == "" {
		return ErrEmptyTopicName
	}
	if len(t.Name) > MaxTopicNameLength {
		return ErrTopicNameTooLong
	}
	if t.RotorThemeID == "" {
		return ErrEmptyRotorThemeID
	}
	if t.DurationWeeks < 1 {
		return ErrInvalidDuration
	}
	if len(t.Description) > MaxTopicDescriptionLength {
		return ErrTopicDescriptionTooLong
	}
	return nil
}

// TopicSchedule status constants.
const (
	ScheduleStatusScheduled = "scheduled"
	ScheduleStatusActive    = "active"
	ScheduleStatusCompleted = "completed"
	ScheduleStatusSkipped   = "skipped"
)

// TopicSchedule represents a topic's scheduled time slot.
// PRE: TopicID and RotorThemeID are non-empty, StartDate is set.
// INVARIANT: Only one topic per theme can be active at a time.
type TopicSchedule struct {
	ID           string
	TopicID      string
	RotorThemeID string
	StartDate    time.Time
	EndDate      time.Time
	Status       string // scheduled, active, completed, skipped
}

// IsActive returns true if the schedule entry is currently active.
// PRE: now is a valid time
// POST: returns true if status is active and now is within the date range
func (s *TopicSchedule) IsActive(now time.Time) bool {
	return s.Status == ScheduleStatusActive &&
		!now.Before(s.StartDate) &&
		(s.EndDate.IsZero() || !now.After(s.EndDate))
}

// NextTopicInQueue returns the next topic in position order after currentTopicID,
// wrapping around to the first topic when the end of the queue is reached.
// PRE: topics is sorted by Position ascending, currentTopicID is non-empty.
// POST: returns the next topic or nil if the queue is empty or has only one topic.
func NextTopicInQueue(topics []Topic, currentTopicID string) *Topic {
	if len(topics) <= 1 {
		return nil
	}
	for i, t := range topics {
		if t.ID == currentTopicID {
			next := (i + 1) % len(topics)
			return &topics[next]
		}
	}
	// currentTopicID not found — return first topic as fallback
	return &topics[0]
}

// Vote represents a member's vote for a topic.
// PRE: TopicID and AccountID are non-empty.
// INVARIANT: One vote per member per topic per rotation cycle.
type Vote struct {
	ID        string
	TopicID   string
	AccountID string
	CreatedAt time.Time
}
