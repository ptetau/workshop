package projections

import (
	"context"
	"strings"
	"time"

	memberStore "workshop/internal/adapters/storage/member"
	"workshop/internal/domain/attendance"
	"workshop/internal/domain/classtype"
	"workshop/internal/domain/grading"
	"workshop/internal/domain/holiday"
	"workshop/internal/domain/member"
	"workshop/internal/domain/program"
	"workshop/internal/domain/schedule"
	"workshop/internal/domain/term"
)

// KidsReadinessTermStore defines the term store interface needed by this projection.
type KidsReadinessTermStore interface {
	List(ctx context.Context) ([]term.Term, error)
}

// KidsReadinessProgramStore defines the program store interface needed by this projection.
type KidsReadinessProgramStore interface {
	List(ctx context.Context) ([]program.Program, error)
}

// KidsReadinessClassTypeStore defines the class type store interface needed by this projection.
type KidsReadinessClassTypeStore interface {
	ListByProgramID(ctx context.Context, programID string) ([]classtype.ClassType, error)
}

// KidsReadinessScheduleStore defines the schedule store interface needed by this projection.
type KidsReadinessScheduleStore interface {
	ListByClassTypeID(ctx context.Context, classTypeID string) ([]schedule.Schedule, error)
}

// KidsReadinessHolidayStore defines the holiday store interface needed by this projection.
type KidsReadinessHolidayStore interface {
	List(ctx context.Context) ([]holiday.Holiday, error)
}

// KidsReadinessMemberStore defines the member store interface needed by this projection.
type KidsReadinessMemberStore interface {
	List(ctx context.Context, filter memberStore.ListFilter) ([]member.Member, error)
}

// KidsReadinessAttendanceStore defines the attendance store interface needed by this projection.
type KidsReadinessAttendanceStore interface {
	ListByMemberIDAndDateRange(ctx context.Context, memberID string, startDate string, endDate string) ([]attendance.Attendance, error)
}

// KidsReadinessGradingRecordStore defines the grading record store interface needed by this projection.
type KidsReadinessGradingRecordStore interface {
	ListByMemberID(ctx context.Context, memberID string) ([]grading.Record, error)
}

// KidsReadinessGradingConfigStore defines the grading config store interface needed by this projection.
type KidsReadinessGradingConfigStore interface {
	GetByProgramAndBelt(ctx context.Context, program, belt string) (grading.Config, error)
}

// GetKidsTermReadinessDeps holds dependencies for the kids term readiness projection.
type GetKidsTermReadinessDeps struct {
	TermStore          KidsReadinessTermStore
	ProgramStore       KidsReadinessProgramStore
	ClassTypeStore     KidsReadinessClassTypeStore
	ScheduleStore      KidsReadinessScheduleStore
	HolidayStore       KidsReadinessHolidayStore
	MemberStore        KidsReadinessMemberStore
	AttendanceStore    KidsReadinessAttendanceStore
	GradingRecordStore KidsReadinessGradingRecordStore
	GradingConfigStore KidsReadinessGradingConfigStore
}

// GetKidsTermReadinessQuery carries input for the kids term readiness projection.
type GetKidsTermReadinessQuery struct {
	TermID string    // optional: if empty, uses the term containing Now
	Now    time.Time // used to find current term if TermID is empty
}

// KidsTermReadinessEntry represents a single kid's readiness for belt promotion.
type KidsTermReadinessEntry struct {
	MemberID      string
	MemberName    string
	CurrentBelt   string
	TargetBelt    string
	Attended      int
	TotalSessions int
	AttendancePct float64
	ThresholdPct  float64
	Eligible      bool
}

// KidsTermReadinessResult carries the output of the kids term readiness projection.
type KidsTermReadinessResult struct {
	TermName string
	TermID   string
	Entries  []KidsTermReadinessEntry
}

