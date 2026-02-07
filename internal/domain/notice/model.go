package notice

import (
	"errors"
	"time"
)

// Notice types
const (
	TypeSchoolWide    = "school_wide"
	TypeClassSpecific = "class_specific"
	TypeHoliday       = "holiday"
)

// Notice statuses
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

// Domain errors
var (
	ErrEmptyTitle    = errors.New("notice title cannot be empty")
	ErrEmptyContent  = errors.New("notice content cannot be empty")
	ErrInvalidType   = errors.New("notice type must be one of: school_wide, class_specific, holiday")
	ErrInvalidStatus = errors.New("notice status must be one of: draft, published")
)

// ValidTypes contains all valid notice types.
var ValidTypes = []string{TypeSchoolWide, TypeClassSpecific, TypeHoliday}

// ValidStatuses contains all valid notice statuses.
var ValidStatuses = []string{StatusDraft, StatusPublished}

// Notice represents a notification in the system.
// Types: school_wide (general announcements), class_specific (coach reminders),
// holiday (auto-generated from Holiday entries).
type Notice struct {
	ID          string
	Type        string // school_wide, class_specific, holiday
	Status      string // draft, published
	Title       string
	Content     string
	CreatedBy   string // AccountID of creator
	PublishedBy string // AccountID of publisher (empty if draft)
	TargetID    string // ClassType ID for class_specific, Holiday ID for holiday, empty for school_wide
	CreatedAt   time.Time
	PublishedAt time.Time
}

// Validate checks if the Notice has valid data.
// PRE: Notice struct is populated
// POST: Returns nil if valid, error otherwise
func (n *Notice) Validate() error {
	if n.Title == "" {
		return ErrEmptyTitle
	}
	if n.Content == "" {
		return ErrEmptyContent
	}
	if !isValidType(n.Type) {
		return ErrInvalidType
	}
	if !isValidStatus(n.Status) {
		return ErrInvalidStatus
	}
	return nil
}

// IsDraft returns true if the notice is in draft state.
// INVARIANT: Status field is not mutated
func (n *Notice) IsDraft() bool {
	return n.Status == StatusDraft
}

// IsPublished returns true if the notice has been published.
// INVARIANT: Status field is not mutated
func (n *Notice) IsPublished() bool {
	return n.Status == StatusPublished
}

// Publish moves the notice from draft to published.
// PRE: Notice is in draft state, publisherID is non-empty
// POST: Status is published, PublishedBy and PublishedAt are set
func (n *Notice) Publish(publisherID string) error {
	if n.IsPublished() {
		return errors.New("notice is already published")
	}
	if publisherID == "" {
		return errors.New("publisher ID is required")
	}
	n.Status = StatusPublished
	n.PublishedBy = publisherID
	n.PublishedAt = time.Now()
	return nil
}

func isValidType(t string) bool {
	for _, v := range ValidTypes {
		if v == t {
			return true
		}
	}
	return false
}

func isValidStatus(s string) bool {
	for _, v := range ValidStatuses {
		if v == s {
			return true
		}
	}
	return false
}
