package schedule

import (
	"errors"
	"strings"
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

func isValidDay(day string) bool {
	for _, d := range ValidDays {
		if d == day {
			return true
		}
	}
	return false
}
