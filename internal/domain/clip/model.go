package clip

import (
	"errors"
	"regexp"
	"time"
)

// Max length constants for user-editable fields.
const (
	MaxTitleLength      = 200
	MaxYouTubeURLLength = 2048
	MaxNotesLength      = 1000
)

// Clip represents a YouTube timestamp loop for technical study.
// PRE: YouTubeURL is a valid YouTube URL. StartSeconds < EndSeconds.
// INVARIANT: A clip always belongs to a theme via ThemeID.
type Clip struct {
	ID           string
	ThemeID      string // references Theme by ID
	Title        string // descriptive title for the clip
	YouTubeURL   string // full YouTube URL
	YouTubeID    string // extracted video ID
	StartSeconds int    // loop start in seconds
	EndSeconds   int    // loop end in seconds
	Notes        string // optional technique notes
	CreatedBy    string // account ID of the creator
	Promoted     bool   // true if promoted to the main library by coach/admin
	PromotedBy   string // account ID who promoted it
	CreatedAt    time.Time
}

var youtubeIDRegex = regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/embed/)([a-zA-Z0-9_-]{11})`)

// Validate checks the clip's invariants.
// PRE: none
// POST: returns nil if valid, error describing the first violation otherwise
func (c *Clip) Validate() error {
	if c.ThemeID == "" {
		return errors.New("clip theme ID cannot be empty")
	}
	if c.Title == "" {
		return errors.New("clip title cannot be empty")
	}
	if len(c.Title) > MaxTitleLength {
		return errors.New("clip title cannot exceed 200 characters")
	}
	if c.YouTubeURL == "" {
		return errors.New("clip YouTube URL cannot be empty")
	}
	if len(c.YouTubeURL) > MaxYouTubeURLLength {
		return errors.New("clip YouTube URL cannot exceed 2048 characters")
	}
	if c.StartSeconds < 0 {
		return errors.New("clip start seconds cannot be negative")
	}
	if c.EndSeconds <= 0 {
		return errors.New("clip end seconds must be positive")
	}
	if c.StartSeconds >= c.EndSeconds {
		return errors.New("clip start must be before end")
	}
	if len(c.Notes) > MaxNotesLength {
		return errors.New("clip notes cannot exceed 1000 characters")
	}
	return nil
}

// ExtractYouTubeID parses the YouTube video ID from the URL.
// PRE: YouTubeURL is set
// POST: sets YouTubeID if a valid ID is found, returns error otherwise
func (c *Clip) ExtractYouTubeID() error {
	matches := youtubeIDRegex.FindStringSubmatch(c.YouTubeURL)
	if len(matches) < 2 {
		return errors.New("could not extract YouTube video ID from URL")
	}
	c.YouTubeID = matches[1]
	return nil
}

// DurationSeconds returns the loop duration in seconds.
// PRE: StartSeconds < EndSeconds
// POST: returns positive duration
func (c *Clip) DurationSeconds() int {
	return c.EndSeconds - c.StartSeconds
}

// Promote marks the clip as promoted to the main library.
// PRE: promotedBy is a valid account ID
// POST: Promoted is true, PromotedBy is set
func (c *Clip) Promote(promotedBy string) error {
	if promotedBy == "" {
		return errors.New("promoted by account ID cannot be empty")
	}
	if c.Promoted {
		return errors.New("clip is already promoted")
	}
	c.Promoted = true
	c.PromotedBy = promotedBy
	return nil
}
