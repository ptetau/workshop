package classtype

import (
	"errors"
	"strings"
)

// Domain errors
var (
	ErrEmptyName      = errors.New("class type name cannot be empty")
	ErrEmptyProgramID = errors.New("program ID cannot be empty")
)

// ClassType represents a specific class within a program (e.g. Fundamentals, No-Gi, Competition).
type ClassType struct {
	ID        string
	ProgramID string
	Name      string
}

// Validate checks if the ClassType has valid data.
// PRE: ClassType struct is populated
// POST: Returns nil if valid, error otherwise
func (c *ClassType) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return ErrEmptyName
	}
	if strings.TrimSpace(c.ProgramID) == "" {
		return ErrEmptyProgramID
	}
	return nil
}
