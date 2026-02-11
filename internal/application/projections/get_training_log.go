package projections

import (
	"context"
	"fmt"
	"sort"
	"time"

	"workshop/internal/domain/attendance"
	"workshop/internal/domain/grading"
	"workshop/internal/domain/member"
)

// TrainingLogAttendanceStore defines the attendance store interface needed by the training log projection.
type TrainingLogAttendanceStore interface {
	ListByMemberID(ctx context.Context, memberID string) ([]attendance.Attendance, error)
}

// TrainingLogMemberStore defines the member store interface needed by the training log projection.
type TrainingLogMemberStore interface {
	GetByID(ctx context.Context, id string) (member.Member, error)
}

// GetTrainingLogQuery carries input for the training log projection.
type GetTrainingLogQuery struct {
	MemberID string
}

// TrainingLogGradingRecordStore defines the grading record store interface.
type TrainingLogGradingRecordStore interface {
	ListByMemberID(ctx context.Context, memberID string) ([]grading.Record, error)
}

// TrainingLogGradingConfigStore defines the grading config store interface.
type TrainingLogGradingConfigStore interface {
	GetByProgramAndBelt(ctx context.Context, program, belt string) (grading.Config, error)
}

// TrainingLogEstimatedHoursStore defines the estimated hours store interface needed by the training log projection.
type TrainingLogEstimatedHoursStore interface {
	SumApprovedByMemberID(ctx context.Context, memberID string) (float64, error)
}

// GetTrainingLogDeps holds dependencies for the training log projection.
type GetTrainingLogDeps struct {
	AttendanceStore     TrainingLogAttendanceStore
	MemberStore         TrainingLogMemberStore
	GradingRecordStore  TrainingLogGradingRecordStore  // optional: nil skips belt lookup
	GradingConfigStore  TrainingLogGradingConfigStore  // optional: nil skips progress bar
	EstimatedHoursStore TrainingLogEstimatedHoursStore // optional: nil skips bulk estimates
}

// TrainingLogEntry represents a single attendance entry in the training log.
type TrainingLogEntry struct {
	Date       string  // YYYY-MM-DD
	CheckIn    string  // HH:MM
	CheckOut   string  // HH:MM or empty
	DurationH  float64 // hours (0 if no checkout)
	ClassDate  string
	ScheduleID string
}

// TrainingLogResult carries the output of the training log projection.
type TrainingLogResult struct {
	MemberID           string
	MemberName         string
	Program            string
	TotalClasses       int
	TotalMatHours      float64 // "flight time" (recorded + session-estimated + bulk-estimated)
	RecordedHours      float64 // hours from checked-out sessions
	EstimatedHours     float64 // hours from default estimate (no checkout)
	BulkEstimatedHours float64 // hours from coach/admin bulk estimates
	CurrentStreak      int     // consecutive weeks with at least one check-in
	LastCheckIn        string  // date of most recent check-in
	Belt               string  // current belt
	Stripe             int     // current stripes
	NextBelt           string  // next belt in progression (empty if at highest)
	ProgressPct        float64 // percentage progress toward next belt (0-100)
	RequiredHours      float64 // hours required for next belt
	Entries            []TrainingLogEntry
}

