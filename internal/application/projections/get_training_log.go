package projections

import (
	"context"
	"fmt"
	"sort"
	"time"

	"workshop/internal/domain/attendance"
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

// GetTrainingLogDeps holds dependencies for the training log projection.
type GetTrainingLogDeps struct {
	AttendanceStore TrainingLogAttendanceStore
	MemberStore     TrainingLogMemberStore
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
	MemberID      string
	MemberName    string
	Program       string
	TotalClasses  int
	TotalMatHours float64 // "flight time"
	CurrentStreak int     // consecutive weeks with at least one check-in
	LastCheckIn   string  // date of most recent check-in
	Entries       []TrainingLogEntry
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

	var totalHours float64
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
				totalHours += duration
			}
		} else {
			// Default 1.5h per class if no checkout
			entry.DurationH = 1.5
			totalHours += 1.5
		}

		entries = append(entries, entry)
	}

	// Reverse entries so most recent is first for display
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	result.TotalClasses = len(records)
	result.TotalMatHours = totalHours
	result.Entries = entries
	result.LastCheckIn = records[len(records)-1].CheckInTime.Format("2006-01-02")
	result.CurrentStreak = calculateWeekStreak(records)

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
