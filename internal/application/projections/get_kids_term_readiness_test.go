package projections

import (
	"context"
	"fmt"
	"testing"
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

// --- Mock stores for kids term readiness tests ---

type mockKRTermStore struct {
	terms []term.Term
}

// List returns all terms.
// PRE: none
// POST: Returns terms list
func (m *mockKRTermStore) List(_ context.Context) ([]term.Term, error) {
	return m.terms, nil
}

type mockKRProgramStore struct {
	programs []program.Program
}

// List returns all programs.
// PRE: none
// POST: Returns programs list
func (m *mockKRProgramStore) List(_ context.Context) ([]program.Program, error) {
	return m.programs, nil
}

type mockKRClassTypeStore struct {
	classTypes map[string][]classtype.ClassType // keyed by programID
}

// ListByProgramID returns class types for a program.
// PRE: programID is non-empty
// POST: Returns class types for the given program
func (m *mockKRClassTypeStore) ListByProgramID(_ context.Context, programID string) ([]classtype.ClassType, error) {
	return m.classTypes[programID], nil
}

type mockKRScheduleStore struct {
	schedules map[string][]schedule.Schedule // keyed by classTypeID
}

// ListByClassTypeID returns schedules for a class type.
// PRE: classTypeID is non-empty
// POST: Returns schedules for the given class type
func (m *mockKRScheduleStore) ListByClassTypeID(_ context.Context, classTypeID string) ([]schedule.Schedule, error) {
	return m.schedules[classTypeID], nil
}

type mockKRHolidayStore struct {
	holidays []holiday.Holiday
}

// List returns all holidays.
// PRE: none
// POST: Returns holidays list
func (m *mockKRHolidayStore) List(_ context.Context) ([]holiday.Holiday, error) {
	return m.holidays, nil
}

type mockKRMemberStore struct {
	members []member.Member
}

// List returns members matching the filter.
// PRE: filter has valid parameters
// POST: Returns matching members
func (m *mockKRMemberStore) List(_ context.Context, filter memberStore.ListFilter) ([]member.Member, error) {
	var result []member.Member
	for _, mem := range m.members {
		if filter.Program != "" && mem.Program != filter.Program {
			continue
		}
		if filter.Status != "" && mem.Status != filter.Status {
			continue
		}
		result = append(result, mem)
	}
	return result, nil
}

type mockKRAttendanceStore struct {
	records []attendance.Attendance
}

// ListByMemberIDAndDateRange returns attendance records for a member within a date range.
// PRE: memberID, startDate, endDate are non-empty
// POST: Returns matching records
func (m *mockKRAttendanceStore) ListByMemberIDAndDateRange(_ context.Context, memberID, startDate, endDate string) ([]attendance.Attendance, error) {
	var result []attendance.Attendance
	for _, a := range m.records {
		d := a.CheckInTime.Format("2006-01-02")
		if a.MemberID == memberID && d >= startDate && d <= endDate {
			result = append(result, a)
		}
	}
	return result, nil
}

type mockKRGradingRecordStore struct {
	records map[string][]grading.Record
}

// ListByMemberID returns grading records for a member.
// PRE: memberID is non-empty
// POST: Returns records for the given member
func (m *mockKRGradingRecordStore) ListByMemberID(_ context.Context, memberID string) ([]grading.Record, error) {
	return m.records[memberID], nil
}

type mockKRGradingConfigStore struct {
	configs map[string]grading.Config // key: program:belt
}

// GetByProgramAndBelt returns the config for a program and belt.
// PRE: program and belt are non-empty
// POST: Returns the config or error if not found
func (m *mockKRGradingConfigStore) GetByProgramAndBelt(_ context.Context, prog, belt string) (grading.Config, error) {
	key := prog + ":" + belt
	if c, ok := m.configs[key]; ok {
		return c, nil
	}
	return grading.Config{}, fmt.Errorf("not found")
}

// --- Helper to build standard test deps ---

func newKidsReadinessTestDeps() GetKidsTermReadinessDeps {
	// Term 1: 2026-01-19 to 2026-04-10 (Mon-Fri, ~12 weeks)
	termStart := time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC)
	termEnd := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	return GetKidsTermReadinessDeps{
		TermStore: &mockKRTermStore{
			terms: []term.Term{
				{ID: "term1", Name: "Term 1 2026", StartDate: termStart, EndDate: termEnd},
			},
		},
		ProgramStore: &mockKRProgramStore{
			programs: []program.Program{
				{ID: "prog-kids", Name: "Kids BJJ", Type: "kids"},
				{ID: "prog-adults", Name: "Adults BJJ", Type: "adults"},
			},
		},
		ClassTypeStore: &mockKRClassTypeStore{
			classTypes: map[string][]classtype.ClassType{
				"prog-kids": {
					{ID: "ct-kids-fund", ProgramID: "prog-kids", Name: "Kids Fundamentals"},
				},
			},
		},
		ScheduleStore: &mockKRScheduleStore{
			schedules: map[string][]schedule.Schedule{
				"ct-kids-fund": {
					{ID: "sched-mon", ClassTypeID: "ct-kids-fund", Day: "monday", StartTime: "16:00", EndTime: "17:00"},
					{ID: "sched-wed", ClassTypeID: "ct-kids-fund", Day: "wednesday", StartTime: "16:00", EndTime: "17:00"},
				},
			},
		},
		HolidayStore: &mockKRHolidayStore{},
		MemberStore: &mockKRMemberStore{
			members: []member.Member{
				{ID: "kid1", Name: "Alice Kid", Program: "kids", Status: "active", Email: "alice@test.com"},
				{ID: "kid2", Name: "Bob Kid", Program: "kids", Status: "active", Email: "bob@test.com"},
			},
		},
		AttendanceStore:    &mockKRAttendanceStore{},
		GradingRecordStore: &mockKRGradingRecordStore{records: make(map[string][]grading.Record)},
		GradingConfigStore: &mockKRGradingConfigStore{
			configs: map[string]grading.Config{
				"kids:grey": {Program: "kids", Belt: grading.BeltGrey, AttendancePct: 80, StripeCount: 4},
			},
		},
	}
}

