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

// Color presets — 7 predefined highlight colours for notices.
const (
	ColorOrange = "orange" // #F9B232 — default
	ColorRed    = "red"    // #e74c3c
	ColorGreen  = "green"  // #27ae60
	ColorBlue   = "blue"   // #2980b9
	ColorPurple = "purple" // #8e44ad
	ColorTeal   = "teal"   // #16a085
	ColorGrey   = "grey"   // #7f8c8d
)

// ColorHex maps preset names to hex values.
var ColorHex = map[string]string{
	ColorOrange: "#F9B232",
	ColorRed:    "#e74c3c",
	ColorGreen:  "#27ae60",
	ColorBlue:   "#2980b9",
	ColorPurple: "#8e44ad",
	ColorTeal:   "#16a085",
	ColorGrey:   "#7f8c8d",
}

// ValidColors contains all valid colour preset names.
var ValidColors = []string{ColorOrange, ColorRed, ColorGreen, ColorBlue, ColorPurple, ColorTeal, ColorGrey}

// Domain errors
var (
	ErrEmptyTitle    = errors.New("notice title cannot be empty")
	ErrEmptyContent  = errors.New("notice content cannot be empty")
	ErrInvalidType   = errors.New("notice type must be one of: school_wide, class_specific, holiday")
	ErrInvalidStatus = errors.New("notice status must be one of: draft, published")
	ErrInvalidColor  = errors.New("notice color must be one of: orange, red, green, blue, purple, teal, grey")
	ErrAlreadyPinned = errors.New("notice is already pinned")
	ErrNotPinned     = errors.New("notice is not pinned")
)

// ValidTypes contains all valid notice types.
var ValidTypes = []string{TypeSchoolWide, TypeClassSpecific, TypeHoliday}

// ValidStatuses contains all valid notice statuses.
var ValidStatuses = []string{StatusDraft, StatusPublished}

// Notice represents a notification in the system.
// Types: school_wide (general announcements), class_specific (coach reminders),
// holiday (auto-generated from Holiday entries).
// Content supports Markdown formatting.
type Notice struct {
	ID           string
	Type         string // school_wide, class_specific, holiday
	Status       string // draft, published
	Title        string
	Content      string // Markdown content
	CreatedBy    string // AccountID of creator
	PublishedBy  string // AccountID of publisher (empty if draft)
	TargetID     string // ClassType ID for class_specific, Holiday ID for holiday, empty for school_wide
	AuthorName   string // Display name of the author
	ShowAuthor   bool   // Whether to show author name when displayed
	Color        string // Highlight colour preset (orange, red, green, blue, purple, teal, grey)
	Pinned       bool   // Whether pinned to top of notice list
	PinnedAt     time.Time
	VisibleFrom  time.Time // Scheduled appearance (zero = immediately)
	VisibleUntil time.Time // Scheduled disappearance (zero = indefinite)
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PublishedAt  time.Time
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
	if n.Color != "" && !isValidColor(n.Color) {
		return ErrInvalidColor
	}
	return nil
}

// EffectiveColor returns the color hex value, defaulting to orange.
func (n *Notice) EffectiveColor() string {
	if n.Color == "" {
		return ColorHex[ColorOrange]
	}
	if hex, ok := ColorHex[n.Color]; ok {
		return hex
	}
	return ColorHex[ColorOrange]
}

// IsVisible returns true if the notice is currently visible based on the scheduled window.
// PRE: now is the current time in UTC
// POST: Returns true if the notice falls within its visibility window
func (n *Notice) IsVisible(now time.Time) bool {
	if !n.VisibleFrom.IsZero() && now.Before(n.VisibleFrom) {
		return false
	}
	if !n.VisibleUntil.IsZero() && now.After(n.VisibleUntil) {
		return false
	}
	return true
}

// Pin marks the notice as pinned.
// PRE: Notice is not already pinned
// POST: Pinned is true, PinnedAt is set
func (n *Notice) Pin(now time.Time) error {
	if n.Pinned {
		return ErrAlreadyPinned
	}
	n.Pinned = true
	n.PinnedAt = now
	return nil
}

// Unpin removes the pinned status.
// PRE: Notice is pinned
// POST: Pinned is false, PinnedAt is zeroed
func (n *Notice) Unpin() error {
	if !n.Pinned {
		return ErrNotPinned
	}
	n.Pinned = false
	n.PinnedAt = time.Time{}
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

func isValidColor(c string) bool {
	for _, v := range ValidColors {
		if v == c {
			return true
		}
	}
	return false
}
