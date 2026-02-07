package program

import (
	"errors"
	"strings"
)

// Program type constants
const (
	TypeAdults = "adults"
	TypeKids   = "kids"
)

// ValidTypes contains all valid program types.
var ValidTypes = []string{TypeAdults, TypeKids}

// Domain errors
var (
	ErrEmptyName   = errors.New("program name cannot be empty")
	ErrInvalidType = errors.New("program type must be 'adults' or 'kids'")
)

// Program represents a top-level training program (e.g. Adults, Kids).
type Program struct {
	ID   string
	Name string
	Type string // "adults" or "kids"
}

// Validate checks if the Program has valid data.
// PRE: Program struct is populated
// POST: Returns nil if valid, error otherwise
func (p *Program) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrEmptyName
	}
	if !isValidType(p.Type) {
		return ErrInvalidType
	}
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