// TestKidsTermReadiness_NoTerm verifies empty result when no term matches.
func TestKidsTermReadiness_NoTerm(t *testing.T) {
	deps := newKidsReadinessTestDeps()
	// Query for a date outside any term
	query := GetKidsTermReadinessQuery{Now: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)}
	result, err := QueryGetKidsTermReadiness(context.Background(), query, deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected no entries outside term, got %d", len(result.Entries))
	}
}

// TestKidsTermReadiness_InTerm verifies entries returned when in a valid term.
func TestKidsTermReadiness_InTerm(t *testing.T) {
	deps := newKidsReadinessTestDeps()
	query := GetKidsTermReadinessQuery{Now: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)}
	result, err := QueryGetKidsTermReadiness(context.Background(), query, deps)
	if err != nil {
		t.Fatal(err)
	}
	if result.TermName != "Term 1 2026" {
		t.Errorf("expected term name 'Term 1 2026', got %q", result.TermName)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries (2 kids), got %d", len(result.Entries))
	}
	// Both kids should have 0 attendance, 0%
	for _, e := range result.Entries {
		if e.Attended != 0 {
			t.Errorf("expected 0 attended for %s, got %d", e.MemberName, e.Attended)
		}
		if e.TotalSessions == 0 {
			t.Error("expected non-zero total sessions")
		}
		if e.Eligible {
			t.Error("should not be eligible with 0 attendance")
		}
	}
}

