package orchestrators

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	memberStore "workshop/internal/adapters/storage/member"
	domain "workshop/internal/domain/member"
)

// mockMemberStoreForImport implements ImportMembersDeps.MemberStore for testing.
type mockMemberStoreForImport struct {
	byEmail map[string]domain.Member
	byID    map[string]domain.Member
	saveErr error
}

// GetByEmail implements memberStore.Store.
// PRE: email is non-empty
// POST: returns member or error if not found
func (m *mockMemberStoreForImport) GetByEmail(_ context.Context, email string) (domain.Member, error) {
	mem, ok := m.byEmail[strings.ToLower(email)]
	if !ok {
		return domain.Member{}, errors.New("not found")
	}
	return mem, nil
}

// Save implements memberStore.Store.
// PRE: member is valid
// POST: member is persisted by ID and email
func (m *mockMemberStoreForImport) Save(_ context.Context, mem domain.Member) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.byID[mem.ID] = mem
	m.byEmail[strings.ToLower(mem.Email)] = mem
	return nil
}

// GetByID implements memberStore.Store.
// PRE: id is non-empty
// POST: returns member or error if not found
func (m *mockMemberStoreForImport) GetByID(_ context.Context, id string) (domain.Member, error) {
	mem, ok := m.byID[id]
	if !ok {
		return domain.Member{}, errors.New("not found")
	}
	return mem, nil
}

// Delete implements memberStore.Store.
// PRE: id is non-empty
// POST: member is removed
func (m *mockMemberStoreForImport) Delete(_ context.Context, _ string) error { return nil }

// List implements memberStore.Store.
// PRE: filter is valid
// POST: returns all stored members
func (m *mockMemberStoreForImport) List(_ context.Context, _ memberStore.ListFilter) ([]domain.Member, error) {
	return nil, nil
}

// Count implements memberStore.Store.
// PRE: filter is valid
// POST: returns count of stored members
func (m *mockMemberStoreForImport) Count(_ context.Context, _ memberStore.ListFilter) (int, error) {
	return len(m.byID), nil
}

// SearchByName implements memberStore.Store.
// PRE: query is non-empty
// POST: returns matching members
func (m *mockMemberStoreForImport) SearchByName(_ context.Context, _ string, _ int) ([]domain.Member, error) {
	return nil, nil
}

func newMockMemberStoreForImport() *mockMemberStoreForImport {
	return &mockMemberStoreForImport{
		byEmail: make(map[string]domain.Member),
		byID:    make(map[string]domain.Member),
	}
}

func importDeps(store *mockMemberStoreForImport) ImportMembersDeps {
	n := 0
	genID := func() string {
		n++
		return fmt.Sprintf("gen-%d", n)
	}
	return ImportMembersDeps{
		MemberStore: store,
		GenerateID:  genID,
	}
}

