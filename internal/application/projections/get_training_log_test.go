package projections

import (
	"context"
	"testing"
	"time"

	"workshop/internal/domain/attendance"
	"workshop/internal/domain/grading"
	"workshop/internal/domain/member"
)

// mockTrainingLogAttendanceStore implements TrainingLogAttendanceStore for testing.
// PRE: memberID is non-empty
// POST: Returns the stored attendance records or an error
type mockTrainingLogAttendanceStore struct {
	records map[string][]attendance.Attendance
}

// ListByMemberID implements TrainingLogAttendanceStore for testing.
// PRE: memberID is non-empty
// POST: Returns stored attendance records
func (m *mockTrainingLogAttendanceStore) ListByMemberID(_ context.Context, memberID string) ([]attendance.Attendance, error) {
	return m.records[memberID], nil
}

// mockTrainingLogMemberStore implements TrainingLogMemberStore for testing.
// PRE: id is non-empty
// POST: Returns the stored member or an error
type mockTrainingLogMemberStore struct {
	members map[string]member.Member
}

// GetByID implements TrainingLogMemberStore for testing.
// PRE: id is non-empty
// POST: Returns the stored member or an error
func (m *mockTrainingLogMemberStore) GetByID(_ context.Context, id string) (member.Member, error) {
	mem, ok := m.members[id]
	if !ok {
		return member.Member{}, context.DeadlineExceeded
	}
	return mem, nil
}

// mockTrainingLogGradingRecordStore implements TrainingLogGradingRecordStore for testing.
// PRE: memberID is non-empty
// POST: Returns the stored grading records or an error
type mockTrainingLogGradingRecordStore struct {
	records map[string][]grading.Record
}

// ListByMemberID implements TrainingLogGradingRecordStore for testing.
// PRE: memberID is non-empty
// POST: Returns stored grading records
func (m *mockTrainingLogGradingRecordStore) ListByMemberID(_ context.Context, memberID string) ([]grading.Record, error) {
	return m.records[memberID], nil
}

// mockTrainingLogGradingConfigStore implements TrainingLogGradingConfigStore for testing.
// PRE: program and belt are non-empty
// POST: Returns the stored config or an error
type mockTrainingLogGradingConfigStore struct {
	configs map[string]grading.Config // key = program+belt
}

// GetByProgramAndBelt implements TrainingLogGradingConfigStore for testing.
// PRE: program and belt are non-empty
// POST: Returns the stored config or an error
func (m *mockTrainingLogGradingConfigStore) GetByProgramAndBelt(_ context.Context, program, belt string) (grading.Config, error) {
	config, ok := m.configs[program+belt]
	if !ok {
		return grading.Config{}, context.DeadlineExceeded
	}
	return config, nil
}

