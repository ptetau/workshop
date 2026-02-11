package orchestrators

import (
	"context"
	"fmt"
	"testing"
	"time"

	"workshop/internal/domain/grading"
	"workshop/internal/domain/member"
)

// mockInferMemberStore implements InferStripeMemberStore for testing.
type mockInferMemberStore struct {
	members map[string]member.Member
}

// GetByID returns a member by ID.
// PRE: id is non-empty
// POST: Returns the member or error if not found
func (m *mockInferMemberStore) GetByID(_ context.Context, id string) (member.Member, error) {
	if mem, ok := m.members[id]; ok {
		return mem, nil
	}
	return member.Member{}, fmt.Errorf("not found")
}

// mockInferAttendanceStore implements InferStripeAttendanceStore for testing.
type mockInferAttendanceStore struct {
	hours map[string]float64
}

// SumMatHoursByMemberID returns total mat hours for a member.
// PRE: memberID is non-empty
// POST: Returns total hours (>=0)
func (m *mockInferAttendanceStore) SumMatHoursByMemberID(_ context.Context, memberID string) (float64, error) {
	return m.hours[memberID], nil
}

// mockInferEstimatedHoursStore implements InferStripeEstimatedHoursStore for testing.
type mockInferEstimatedHoursStore struct {
	hours map[string]float64
}

// SumApprovedByMemberID returns total approved estimated hours for a member.
// PRE: memberID is non-empty
// POST: Returns total hours (>=0)
func (m *mockInferEstimatedHoursStore) SumApprovedByMemberID(_ context.Context, memberID string) (float64, error) {
	return m.hours[memberID], nil
}

// mockInferGradingRecordStore implements InferStripeGradingRecordStore for testing.
type mockInferGradingRecordStore struct {
	records map[string][]grading.Record
	saved   []grading.Record
}

// ListByMemberID returns grading records for a member.
// PRE: memberID is non-empty
// POST: Returns records for the given member
func (m *mockInferGradingRecordStore) ListByMemberID(_ context.Context, memberID string) ([]grading.Record, error) {
	return m.records[memberID], nil
}

// Save persists a grading record.
// PRE: r has been validated
// POST: Record is appended to saved slice
func (m *mockInferGradingRecordStore) Save(_ context.Context, r grading.Record) error {
	m.saved = append(m.saved, r)
	return nil
}

// mockInferGradingConfigStore implements InferStripeGradingConfigStore for testing.
type mockInferGradingConfigStore struct {
	configs map[string]grading.Config // key: program+belt
}

// GetByProgramAndBelt returns the config for a program and belt.
// PRE: program and belt are non-empty
// POST: Returns the config or error if not found
func (m *mockInferGradingConfigStore) GetByProgramAndBelt(_ context.Context, program, belt string) (grading.Config, error) {
	key := program + ":" + belt
	if c, ok := m.configs[key]; ok {
		return c, nil
	}
	return grading.Config{}, fmt.Errorf("not found")
}

func newInferTestDeps() (InferStripeDeps, *mockInferGradingRecordStore) {
	recordStore := &mockInferGradingRecordStore{
		records: make(map[string][]grading.Record),
	}
	return InferStripeDeps{
		MemberStore: &mockInferMemberStore{
			members: map[string]member.Member{
				"m1": {ID: "m1", Name: "Test Member", Program: "adults", Status: "active"},
			},
		},
		AttendanceStore: &mockInferAttendanceStore{
			hours: map[string]float64{"m1": 0},
		},
		EstimatedHoursStore: &mockInferEstimatedHoursStore{
			hours: map[string]float64{},
		},
		GradingRecordStore: recordStore,
		GradingConfigStore: &mockInferGradingConfigStore{
			configs: map[string]grading.Config{
				"adults:blue":   {Program: "adults", Belt: grading.BeltBlue, FlightTimeHours: 150, StripeCount: 4},
				"adults:purple": {Program: "adults", Belt: grading.BeltPurple, FlightTimeHours: 300, StripeCount: 4},
			},
		},
	}, recordStore
}