// QueryGetTrainingLog computes the training log for a member from their attendance history.
func QueryGetTrainingLog(ctx context.Context, query GetTrainingLogQuery, deps GetTrainingLogDeps) (TrainingLogResult, error) {
	m, err := deps.MemberStore.GetByID(ctx, query.MemberID)
	if err != nil {
		return TrainingLogResult{}, err
	}

	records, err := deps.AttendanceStore.ListByMemberID(ctx, query.MemberID)
	if err != nil {
		return TrainingLogResult{}, err
	}

	// Sort by check-in time ascending for streak calculation
	sort.Slice(records, func(i, j int) bool {
		return records[i].CheckInTime.Before(records[j].CheckInTime)
	})

	result := TrainingLogResult{
		MemberID:   m.ID,
		MemberName: m.Name,
		Program:    m.Program,
	}

	if len(records) == 0 {
		return result, nil
	}

	var recordedHours, estimatedHours float64
	entries := make([]TrainingLogEntry, 0, len(records))

	for _, r := range records {
		entry := TrainingLogEntry{
			Date:       r.CheckInTime.Format("2006-01-02"),
			CheckIn:    r.CheckInTime.Format("15:04"),
			ClassDate:  r.ClassDate,
			ScheduleID: r.ScheduleID,
		}

		if !r.CheckOutTime.IsZero() {
			entry.CheckOut = r.CheckOutTime.Format("15:04")
			duration := r.CheckOutTime.Sub(r.CheckInTime).Hours()
			if duration > 0 {
				entry.DurationH = duration
				recordedHours += duration
			}
		} else {
			// Default 1.5h per class if no checkout
			entry.DurationH = 1.5
			estimatedHours += 1.5
		}

		entries = append(entries, entry)
	}

	// Reverse entries so most recent is first for display
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	result.TotalClasses = len(records)
	result.RecordedHours = recordedHours
	result.EstimatedHours = estimatedHours
	result.TotalMatHours = recordedHours + estimatedHours

	// Add bulk-estimated hours (from §3.4)
	if deps.EstimatedHoursStore != nil {
		bulkHours, err := deps.EstimatedHoursStore.SumApprovedByMemberID(ctx, query.MemberID)
		if err == nil {
			result.BulkEstimatedHours = bulkHours
			result.TotalMatHours += bulkHours
		}
	}
	result.Entries = entries
	result.LastCheckIn = records[len(records)-1].CheckInTime.Format("2006-01-02")
	result.CurrentStreak = calculateWeekStreak(records)

	// Belt and progress bar (optional deps)
	if deps.GradingRecordStore != nil {
		gradingRecords, err := deps.GradingRecordStore.ListByMemberID(ctx, query.MemberID)
		if err == nil && len(gradingRecords) > 0 {
			latest := gradingRecords[0]
			for _, r := range gradingRecords[1:] {
				if r.PromotedAt.After(latest.PromotedAt) {
					latest = r
				}
			}
			result.Belt = latest.Belt
			result.Stripe = latest.Stripe
		}
	}

	// Progress toward next belt
	currentBelt := result.Belt
	if currentBelt == "" {
		currentBelt = "white"
		result.Belt = "white"
	}
	result.NextBelt = nextBeltInProgression(currentBelt, m.Program)
	if result.NextBelt != "" && deps.GradingConfigStore != nil {
		config, err := deps.GradingConfigStore.GetByProgramAndBelt(ctx, m.Program, result.NextBelt)
		if err == nil && config.FlightTimeHours > 0 {
			result.RequiredHours = config.FlightTimeHours
			pct := (result.TotalMatHours / config.FlightTimeHours) * 100
			if pct > 100 {
				pct = 100
			}
			result.ProgressPct = pct
		}
	}

	return result, nil
}

// calculateWeekStreak counts consecutive weeks (ending with this week) that have
// at least one check-in. A "week" runs Monday–Sunday.
func calculateWeekStreak(records []attendance.Attendance) int {
	if len(records) == 0 {
		return 0
	}

	// Collect unique ISO weeks that have check-ins
	weekSet := make(map[string]bool)
	for _, r := range records {
		y, w := r.CheckInTime.ISOWeek()
		key := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local).Format("2006") + "-" + padWeek(w)
		weekSet[key] = true
	}

	// Walk backwards from current week
	now := time.Now()
	streak := 0
	for {
		y, w := now.ISOWeek()
		key := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local).Format("2006") + "-" + padWeek(w)
		if !weekSet[key] {
			break
		}
		streak++
		now = now.AddDate(0, 0, -7)
	}

	return streak
}

func padWeek(w int) string {
	return fmt.Sprintf("%02d", w)
}

// nextBeltInProgression returns the next belt in the progression, or "" if at highest.
func nextBeltInProgression(current, program string) string {
	var progression []string
	if program == "kids" {
		progression = grading.KidsBelts
	} else {
		progression = grading.AdultBelts
	}
	for i, b := range progression {
		if b == current && i+1 < len(progression) {
			return progression[i+1]
		}
	}
	return ""
}
