package bugbox

import (
	"errors"
	"time"
)

// Submission represents a bug report submitted via the Bug Box.
// PRE: Summary and Description are non-empty; Route is the page path at time of submission.
// POST: A GitHub issue is created and linked via GitHubIssueNumber/GitHubIssueURL.
// INVARIANT: Submissions never contain cookies, CSRF tokens, passwords, or raw session data.
type Submission struct {
	ID                string
	Summary           string
	Description       string
	Steps             string
	Expected          string
	Actual            string
	Route             string
	UserAgent         string
	Viewport          string
	Role              string
	ImpersonatedRole  string
	SubmittedAt       time.Time
	ScreenshotPath    string
	GitHubIssueNumber int
	GitHubIssueURL    string
}

// Validate checks that the required fields are present.
// PRE: none
// POST: returns error if Summary or Description is empty
func (s *Submission) Validate() error {
	if s.Summary == "" {
		return errors.New("summary is required")
	}
	if s.Description == "" {
		return errors.New("description is required")
	}
	return nil
}
