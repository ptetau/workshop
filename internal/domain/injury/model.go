package injury

import (
	"errors"
	"time"
)

// Max length constants for user-editable fields.
const (
	MaxDescriptionLength = 1000
)

// Body part constants
const (
	BodyPartKnee     = "knee"
	BodyPartShoulder = "shoulder"
	BodyPartBack     = "back"
	BodyPartNeck     = "neck"
	BodyPartAnkle    = "ankle"
	BodyPartWrist    = "wrist"
	BodyPartRib      = "rib"
	BodyPartOther    = "other"
)

// Injury holds state for the concept.
type Injury struct {
	ID          string
	BodyPart    string
	Description string
	MemberID    string
	ReportedAt  time.Time
}

// Validate checks if the Injury has valid data.
// PRE: Injury struct is initialized
// POST: Returns error if validation fails, nil otherwise
// INVARIANT: MemberID and BodyPart must not be empty
func (i *Injury) Validate() error {
	if i.MemberID == "" {
		return errors.New("injury must be associated with a member")
	}
	if i.BodyPart == "" {
		return errors.New("body part must be specified")
	}
	if len(i.Description) > MaxDescriptionLength {
		return errors.New("injury description cannot exceed 1000 characters")
	}
	if i.ReportedAt.IsZero() {
		return errors.New("reported date must be set")
	}
	return nil
}

// IsActive returns true if the injury is still active (within 7 days).
// PRE: Injury is initialized
// POST: Returns boolean indicating active status
// INVARIANT: Injuries are considered active for 7 days
func (i *Injury) IsActive() bool {
	return time.Since(i.ReportedAt) < 7*24*time.Hour
}

// GetSeverity returns a severity indicator based on the injury.
// PRE: Injury is initialized
// POST: Returns severity string ("high", "medium", "low")
func (i *Injury) GetSeverity() string {
	// Simple heuristic: neck/back are high severity
	if i.BodyPart == BodyPartNeck || i.BodyPart == BodyPartBack {
		return "high"
	}
	// Shoulder/knee are medium
	if i.BodyPart == BodyPartShoulder || i.BodyPart == BodyPartKnee {
		return "medium"
	}
	return "low"
}
