package projections

import (
	"context"
	"testing"
	"time"

	"workshop/internal/adapters/storage/injury"
	"workshop/internal/adapters/storage/member"
	domainGrading "workshop/internal/domain/grading"
	domainInjury "workshop/internal/domain/injury"
	domainMember "workshop/internal/domain/member"
)

type mockGetMemberListMemberStore struct {
	members []domainMember.Member
}

// GetByID returns a seeded member by ID.
// PRE: id is non-empty
// POST: Returns the seeded member or an error
func (m *mockGetMemberListMemberStore) GetByID(_ context.Context, id string) (domainMember.Member, error) {
	for _, mem := range m.members {
		if mem.ID == id {
			return mem, nil
		}
	}
	return domainMember.Member{}, context.DeadlineExceeded
}

// List returns all seeded members.
// PRE: filter is valid
// POST: Returns all seeded members
func (m *mockGetMemberListMemberStore) List(_ context.Context, _ member.ListFilter) ([]domainMember.Member, error) {
	return m.members, nil
}

// Count returns the number of seeded members.
// PRE: filter is valid
// POST: Returns count >= 0
func (m *mockGetMemberListMemberStore) Count(_ context.Context, _ member.ListFilter) (int, error) {
	return len(m.members), nil
}

type mockGetMemberListInjuryStore struct{}

// List returns no injuries.
// PRE: filter is valid
// POST: Returns an empty injury list
func (m *mockGetMemberListInjuryStore) List(_ context.Context, _ injury.ListFilter) ([]domainInjury.Injury, error) {
	return nil, nil
}

type mockGetMemberListGradingRecordStore struct {
	records map[string][]domainGrading.Record
}

// ListByMemberID returns seeded grading records for the member.
// PRE: memberID is non-empty
// POST: Returns any seeded grading records
func (m *mockGetMemberListGradingRecordStore) ListByMemberID(_ context.Context, memberID string) ([]domainGrading.Record, error) {
	return m.records[memberID], nil
}

// TestQueryGetMemberList_IncludesLatestBeltAndStripe verifies the list projection uses the latest grading record.
func TestQueryGetMemberList_IncludesLatestBeltAndStripe(t *testing.T) {
	now := time.Now()
	m1 := domainMember.Member{ID: "m1", Name: "Alice", Email: "alice@test.com", Program: "Adults", Status: "active"}
	m2 := domainMember.Member{ID: "m2", Name: "Bob", Email: "bob@test.com", Program: "Kids", Status: "trial"}

	deps := GetMemberListDeps{
		MemberStore: &mockGetMemberListMemberStore{members: []domainMember.Member{m1, m2}},
		InjuryStore: &mockGetMemberListInjuryStore{},
		GradingRecordStore: &mockGetMemberListGradingRecordStore{records: map[string][]domainGrading.Record{
			m1.ID: {
				{ID: "r1", MemberID: m1.ID, Belt: "white", Stripe: 1, PromotedAt: now.Add(-30 * 24 * time.Hour)},
				{ID: "r2", MemberID: m1.ID, Belt: "blue", Stripe: 2, PromotedAt: now.Add(-10 * 24 * time.Hour)},
			},
		}},
	}

	res, err := QueryGetMemberList(context.Background(), GetMemberListQuery{}, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Members) != 2 {
		t.Fatalf("members=%d want 2", len(res.Members))
	}

	if res.Members[0].Belt != "blue" || res.Members[0].Stripe != 2 {
		t.Fatalf("member[0] belt/stripe=%q/%d want blue/2", res.Members[0].Belt, res.Members[0].Stripe)
	}
	if res.Members[1].Belt != "" || res.Members[1].Stripe != 0 {
		t.Fatalf("member[1] belt/stripe=%q/%d want empty/0", res.Members[1].Belt, res.Members[1].Stripe)
	}
}

// TestQueryGetMemberList_GradingStoreNil_ReturnsEmptyBelt verifies belt/stripe remain empty when the grading store is nil.
func TestQueryGetMemberList_GradingStoreNil_ReturnsEmptyBelt(t *testing.T) {
	m1 := domainMember.Member{ID: "m1", Name: "Alice", Email: "alice@test.com", Program: "Adults", Status: "active"}

	deps := GetMemberListDeps{
		MemberStore: &mockGetMemberListMemberStore{members: []domainMember.Member{m1}},
		InjuryStore: &mockGetMemberListInjuryStore{},
		// GradingRecordStore nil
	}

	res, err := QueryGetMemberList(context.Background(), GetMemberListQuery{}, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Members) != 1 {
		t.Fatalf("members=%d want 1", len(res.Members))
	}
	if res.Members[0].Belt != "" || res.Members[0].Stripe != 0 {
		t.Fatalf("belt/stripe=%q/%d want empty/0", res.Members[0].Belt, res.Members[0].Stripe)
	}
}