// TestKidsTermReadiness_Eligible verifies a kid is marked eligible at 80%+ attendance.
func TestKidsTermReadiness_Eligible(t *testing.T) {
	deps := newKidsReadinessTestDeps()

	// First, find total sessions to know how many we need
	query := GetKidsTermReadinessQuery{Now: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)}
	result, _ := QueryGetKidsTermReadiness(context.Background(), query, deps)
	totalSessions := result.Entries[0].TotalSessions

	// Create enough attendance records for kid1 to reach 80%
	needed := int(float64(totalSessions)*0.85) + 1
	var records []attendance.Attendance
	// Generate attendance on Mondays within the term
	d := time.Date(2026, 1, 19, 16, 0, 0, 0, time.UTC) // first Monday of term
	count := 0
	for count < needed && d.Before(time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)) {
		if d.Weekday() == time.Monday || d.Weekday() == time.Wednesday {
			records = append(records, attendance.Attendance{
				ID:          fmt.Sprintf("att-%d", count),
				MemberID:    "kid1",
				ScheduleID:  "sched-mon",
				CheckInTime: d,
				ClassDate:   d.Format("2006-01-02"),
			})
			count++
		}
		d = d.AddDate(0, 0, 1)
	}
	deps.AttendanceStore = &mockKRAttendanceStore{records: records}

	result, err := QueryGetKidsTermReadiness(context.Background(), query, deps)
	if err != nil {
		t.Fatal(err)
	}

	var kid1 *KidsTermReadinessEntry
	for i := range result.Entries {
		if result.Entries[i].MemberID == "kid1" {
			kid1 = &result.Entries[i]
			break
		}
	}
	if kid1 == nil {
		t.Fatal("kid1 not found in results")
	}
	if !kid1.Eligible {
		t.Errorf("kid1 should be eligible with %d/%d sessions (%.0f%%)", kid1.Attended, kid1.TotalSessions, kid1.AttendancePct)
	}
	if kid1.TargetBelt != grading.BeltGrey {
		t.Errorf("expected target belt grey, got %s", kid1.TargetBelt)
	}
}

// TestKidsTermReadiness_HolidayReducesSessions verifies holidays reduce total available sessions.
func TestKidsTermReadiness_HolidayReducesSessions(t *testing.T) {
	deps := newKidsReadinessTestDeps()
	query := GetKidsTermReadinessQuery{Now: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)}

	// Get baseline sessions
	result1, _ := QueryGetKidsTermReadiness(context.Background(), query, deps)
	baselineSessions := result1.Entries[0].TotalSessions

	// Add a holiday that covers a Monday (2026-01-26)
	deps.HolidayStore = &mockKRHolidayStore{
		holidays: []holiday.Holiday{
			{ID: "h1", Name: "Anniversary Day", StartDate: time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC)},
		},
	}

	result2, err := QueryGetKidsTermReadiness(context.Background(), query, deps)
	if err != nil {
		t.Fatal(err)
	}
	// Should have fewer sessions (2 fewer: Mon schedule on that day = 1 schedule occurrence)
	reducedSessions := result2.Entries[0].TotalSessions
	if reducedSessions >= baselineSessions {
		t.Errorf("expected fewer sessions with holiday, got %d (baseline %d)", reducedSessions, baselineSessions)
	}
}

// TestKidsTermReadiness_ByTermID verifies looking up by explicit term ID.
func TestKidsTermReadiness_ByTermID(t *testing.T) {
	deps := newKidsReadinessTestDeps()
	query := GetKidsTermReadinessQuery{TermID: "term1"}
	result, err := QueryGetKidsTermReadiness(context.Background(), query, deps)
	if err != nil {
		t.Fatal(err)
	}
	if result.TermID != "term1" {
		t.Errorf("expected term ID 'term1', got %q", result.TermID)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
}

// TestKidsTermReadiness_HoursMetricExcluded verifies kids toggled to hours mode are excluded.
func TestKidsTermReadiness_HoursMetricExcluded(t *testing.T) {
	deps := newKidsReadinessTestDeps()
	// Toggle kid1 to hours mode
	deps.MemberStore = &mockKRMemberStore{
		members: []member.Member{
			{ID: "kid1", Name: "Alice Kid", Program: "kids", Status: "active", Email: "alice@test.com", GradingMetric: "hours"},
			{ID: "kid2", Name: "Bob Kid", Program: "kids", Status: "active", Email: "bob@test.com", GradingMetric: "sessions"},
		},
	}

	query := GetKidsTermReadinessQuery{Now: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)}
	result, err := QueryGetKidsTermReadiness(context.Background(), query, deps)
	if err != nil {
		t.Fatal(err)
	}
	// Only kid2 should appear (kid1 is in hours mode)
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry (kid1 hours-mode excluded), got %d", len(result.Entries))
	}
	if result.Entries[0].MemberID != "kid2" {
		t.Errorf("expected kid2, got %s", result.Entries[0].MemberID)
	}
}