// TestQueryGetTrainingLog_BeltAndProgress verifies belt lookup and progress bar computation.
func TestQueryGetTrainingLog_BeltAndProgress(t *testing.T) {
	now := time.Now()
	memberID := "m1"

	deps := GetTrainingLogDeps{
		AttendanceStore: &mockTrainingLogAttendanceStore{
			records: map[string][]attendance.Attendance{
				memberID: {
					{ID: "a1", MemberID: memberID, CheckInTime: now.Add(-2 * time.Hour), CheckOutTime: now.Add(-1 * time.Hour), ClassDate: now.Format("2006-01-02")},
					{ID: "a2", MemberID: memberID, CheckInTime: now.Add(-4 * time.Hour), ClassDate: now.Format("2006-01-02")},
				},
			},
		},
		MemberStore: &mockTrainingLogMemberStore{
			members: map[string]member.Member{
				memberID: {ID: memberID, Name: "Alice", Program: "adults"},
			},
		},
		GradingRecordStore: &mockTrainingLogGradingRecordStore{
			records: map[string][]grading.Record{
				memberID: {
					{ID: "g1", MemberID: memberID, Belt: "blue", Stripe: 2, PromotedAt: now.Add(-30 * 24 * time.Hour)},
				},
			},
		},
		GradingConfigStore: &mockTrainingLogGradingConfigStore{
			configs: map[string]grading.Config{
				"adultspurple": {ID: "c1", Program: "adults", Belt: "purple", FlightTimeHours: 300},
			},
		},
	}

	result, err := QueryGetTrainingLog(context.Background(), GetTrainingLogQuery{MemberID: memberID}, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Belt != "blue" {
		t.Errorf("expected belt=blue, got %q", result.Belt)
	}
	if result.Stripe != 2 {
		t.Errorf("expected stripe=2, got %d", result.Stripe)
	}
	if result.NextBelt != "purple" {
		t.Errorf("expected next belt=purple, got %q", result.NextBelt)
	}
	if result.RequiredHours != 300 {
		t.Errorf("expected required hours=300, got %f", result.RequiredHours)
	}
	if result.ProgressPct <= 0 {
		t.Error("expected progress > 0")
	}
}

// TestQueryGetTrainingLog_RecordedAndEstimatedHours verifies hours split.
func TestQueryGetTrainingLog_RecordedAndEstimatedHours(t *testing.T) {
	now := time.Now()
	memberID := "m1"

	deps := GetTrainingLogDeps{
		AttendanceStore: &mockTrainingLogAttendanceStore{
			records: map[string][]attendance.Attendance{
				memberID: {
					// 1 hour recorded session
					{ID: "a1", MemberID: memberID, CheckInTime: now.Add(-2 * time.Hour), CheckOutTime: now.Add(-1 * time.Hour), ClassDate: now.Format("2006-01-02")},
					// No checkout = 1.5h estimated
					{ID: "a2", MemberID: memberID, CheckInTime: now.Add(-4 * time.Hour), ClassDate: now.Format("2006-01-02")},
				},
			},
		},
		MemberStore: &mockTrainingLogMemberStore{
			members: map[string]member.Member{
				memberID: {ID: memberID, Name: "Alice", Program: "adults"},
			},
		},
	}

	result, err := QueryGetTrainingLog(context.Background(), GetTrainingLogQuery{MemberID: memberID}, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalClasses != 2 {
		t.Errorf("expected 2 classes, got %d", result.TotalClasses)
	}
	if result.RecordedHours != 1.0 {
		t.Errorf("expected recorded=1.0, got %f", result.RecordedHours)
	}
	if result.EstimatedHours != 1.5 {
		t.Errorf("expected estimated=1.5, got %f", result.EstimatedHours)
	}
	if result.TotalMatHours != 2.5 {
		t.Errorf("expected total=2.5, got %f", result.TotalMatHours)
	}
}

// TestQueryGetTrainingLog_EmptyRecords verifies zero state.
func TestQueryGetTrainingLog_EmptyRecords(t *testing.T) {
	memberID := "m1"

	deps := GetTrainingLogDeps{
		AttendanceStore: &mockTrainingLogAttendanceStore{
			records: map[string][]attendance.Attendance{},
		},
		MemberStore: &mockTrainingLogMemberStore{
			members: map[string]member.Member{
				memberID: {ID: memberID, Name: "Bob", Program: "adults"},
			},
		},
	}

	result, err := QueryGetTrainingLog(context.Background(), GetTrainingLogQuery{MemberID: memberID}, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalClasses != 0 {
		t.Errorf("expected 0 classes, got %d", result.TotalClasses)
	}
	if result.Belt != "" {
		t.Errorf("expected empty belt for no records + nil grading store, got %q", result.Belt)
	}
}

// TestQueryGetTrainingLog_WhiteBeltDefault verifies default belt is white when no grading records exist.
func TestQueryGetTrainingLog_WhiteBeltDefault(t *testing.T) {
	now := time.Now()
	memberID := "m1"

	deps := GetTrainingLogDeps{
		AttendanceStore: &mockTrainingLogAttendanceStore{
			records: map[string][]attendance.Attendance{
				memberID: {
					{ID: "a1", MemberID: memberID, CheckInTime: now.Add(-1 * time.Hour), CheckOutTime: now, ClassDate: now.Format("2006-01-02")},
				},
			},
		},
		MemberStore: &mockTrainingLogMemberStore{
			members: map[string]member.Member{
				memberID: {ID: memberID, Name: "Charlie", Program: "adults"},
			},
		},
		GradingRecordStore: &mockTrainingLogGradingRecordStore{
			records: map[string][]grading.Record{},
		},
		GradingConfigStore: &mockTrainingLogGradingConfigStore{
			configs: map[string]grading.Config{
				"adultsblue": {ID: "c1", Program: "adults", Belt: "blue", FlightTimeHours: 150},
			},
		},
	}

	result, err := QueryGetTrainingLog(context.Background(), GetTrainingLogQuery{MemberID: memberID}, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Belt != "white" {
		t.Errorf("expected default belt=white, got %q", result.Belt)
	}
	if result.NextBelt != "blue" {
		t.Errorf("expected next belt=blue, got %q", result.NextBelt)
	}
	if result.RequiredHours != 150 {
		t.Errorf("expected required hours=150, got %f", result.RequiredHours)
	}
}

// TestNextBeltInProgression verifies belt progression lookup.
func TestNextBeltInProgression(t *testing.T) {
	tests := []struct {
		current  string
		program  string
		expected string
	}{
		{"white", "adults", "blue"},
		{"blue", "adults", "purple"},
		{"purple", "adults", "brown"},
		{"brown", "adults", "black"},
		{"black", "adults", ""},
		{"white", "kids", "grey"},
		{"green", "kids", "blue"},
		{"blue", "kids", ""},
	}

	for _, tt := range tests {
		got := nextBeltInProgression(tt.current, tt.program)
		if got != tt.expected {
			t.Errorf("nextBeltInProgression(%q, %q) = %q, want %q", tt.current, tt.program, got, tt.expected)
		}
	}
}
