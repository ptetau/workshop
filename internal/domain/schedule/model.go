package schedule

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Day of week constants
const (
	Monday    = "monday"
	Tuesday   = "tuesday"
	Wednesday = "wednesday"
	Thursday  = "thursday"
	Friday    = "friday"
	Saturday  = "saturday"
	Sunday    = "sunday"
)

// ValidDays contains all valid day values.
var ValidDays = []string{Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday}

// Domain errors
var (
	ErrEmptyClassTypeID = errors.New("class type ID cannot be empty")
	ErrInvalidDay       = errors.New("day must be a valid day of the week")
	ErrEmptyStartTime   = errors.New("start time cannot be empty")
	ErrEmptyEndTime     = errors.New("end time cannot be empty")
)

// Schedule represents a recurring weekly class slot.
// Classes are resolved on-the-fly from Schedule + Terms - Holidays.
type Schedule struct {
	ID          string
	ClassTypeID string
	Day         string // monday, tuesday, etc.
	StartTime   string // HH:MM format
	EndTime     string // HH:MM format
}

// Validate checks if the Schedule has valid data.
// PRE: Schedule struct is populated
// POST: Returns nil if valid, error otherwise
func (s *Schedule) Validate() error {
	if strings.TrimSpace(s.ClassTypeID) == "" {
		return ErrEmptyClassTypeID
	}
	if !isValidDay(s.Day) {
		return ErrInvalidDay
	}
	if strings.TrimSpace(s.StartTime) == "" {
		return ErrEmptyStartTime
	}
	if strings.TrimSpace(s.EndTime) == "" {
		return ErrEmptyEndTime
	}
	return nil
}

// DurationHours returns the session duration in hours.
// PRE: StartTime and EndTime are in HH:MM format
// POST: Returns duration as float64 hours, or error if times can't be parsed
func (s *Schedule) DurationHours() (float64, error) {
	start, err := time.Parse("15:04", s.StartTime)
	if err != nil {
		return 0, fmt.Errorf("invalid start time %q: %w", s.StartTime, err)
	}
	end, err := time.Parse("15:04", s.EndTime)
	if err != nil {
		return 0, fmt.Errorf("invalid end time %q: %w", s.EndTime, err)
	}
	dur := end.Sub(start)
	if dur <= 0 {
		dur += 24 * time.Hour // handle overnight classes
	}
	return dur.Hours(), nil
}

func isValidDay(day string) bool {
	for _, d := range ValidDays {
		if d == day {
			return true
		}
	}
	return false
}