// QueryGetKidsTermReadiness computes kids grading readiness by term attendance percentage.
// Algorithm:
// 1. Find the target term (by ID or current date)
// 2. Find all kids program schedules
// 3. Count available sessions in the term (schedule occurrences minus holidays)
// 4. For each active kids member, count attendance in the term for kids schedules
// 5. Calculate attendance percentage and eligibility against the config threshold
func QueryGetKidsTermReadiness(ctx context.Context, query GetKidsTermReadinessQuery, deps GetKidsTermReadinessDeps) (KidsTermReadinessResult, error) {
	// Step 1: Find the target term
	terms, err := deps.TermStore.List(ctx)
	if err != nil {
		return KidsTermReadinessResult{}, err
	}

	var targetTerm term.Term
	found := false
	if query.TermID != "" {
		for _, t := range terms {
			if t.ID == query.TermID {
				targetTerm = t
				found = true
				break
			}
		}
	} else {
		for _, t := range terms {
			if t.Contains(query.Now) {
				targetTerm = t
				found = true
				break
			}
		}
	}
	if !found {
		return KidsTermReadinessResult{}, nil // no matching term
	}

	// Step 2: Find all kids program IDs and their schedules
	programs, err := deps.ProgramStore.List(ctx)
	if err != nil {
		return KidsTermReadinessResult{}, err
	}

	var kidsSchedules []schedule.Schedule
	kidsScheduleIDs := make(map[string]bool)
	for _, p := range programs {
		if p.Type != "kids" {
			continue
		}
		classTypes, err := deps.ClassTypeStore.ListByProgramID(ctx, p.ID)
		if err != nil {
			continue
		}
		for _, ct := range classTypes {
			scheds, err := deps.ScheduleStore.ListByClassTypeID(ctx, ct.ID)
			if err != nil {
				continue
			}
			for _, s := range scheds {
				kidsSchedules = append(kidsSchedules, s)
				kidsScheduleIDs[s.ID] = true
			}
		}
	}

	if len(kidsSchedules) == 0 {
		return KidsTermReadinessResult{TermName: targetTerm.Name, TermID: targetTerm.ID}, nil
	}

	// Step 3: Count available sessions in the term (schedule day occurrences minus holidays)
	holidays, err := deps.HolidayStore.List(ctx)
	if err != nil {
		return KidsTermReadinessResult{}, err
	}

	totalSessions := countSessionsInTerm(kidsSchedules, targetTerm, holidays)
	if totalSessions == 0 {
		return KidsTermReadinessResult{TermName: targetTerm.Name, TermID: targetTerm.ID}, nil
	}

	// Step 4: For each active kids member, count attendance
	startDate := targetTerm.StartDate.Format("2006-01-02")
	endDate := targetTerm.EndDate.Format("2006-01-02")

	result := KidsTermReadinessResult{
		TermName: targetTerm.Name,
		TermID:   targetTerm.ID,
	}

	members, err := deps.MemberStore.List(ctx, memberStore.ListFilter{
		Limit:   10000,
		Program: "kids",
		Status:  "active",
	})
	if err != nil {
		return KidsTermReadinessResult{}, err
	}

	for _, m := range members {

		// Get current belt
		currentBelt := grading.BeltWhite
		records, err := deps.GradingRecordStore.ListByMemberID(ctx, m.ID)
		if err == nil && len(records) > 0 {
			latest := records[0]
			for _, r := range records[1:] {
				if r.PromotedAt.After(latest.PromotedAt) {
					latest = r
				}
			}
			currentBelt = latest.Belt
		}

		// Find next belt and threshold
		nextBelt := nextKidsBelt(currentBelt)
		if nextBelt == "" {
			continue // at highest kids belt
		}

		thresholdPct := 80.0 // default
		config, err := deps.GradingConfigStore.GetByProgramAndBelt(ctx, "kids", nextBelt)
		if err == nil && config.AttendancePct > 0 {
			thresholdPct = config.AttendancePct
		}

		// Count attendance in term for kids schedules
		attendanceRecords, err := deps.AttendanceStore.ListByMemberIDAndDateRange(ctx, m.ID, startDate, endDate)
		if err != nil {
			continue
		}
		attended := 0
		for _, a := range attendanceRecords {
			if kidsScheduleIDs[a.ScheduleID] {
				attended++
			}
		}

		pct := 0.0
		if totalSessions > 0 {
			pct = (float64(attended) / float64(totalSessions)) * 100
		}

		result.Entries = append(result.Entries, KidsTermReadinessEntry{
			MemberID:      m.ID,
			MemberName:    m.Name,
			CurrentBelt:   currentBelt,
			TargetBelt:    nextBelt,
			Attended:      attended,
			TotalSessions: totalSessions,
			AttendancePct: pct,
			ThresholdPct:  thresholdPct,
			Eligible:      pct >= thresholdPct,
		})
	}

	return result, nil
}

// countSessionsInTerm counts how many schedule occurrences fall within the term,
// excluding dates that overlap with holidays.
func countSessionsInTerm(schedules []schedule.Schedule, t term.Term, holidays []holiday.Holiday) int {
	// Build a set of days-of-week that have kids schedules
	dayCount := make(map[string]int)
	for _, s := range schedules {
		dayCount[s.Day]++
	}

	count := 0
	start := t.StartDate.Truncate(24 * time.Hour)
	end := t.EndDate.Truncate(24 * time.Hour)

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dayName := strings.ToLower(d.Weekday().String())
		n, ok := dayCount[dayName]
		if !ok || n == 0 {
			continue
		}

		// Check if this date is a holiday
		isHoliday := false
		for _, h := range holidays {
			if h.Contains(d) {
				isHoliday = true
				break
			}
		}
		if !isHoliday {
			count += n // add number of kids sessions on this day
		}
	}

	return count
}

// nextKidsBelt returns the next belt in the kids progression, or "" if at highest.
func nextKidsBelt(current string) string {
	for i, b := range grading.KidsBelts {
		if b == current && i+1 < len(grading.KidsBelts) {
			return grading.KidsBelts[i+1]
		}
	}
	return ""
}
