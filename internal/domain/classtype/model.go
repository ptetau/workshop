package classtype

import (
	"errors"
	"fmt"
	"strings"
)

// Domain errors
var (
	ErrEmptyName      = errors.New("class type name cannot be empty")
	ErrEmptyProgramID = errors.New("program ID cannot be empty")
	ErrInvalidAttire  = errors.New("attire must be 'gi', 'nogi', or 'both'")
)

// Attire constants.
const (
	AttireGi   = "gi"
	AttireNoGi = "nogi"
	AttireBoth = "both"
)

// Max length constants.
const (
	MaxNameLength        = 200
	MaxLevelLength       = 100
	MaxDescriptionLength = 2000
)

// ClassType represents a specific class within a program (e.g. Fundamentals, No-Gi, Competition).
type ClassType struct {
	ID        string
	ProgramID string
	Name      string

	// Optional metadata for timetable display and filtering.
	Description string // optional, markdown/plain text
	Attire      string // "gi", "nogi", or "both" (optional)
	Level       string // optional free-form label (e.g. Beginner, All-levels)
}

// Validate checks if the ClassType has valid data.
// PRE: ClassType struct is populated
// POST: Returns nil if valid, error otherwise
func (c *ClassType) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return ErrEmptyName
	}
	if len(c.Name) > MaxNameLength {
		return fmt.Errorf("class type name cannot exceed %d characters", MaxNameLength)
	}
	if strings.TrimSpace(c.ProgramID) == "" {
		return ErrEmptyProgramID
	}
	if len(c.Level) > MaxLevelLength {
		return fmt.Errorf("class type level cannot exceed %d characters", MaxLevelLength)
	}
	if len(c.Description) > MaxDescriptionLength {
		return fmt.Errorf("class type description cannot exceed %d characters", MaxDescriptionLength)
	}
	if strings.TrimSpace(c.Attire) != "" && c.Attire != AttireGi && c.Attire != AttireNoGi && c.Attire != AttireBoth {
		return ErrInvalidAttire
	}
	return nil
}
