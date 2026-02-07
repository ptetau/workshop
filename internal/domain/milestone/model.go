package milestone

import (
	"errors"
)

// Domain errors
var (
	ErrEmptyName     = errors.New("milestone name cannot be empty")
	ErrInvalidMetric = errors.New("metric must be one of: classes, mat_hours, streak_weeks")
	ErrZeroThreshold = errors.New("threshold must be greater than zero")
)

// Metric constants
const (
	MetricClasses     = "classes"
	MetricMatHours    = "mat_hours"
	MetricStreakWeeks = "streak_weeks"
)

// ValidMetrics contains all valid milestone metrics.
var ValidMetrics = []string{MetricClasses, MetricMatHours, MetricStreakWeeks}

// Milestone represents an admin-configured achievement (e.g., "100 classes", "200 mat hours").
type Milestone struct {
	ID        string
	Name      string  // e.g., "Century Club"
	Metric    string  // classes, mat_hours, streak_weeks
	Threshold float64 // value that must be reached
	BadgeIcon string  // optional icon/emoji
}

// Validate checks if the Milestone has valid data.
// PRE: Milestone struct is populated
// POST: Returns nil if valid, error otherwise
func (m *Milestone) Validate() error {
	if m.Name == "" {
		return ErrEmptyName
	}
	if !isValidMetric(m.Metric) {
		return ErrInvalidMetric
	}
	if m.Threshold <= 0 {
		return ErrZeroThreshold
	}
	return nil
}

func isValidMetric(metric string) bool {
	for _, v := range ValidMetrics {
		if v == metric {
			return true
		}
	}
	return false
}