// TestExecuteImportMembers_CreatesNewMembers verifies new members are created from valid CSV.
// PRE: empty store, valid CSV with NAME+EMAIL.
// POST: created=2, no errors.
func TestExecuteImportMembers_CreatesNewMembers(t *testing.T) {
	store := newMockMemberStoreForImport()
	csv := "NAME,EMAIL,PROGRAM,STATUS\nAlice,alice@test.com,adults,active\nBob,bob@test.com,kids,active\n"
	result, err := ExecuteImportMembers(context.Background(), ImportMembersInput{
		Reader:         strings.NewReader(csv),
		AdminAccountID: "admin-1",
	}, importDeps(store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created != 2 {
		t.Errorf("created=%d want 2", result.Created)
	}
	if result.Total != 2 {
		t.Errorf("total=%d want 2", result.Total)
	}
	if len(result.Errors) != 0 {
		t.Errorf("errors=%v want none", result.Errors)
	}
}

// TestExecuteImportMembers_SkipsDuplicatesByDefault verifies existing emails are skipped.
// PRE: member with email exists, CSV contains same email.
// POST: skipped=1, created=0, existing member unchanged.
func TestExecuteImportMembers_SkipsDuplicatesByDefault(t *testing.T) {
	store := newMockMemberStoreForImport()
	store.byEmail["alice@test.com"] = domain.Member{ID: "orig-1", Name: "Alice Original", Email: "alice@test.com", Program: "adults", Status: "active"}
	store.byID["orig-1"] = store.byEmail["alice@test.com"]

	csv := "NAME,EMAIL\nAlice Updated,alice@test.com\n"
	result, err := ExecuteImportMembers(context.Background(), ImportMembersInput{
		Reader:         strings.NewReader(csv),
		AdminAccountID: "admin-1",
	}, importDeps(store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("skipped=%d want 1", result.Skipped)
	}
	if result.Created != 0 {
		t.Errorf("created=%d want 0", result.Created)
	}
	if store.byID["orig-1"].Name != "Alice Original" {
		t.Error("original member name should not be changed")
	}
}

// TestExecuteImportMembers_UpdateModePreservesID verifies update_mode upserts preserving ID.
// PRE: member exists, CSV has same email with new name.
// POST: updated=1, ID preserved, name updated.
func TestExecuteImportMembers_UpdateModePreservesID(t *testing.T) {
	store := newMockMemberStoreForImport()
	store.byEmail["alice@test.com"] = domain.Member{ID: "orig-1", Name: "Alice Old", Email: "alice@test.com", Program: "adults", Status: "active"}
	store.byID["orig-1"] = store.byEmail["alice@test.com"]

	csv := "NAME,EMAIL\nAlice New,alice@test.com\n"
	result, err := ExecuteImportMembers(context.Background(), ImportMembersInput{
		Reader:         strings.NewReader(csv),
		AdminAccountID: "admin-1",
		UpdateMode:     true,
	}, importDeps(store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Updated != 1 {
		t.Errorf("updated=%d want 1", result.Updated)
	}
	if store.byEmail["alice@test.com"].ID != "orig-1" {
		t.Error("ID must be preserved on update")
	}
	if store.byEmail["alice@test.com"].Name != "Alice New" {
		t.Error("name should be updated")
	}
}

// TestExecuteImportMembers_DryRunDoesNotWrite verifies dry_run=true returns counts without writing.
// PRE: empty store, valid CSV.
// POST: created=1 in result, store still empty.
func TestExecuteImportMembers_DryRunDoesNotWrite(t *testing.T) {
	store := newMockMemberStoreForImport()
	csv := "NAME,EMAIL\nDry Person,dry@test.com\n"
	result, err := ExecuteImportMembers(context.Background(), ImportMembersInput{
		Reader:         strings.NewReader(csv),
		AdminAccountID: "admin-1",
		DryRun:         true,
	}, importDeps(store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.DryRun {
		t.Error("DryRun should be true in result")
	}
	if result.Created != 1 {
		t.Errorf("created=%d want 1", result.Created)
	}
	if len(store.byID) != 0 {
		t.Error("no members should be written during dry run")
	}
}

// TestExecuteImportMembers_InvalidEmailReported verifies bad emails produce per-row errors.
// PRE: CSV with invalid email.
// POST: errors=1, created=0.
func TestExecuteImportMembers_InvalidEmailReported(t *testing.T) {
	store := newMockMemberStoreForImport()
	csv := "NAME,EMAIL\nBad Person,notanemail\n"
	result, err := ExecuteImportMembers(context.Background(), ImportMembersInput{
		Reader:         strings.NewReader(csv),
		AdminAccountID: "admin-1",
	}, importDeps(store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Errors) != 1 {
		t.Errorf("errors=%d want 1", len(result.Errors))
	}
	if result.Created != 0 {
		t.Errorf("created=%d want 0", result.Created)
	}
}

// TestExecuteImportMembers_MissingNameReported verifies empty name produces per-row error.
// PRE: CSV row with empty NAME.
// POST: errors=1.
func TestExecuteImportMembers_MissingNameReported(t *testing.T) {
	store := newMockMemberStoreForImport()
	csv := "NAME,EMAIL\n,valid@test.com\n"
	result, err := ExecuteImportMembers(context.Background(), ImportMembersInput{
		Reader:         strings.NewReader(csv),
		AdminAccountID: "admin-1",
	}, importDeps(store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Errors) != 1 {
		t.Errorf("errors=%d want 1", len(result.Errors))
	}
}

// TestExecuteImportMembers_MissingRequiredColumn returns validation error for missing NAME column.
// PRE: CSV without NAME column.
// POST: returns ImportMembersValidationError.
func TestExecuteImportMembers_MissingRequiredColumn(t *testing.T) {
	store := newMockMemberStoreForImport()
	csv := "EMAIL\nalice@test.com\n"
	_, err := ExecuteImportMembers(context.Background(), ImportMembersInput{
		Reader:         strings.NewReader(csv),
		AdminAccountID: "admin-1",
	}, importDeps(store))
	if err == nil {
		t.Fatal("expected error for missing NAME column")
	}
	var ve *ImportMembersValidationError
	if !errors.As(err, &ve) {
		t.Errorf("expected ImportMembersValidationError, got %T: %v", err, err)
	}
}

// TestExecuteImportMembers_UnknownColumnsReported verifies unknown columns are listed in result.
// PRE: CSV with extra unknown column.
// POST: Unknown slice contains the extra column name.
func TestExecuteImportMembers_UnknownColumnsReported(t *testing.T) {
	store := newMockMemberStoreForImport()
	csv := "NAME,EMAIL,BELT\nAlice,alice@test.com,blue\n"
	result, err := ExecuteImportMembers(context.Background(), ImportMembersInput{
		Reader:         strings.NewReader(csv),
		AdminAccountID: "admin-1",
	}, importDeps(store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Unknown) != 1 || result.Unknown[0] != "BELT" {
		t.Errorf("unknown=%v want [BELT]", result.Unknown)
	}
}

// TestExecuteImportMembers_SaveErrorReportedPerRow verifies store save errors produce per-row errors.
// PRE: store returns error on Save.
// POST: errors=1, created=0.
func TestExecuteImportMembers_SaveErrorReportedPerRow(t *testing.T) {
	store := newMockMemberStoreForImport()
	store.saveErr = errors.New("disk full")
	csv := "NAME,EMAIL\nAlice,alice@test.com\n"
	result, err := ExecuteImportMembers(context.Background(), ImportMembersInput{
		Reader:         strings.NewReader(csv),
		AdminAccountID: "admin-1",
	}, importDeps(store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Errors) != 1 {
		t.Errorf("errors=%d want 1", len(result.Errors))
	}
	if strings.Contains(result.Errors[0].Message, "disk full") {
		t.Error("internal error detail must not be exposed in row message")
	}
}
