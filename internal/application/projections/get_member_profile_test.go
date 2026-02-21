package projections

import (
	"context"
	"testing"
	"time"

	"workshop/internal/adapters/storage/attendance"
	"workshop/internal/adapters/storage/injury"
	"workshop/internal/adapters/storage/member"
	"workshop/internal/adapters/storage/waiver"
	domainAttendance "workshop/internal/domain/attendance"
	domainGrading "workshop/internal/domain/grading"
	domainInjury "workshop/internal/domain/injury"
	domainMember "workshop/internal/domain/member"
	domainWaiver "workshop/internal/domain/waiver"
)

type mockGetMemberProfileMemberStore struct {
	member domainMember.Member
}

// GetByID returns the seeded member.
// PRE: id is non-empty
// POST: Returns the seeded member
func (m *mockGetMemberProfileMemberStore) GetByID(_ context.Context, _ string) (domainMember.Member, error) {
	return m.member, nil
}

// List is a stub to satisfy the projections.MemberStore interface.
// PRE: filter is valid
// POST: Returns an empty member list
func (m *mockGetMemberProfileMemberStore) List(_ context.Context, _ member.ListFilter) ([]domainMember.Member, error) {
	return nil, nil
}

// Count is a stub to satisfy the projections.MemberStore interface.
// PRE: filter is valid
// POST: Returns count >= 0
func (m *mockGetMemberProfileMemberStore) Count(_ context.Context, _ member.ListFilter) (int, error) {
	return 0, nil
}

type mockGetMemberProfileWaiverStore struct{}

// List returns no waivers.
// PRE: filter is valid
// POST: Returns an empty waiver list
func (m *mockGetMemberProfileWaiverStore) List(_ context.Context, _ waiver.ListFilter) ([]domainWaiver.Waiver, error) {
	return nil, nil
}

type mockGetMemberProfileInjuryStore struct{}

// List returns no injuries.
// PRE: filter is valid
// POST: Returns an empty injury list
func (m *mockGetMemberProfileInjuryStore) List(_ context.Context, _ injury.ListFilter) ([]domainInjury.Injury, error) {
	return nil, nil
}

type mockGetMemberProfileAttendanceStore struct{}

// List returns no attendance records.
// PRE: filter is valid
// POST: Returns an empty attendance list
func (m *mockGetMemberProfileAttendanceStore) List(_ context.Context, _ attendance.ListFilter) ([]domainAttendance.Attendance, error) {
	return nil, nil
}

type mockGetMemberProfileGradingRecordStore struct {
	records []domainGrading.Record
}

// ListByMemberID returns seeded grading records.
// PRE: memberID is non-empty
// POST: Returns any seeded grading records
func (m *mockGetMemberProfileGradingRecordStore) ListByMemberID(_ context.Context, _ string) ([]domainGrading.Record, error) {
	return m.records, nil
}

// TestQueryGetMemberProfile_IncludesLatestBeltAndStripe verifies the profile projection uses the latest grading record.
func TestQueryGetMemberProfile_IncludesLatestBeltAndStripe(t *testing.T) {
	now := time.Now()

	deps := GetMemberProfileDeps{
		MemberStore:        &mockGetMemberProfileMemberStore{member: domainMember.Member{ID: "m1", Name: "Alice", Email: "alice@test.com", Program: "Adults", Status: "active"}},
		WaiverStore:        &mockGetMemberProfileWaiverStore{},
		InjuryStore:        &mockGetMemberProfileInjuryStore{},
		AttendanceStore:    &mockGetMemberProfileAttendanceStore{},
		GradingRecordStore: &mockGetMemberProfileGradingRecordStore{records: []domainGrading.Record{{ID: "r1", MemberID: "m1", Belt: "white", Stripe: 0, PromotedAt: now.Add(-100 * time.Hour)}, {ID: "r2", MemberID: "m1", Belt: "blue", Stripe: 3, PromotedAt: now.Add(-10 * time.Hour)}}},
	}

	res, err := QueryGetMemberProfile(context.Background(), GetMemberProfileQuery{MemberID: "m1"}, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Belt != "blue" || res.Stripe != 3 {
		t.Fatalf("belt/stripe=%q/%d want blue/3", res.Belt, res.Stripe)
	}
}

// TestQueryGetMemberProfile_GradingStoreNil_ReturnsEmptyBelt verifies belt/stripe remain empty when the grading store is nil.
func TestQueryGetMemberProfile_GradingStoreNil_ReturnsEmptyBelt(t *testing.T) {
	deps := GetMemberProfileDeps{
		MemberStore:     &mockGetMemberProfileMemberStore{member: domainMember.Member{ID: "m1", Name: "Alice", Email: "alice@test.com", Program: "Adults", Status: "active"}},
		WaiverStore:     &mockGetMemberProfileWaiverStore{},
		InjuryStore:     &mockGetMemberProfileInjuryStore{},
		AttendanceStore: &mockGetMemberProfileAttendanceStore{},
	}

	res, err := QueryGetMemberProfile(context.Background(), GetMemberProfileQuery{MemberID: "m1"}, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Belt != "" || res.Stripe != 0 {
		t.Fatalf("belt/stripe=%q/%d want empty/0", res.Belt, res.Stripe)
	}
}