// TestKidsTermReadiness_HighestBeltSkipped verifies kids at highest belt are excluded.
func TestKidsTermReadiness_HighestBeltSkipped(t *testing.T) {
	deps := newKidsReadinessTestDeps()
	// Give kid1 a blue belt (highest for kids)
	deps.GradingRecordStore = &mockKRGradingRecordStore{
		records: map[string][]grading.Record{
			"kid1": {{ID: "r1", MemberID: "kid1", Belt: grading.BeltBlue, Stripe: 0, PromotedAt: time.Now(), Method: grading.MethodStandard}},
		},
	}

	query := GetKidsTermReadinessQuery{Now: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)}
	result, err := QueryGetKidsTermReadiness(context.Background(), query, deps)
	if err != nil {
		t.Fatal(err)
	}
	// Only kid2 should appear (kid1 at highest belt)
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry (kid1 at highest belt excluded), got %d", len(result.Entries))
	}
	if result.Entries[0].MemberID != "kid2" {
		t.Errorf("expected kid2, got %s", result.Entries[0].MemberID)
	}
}

// TestKidsTermReadiness_CrossTermIsolation verifies attendance from a previous term
// does not count in a new term (#59: term attendance counts reset).
func TestKidsTermReadiness_CrossTermIsolation(t *testing.T) {
	deps := newKidsReadinessTestDeps()

	// Add a second term: Term 2 runs 2026-04-27 to 2026-07-03
	term2Start := time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)
	term2End := time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)
	deps.TermStore = &mockKRTermStore{
		terms: []term.Term{
			{ID: "term1", Name: "Term 1 2026", StartDate: time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)},
			{ID: "term2", Name: "Term 2 2026", StartDate: term2Start, EndDate: term2End},
		},
	}

	// kid1 attended many sessions in Term 1 (Mondays in Jan-Mar)
	var records []attendance.Attendance
	d := time.Date(2026, 1, 19, 16, 0, 0, 0, time.UTC)
	for i := 0; d.Before(time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)); d = d.AddDate(0, 0, 7) {
		records = append(records, attendance.Attendance{
			ID: fmt.Sprintf("att-t1-%d", i), MemberID: "kid1", ScheduleID: "sched-mon",
			CheckInTime: d, ClassDate: d.Format("2006-01-02"),
		})
		i++
	}
	deps.AttendanceStore = &mockKRAttendanceStore{records: records}

	// Query Term 1 — kid1 should have attendance
	q1 := GetKidsTermReadinessQuery{Now: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)}
	r1, err := QueryGetKidsTermReadiness(context.Background(), q1, deps)
	if err != nil {
		t.Fatal(err)
	}
	var kid1T1 *KidsTermReadinessEntry
	for _, e := range r1.Entries {
		if e.MemberID == "kid1" {
			kid1T1 = &e
			break
		}
	}
	if kid1T1 == nil || kid1T1.Attended == 0 {
		t.Fatal("expected kid1 to have attendance in Term 1")
	}

	// Query Term 2 — kid1 should have 0 attendance (reset)
	q2 := GetKidsTermReadinessQuery{Now: time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)}
	r2, err := QueryGetKidsTermReadiness(context.Background(), q2, deps)
	if err != nil {
		t.Fatal(err)
	}
	if r2.TermName != "Term 2 2026" {
		t.Fatalf("expected Term 2, got %q", r2.TermName)
	}
	for _, e := range r2.Entries {
		if e.MemberID == "kid1" && e.Attended != 0 {
			t.Errorf("kid1 should have 0 attendance in Term 2 (got %d) — term counts must reset", e.Attended)
		}
	}
}