// TestExecuteInferStripe_NoHours verifies no record is created when member has 0 hours.
func TestExecuteInferStripe_NoHours(t *testing.T) {
	deps, recordStore := newInferTestDeps()
	err := ExecuteInferStripe(context.Background(), "m1", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(recordStore.saved) != 0 {
		t.Errorf("expected no saved records, got %d", len(recordStore.saved))
	}
}

// TestExecuteInferStripe_FirstStripe verifies stripe 1 is created at 37.5h (150/4).
func TestExecuteInferStripe_FirstStripe(t *testing.T) {
	deps, recordStore := newInferTestDeps()
	deps.AttendanceStore = &mockInferAttendanceStore{
		hours: map[string]float64{"m1": 38},
	}

	err := ExecuteInferStripe(context.Background(), "m1", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(recordStore.saved) != 1 {
		t.Fatalf("expected 1 saved record, got %d", len(recordStore.saved))
	}
	r := recordStore.saved[0]
	if r.Belt != grading.BeltWhite {
		t.Errorf("expected belt white, got %s", r.Belt)
	}
	if r.Stripe != 1 {
		t.Errorf("expected stripe 1, got %d", r.Stripe)
	}
	if r.Method != grading.MethodInferred {
		t.Errorf("expected method inferred, got %s", r.Method)
	}
	if r.MemberID != "m1" {
		t.Errorf("expected member_id m1, got %s", r.MemberID)
	}
}

// TestExecuteInferStripe_MultipleStripes verifies stripe jumps from 0 to 2 at 75h.
func TestExecuteInferStripe_MultipleStripes(t *testing.T) {
	deps, recordStore := newInferTestDeps()
	deps.AttendanceStore = &mockInferAttendanceStore{
		hours: map[string]float64{"m1": 80},
	}

	err := ExecuteInferStripe(context.Background(), "m1", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(recordStore.saved) != 1 {
		t.Fatalf("expected 1 saved record, got %d", len(recordStore.saved))
	}
	if recordStore.saved[0].Stripe != 2 {
		t.Errorf("expected stripe 2, got %d", recordStore.saved[0].Stripe)
	}
}

// TestExecuteInferStripe_NoChangeWhenAlreadyAtStripe verifies no record created when stripe matches.
func TestExecuteInferStripe_NoChangeWhenAlreadyAtStripe(t *testing.T) {
	deps, recordStore := newInferTestDeps()
	deps.AttendanceStore = &mockInferAttendanceStore{
		hours: map[string]float64{"m1": 38},
	}
	// Member already has stripe 1
	deps.GradingRecordStore = &mockInferGradingRecordStore{
		records: map[string][]grading.Record{
			"m1": {{ID: "r1", MemberID: "m1", Belt: grading.BeltWhite, Stripe: 1, PromotedAt: time.Now(), Method: grading.MethodInferred}},
		},
	}
	recordStore = deps.GradingRecordStore.(*mockInferGradingRecordStore)

	err := ExecuteInferStripe(context.Background(), "m1", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(recordStore.saved) != 0 {
		t.Errorf("expected no saved records when stripe already matches, got %d", len(recordStore.saved))
	}
}

// TestExecuteInferStripe_IncrementFromExistingStripe verifies stripe increments correctly.
func TestExecuteInferStripe_IncrementFromExistingStripe(t *testing.T) {
	deps, _ := newInferTestDeps()
	deps.AttendanceStore = &mockInferAttendanceStore{
		hours: map[string]float64{"m1": 80},
	}
	// Member already has stripe 1
	gradingStore := &mockInferGradingRecordStore{
		records: map[string][]grading.Record{
			"m1": {{ID: "r1", MemberID: "m1", Belt: grading.BeltWhite, Stripe: 1, PromotedAt: time.Now(), Method: grading.MethodInferred}},
		},
	}
	deps.GradingRecordStore = gradingStore

	err := ExecuteInferStripe(context.Background(), "m1", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(gradingStore.saved) != 1 {
		t.Fatalf("expected 1 saved record, got %d", len(gradingStore.saved))
	}
	if gradingStore.saved[0].Stripe != 2 {
		t.Errorf("expected stripe 2, got %d", gradingStore.saved[0].Stripe)
	}
}

// TestExecuteInferStripe_WithBulkEstimates verifies bulk estimated hours are included.
func TestExecuteInferStripe_WithBulkEstimates(t *testing.T) {
	deps, recordStore := newInferTestDeps()
	// 30h attendance + 10h bulk = 40h total → stripe 1
	deps.AttendanceStore = &mockInferAttendanceStore{
		hours: map[string]float64{"m1": 30},
	}
	deps.EstimatedHoursStore = &mockInferEstimatedHoursStore{
		hours: map[string]float64{"m1": 10},
	}

	err := ExecuteInferStripe(context.Background(), "m1", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(recordStore.saved) != 1 {
		t.Fatalf("expected 1 saved record, got %d", len(recordStore.saved))
	}
	if recordStore.saved[0].Stripe != 1 {
		t.Errorf("expected stripe 1, got %d", recordStore.saved[0].Stripe)
	}
}

// TestExecuteInferStripe_HighestBeltSkipsInference verifies no action for black belt.
func TestExecuteInferStripe_HighestBeltSkipsInference(t *testing.T) {
	deps, _ := newInferTestDeps()
	deps.AttendanceStore = &mockInferAttendanceStore{
		hours: map[string]float64{"m1": 1000},
	}
	gradingStore := &mockInferGradingRecordStore{
		records: map[string][]grading.Record{
			"m1": {{ID: "r1", MemberID: "m1", Belt: grading.BeltBlack, Stripe: 0, PromotedAt: time.Now(), Method: grading.MethodStandard}},
		},
	}
	deps.GradingRecordStore = gradingStore

	err := ExecuteInferStripe(context.Background(), "m1", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(gradingStore.saved) != 0 {
		t.Errorf("expected no saved records for highest belt, got %d", len(gradingStore.saved))
	}
}

// TestExecuteInferStripe_CapsAtStripeCount verifies stripe doesn't exceed StripeCount.
func TestExecuteInferStripe_CapsAtStripeCount(t *testing.T) {
	deps, recordStore := newInferTestDeps()
	deps.AttendanceStore = &mockInferAttendanceStore{
		hours: map[string]float64{"m1": 200}, // well over 150h
	}

	err := ExecuteInferStripe(context.Background(), "m1", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(recordStore.saved) != 1 {
		t.Fatalf("expected 1 saved record, got %d", len(recordStore.saved))
	}
	if recordStore.saved[0].Stripe != 4 {
		t.Errorf("expected stripe capped at 4, got %d", recordStore.saved[0].Stripe)
	}
}

// TestExecuteInferStripe_UnknownMemberNoError verifies unknown member is silently skipped.
func TestExecuteInferStripe_UnknownMemberNoError(t *testing.T) {
	deps, recordStore := newInferTestDeps()
	err := ExecuteInferStripe(context.Background(), "unknown", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(recordStore.saved) != 0 {
		t.Errorf("expected no saved records for unknown member, got %d", len(recordStore.saved))
	}
}
