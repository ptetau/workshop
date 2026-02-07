package projections

import (
	"context"
	"strings"
	"time"

	"workshop/internal/domain/classtype"
	"workshop/internal/domain/holiday"
	"workshop/internal/domain/program"
	"workshop/internal/domain/schedule"
	"workshop/internal/domain/term"
)

// TodaysClassesScheduleStore defines the store interface needed by this projection.
type TodaysClassesScheduleStore interface {
	ListByDay(ctx context.Context, day string) ([]schedule.Schedule, error)
}

// TodaysClassesTermStore defines the store interface needed by this projection.
type TodaysClassesTermStore interface {
	List(ctx context.Context) ([]term.Term, error)
}

// TodaysClassesHolidayStore defines the store interface needed by this projection.
type TodaysClassesHolidayStore interface {
	List(ctx context.Context) ([]holiday.Holiday, error)
}

// TodaysClassesClassTypeStore defines the store interface needed by this projection.
type TodaysClassesClassTypeStore interface {
	GetByID(ctx context.Context, id string) (classtype.ClassType, error)
}

// TodaysClassesProgramStore defines the store interface needed by this projection.
type TodaysClassesProgramStore interface {
	GetByID(ctx context.Context, id string) (program.Program, error)
}

// GetTodaysClassesDeps holds dependencies for the projection.
type GetTodaysClassesDeps struct {
	ScheduleStore  TodaysClassesScheduleStore
	TermStore      TodaysClassesTermStore
	HolidayStore   TodaysClassesHolidayStore
	ClassTypeStore TodaysClassesClassTypeStore
	ProgramStore   TodaysClassesProgramStore
}

// TodaysClassResult represents a single class session resolved for today.
type TodaysClassResult struct {
	ScheduleID   string
	ClassTypeID  string
	ClassTypeName string
	ProgramID    string
	ProgramName  string
	ProgramType  string
	Day          string
	StartTime    string
	EndTime      string
}

// QueryGetTodaysClasses resolves today's classes on-the-fly from Schedule + Terms - Holidays.
// Algorithm: 1) Get today's day-of-week, 2) Check if today is within a term,
// 3) Check if today is a holiday, 4) If in-term and not-holiday, return matching schedules.
func QueryGetTodaysClasses(ctx context.Context, now time.Time, deps GetTodaysClassesDeps) ([]TodaysClassResult, error) {
	// Step 1: Check if today falls within any term
	terms, err := deps.TermStore.List(ctx)
	if err != nil {
		return nil, err
	}

	inTerm := false
	for _, t := range terms {
		if t.Contains(now) {
			inTerm = true
			break
		}
	}

	if !inTerm {
		return nil, nil // No classes outside of term time
	}

	// Step 2: Check if today is a holiday
	holidays, err := deps.HolidayStore.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, h := range holidays {
		if h.Contains(now) {
			return nil, nil // No classes on holidays
		}
	}

	// Step 3: Get schedules for today's day of week
	dayName := strings.ToLower(now.Weekday().String())
	schedules, err := deps.ScheduleStore.ListByDay(ctx, dayName)
	if err != nil {
		return nil, err
	}

	// Step 4: Enrich with class type and program info
	var results []TodaysClassResult
	for _, s := range schedules {
		ct, err := deps.ClassTypeStore.GetByID(ctx, s.ClassTypeID)
		if err != nil {
			continue // Skip if class type not found
		}

		p, err := deps.ProgramStore.GetByID(ctx, ct.ProgramID)
		if err != nil {
			continue // Skip if program not found
		}

		results = append(results, TodaysClassResult{
			ScheduleID:    s.ID,
			ClassTypeID:   s.ClassTypeID,
			ClassTypeName: ct.Name,
			ProgramID:     p.ID,
			ProgramName:   p.Name,
			ProgramType:   p.Type,
			Day:           s.Day,
			StartTime:     s.StartTime,
			EndTime:       s.EndTime,
		})
	}

	return results, nil
}
