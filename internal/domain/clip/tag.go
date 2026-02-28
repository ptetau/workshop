package clip

import (
	"errors"
	"time"
)

// MaxTagLength is the maximum length for a tag name.
const MaxTagLength = 50

// Tag represents a label that can be applied to clips for organization and search.
type Tag struct {
	ID        string
	Name      string
	CreatedBy string
	CreatedAt time.Time
}

// Validate checks the tag's invariants.
// PRE: none
// POST: returns nil if valid, error describing first violation otherwise
func (t *Tag) Validate() error {
	if t.Name == "" {
		return errors.New("tag name cannot be empty")
	}
	if len(t.Name) > MaxTagLength {
		return errors.New("tag name cannot exceed 50 characters")
	}
	return nil
}

// ClipTag represents the many-to-many relationship between clips and tags.
type ClipTag struct {
	ClipID    string
	TagID     string
	CreatedBy string
	CreatedAt time.Time
}
