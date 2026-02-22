package clip

import (
	"errors"
	"time"
)

// MaxClipsInComparison is the maximum number of clips in a 4-Up comparison.
const MaxClipsInComparison = 4

// ComparisonSession represents a 4-Up comparison session with up to 4 clips.
type ComparisonSession struct {
	ID        string
	Name      string   // optional name for the comparison
	ClipIDs   []string // up to 4 clip IDs
	CreatedBy string
	CreatedAt time.Time
}

// Validate checks the comparison session invariants.
// PRE: none
// POST: returns nil if valid, error describing first violation otherwise
func (c *ComparisonSession) Validate() error {
	if len(c.ClipIDs) == 0 {
		return errors.New("comparison must have at least one clip")
	}
	if len(c.ClipIDs) > MaxClipsInComparison {
		return errors.New("comparison cannot exceed 4 clips")
	}
	// Check for duplicate clip IDs
	seen := make(map[string]bool)
	for _, id := range c.ClipIDs {
		if seen[id] {
			return errors.New("duplicate clip in comparison")
		}
		seen[id] = true
	}
	return nil
}

// AddClip adds a clip to the comparison if there's room.
// PRE: clipID is non-empty
// POST: clip added to ClipIDs if valid; error if full or duplicate
func (c *ComparisonSession) AddClip(clipID string) error {
	if len(c.ClipIDs) >= MaxClipsInComparison {
		return errors.New("comparison is full (max 4 clips)")
	}
	for _, id := range c.ClipIDs {
		if id == clipID {
			return errors.New("clip already in comparison")
		}
	}
	c.ClipIDs = append(c.ClipIDs, clipID)
	return nil
}

// RemoveClip removes a clip from the comparison.
// PRE: clipID is non-empty
// POST: clip removed from ClipIDs if found; error otherwise
func (c *ComparisonSession) RemoveClip(clipID string) error {
	for i, id := range c.ClipIDs {
		if id == clipID {
			c.ClipIDs = append(c.ClipIDs[:i], c.ClipIDs[i+1:]...)
			return nil
		}
	}
	return errors.New("clip not found in comparison")
}

// ResearchNote represents private research notes for a comparison session.
type ResearchNote struct {
	ID                  string
	ComparisonSessionID string
	Content             string
	CreatedBy           string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// MaxResearchNoteLength is the maximum length for research notes.
const MaxResearchNoteLength = 5000

// Validate checks the research note invariants.
// PRE: none
// POST: returns nil if valid, error describing violation otherwise
func (r *ResearchNote) Validate() error {
	if r.ComparisonSessionID == "" {
		return errors.New("comparison session ID is required")
	}
	if r.Content == "" {
		return errors.New("content cannot be empty")
	}
	if len(r.Content) > MaxResearchNoteLength {
		return errors.New("content exceeds maximum length")
	}
	return nil
}
