package orchestrators

import (
	"context"
	"fmt"
	"testing"

	"workshop/internal/domain/account"
	"workshop/internal/domain/member"
)

// --- in-memory test doubles ---

type memTestAcctStore struct {
	accounts map[string]account.Account // keyed by email
}

func newMemTestAcctStore() *memTestAcctStore {
	return &memTestAcctStore{accounts: make(map[string]account.Account)}
}

// Save persists an account in memory.
// PRE: account has valid email
// POST: account is stored in memory map
func (s *memTestAcctStore) Save(_ context.Context, a account.Account) error {
	s.accounts[a.Email] = a
	return nil
}

// GetByEmail retrieves an account by email from memory.
// PRE: email is non-empty
// POST: returns account or error if not found
func (s *memTestAcctStore) GetByEmail(_ context.Context, email string) (account.Account, error) {
	a, ok := s.accounts[email]
	if !ok {
		return account.Account{}, fmt.Errorf("not found")
	}
	return a, nil
}

type memTestMemberStore struct {
	members []member.Member
}

// Save persists a member in memory.
// PRE: member has valid fields
// POST: member is appended to slice
func (s *memTestMemberStore) Save(_ context.Context, m member.Member) error {
	s.members = append(s.members, m)
	return nil
}

// --- tests ---

// TestSeedTestAccounts_CreatesAllAccounts verifies all 4 test accounts are created with correct roles.
func TestSeedTestAccounts_CreatesAllAccounts(t *testing.T) {
	acctStore := newMemTestAcctStore()
	memberStore := &memTestMemberStore{}
	deps := TestAccountSeedDeps{AccountStore: acctStore, MemberStore: memberStore}

	if err := ExecuteSeedTestAccounts(context.Background(), deps); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create 4 accounts
	if len(acctStore.accounts) != 4 {
		t.Errorf("expected 4 accounts, got %d", len(acctStore.accounts))
	}

	// Verify roles
	expected := map[string]string{
		"info+admin@workshopjiujitsu.co.nz":  account.RoleAdmin,
		"info+coach@workshopjiujitsu.co.nz":  account.RoleCoach,
		"info+member@workshopjiujitsu.co.nz": account.RoleMember,
		"info+trial@workshopjiujitsu.co.nz":  account.RoleTrial,
	}
	for email, role := range expected {
		acct, ok := acctStore.accounts[email]
		if !ok {
			t.Errorf("account %s not found", email)
			continue
		}
		if acct.Role != role {
			t.Errorf("account %s: expected role %s, got %s", email, role, acct.Role)
		}
	}
}

// TestSeedTestAccounts_Idempotent verifies running seed twice creates no duplicates.
func TestSeedTestAccounts_Idempotent(t *testing.T) {
	acctStore := newMemTestAcctStore()
	memberStore := &memTestMemberStore{}
	deps := TestAccountSeedDeps{AccountStore: acctStore, MemberStore: memberStore}

	// Seed twice
	if err := ExecuteSeedTestAccounts(context.Background(), deps); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := ExecuteSeedTestAccounts(context.Background(), deps); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	// Still only 4 accounts
	if len(acctStore.accounts) != 4 {
		t.Errorf("expected 4 accounts after double seed, got %d", len(acctStore.accounts))
	}

	// Only 3 member records (not 6 — second run should skip)
	if len(memberStore.members) != 3 {
		t.Errorf("expected 3 members after double seed, got %d", len(memberStore.members))
	}
}

// TestSeedTestAccounts_PasswordsValidate verifies each test account's password is correct.
func TestSeedTestAccounts_PasswordsValidate(t *testing.T) {
	acctStore := newMemTestAcctStore()
	memberStore := &memTestMemberStore{}
	deps := TestAccountSeedDeps{AccountStore: acctStore, MemberStore: memberStore}

	if err := ExecuteSeedTestAccounts(context.Background(), deps); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify each account's password
	for _, def := range testAccounts() {
		acct, ok := acctStore.accounts[def.Email]
		if !ok {
			t.Errorf("account %s not found", def.Email)
			continue
		}
		if err := acct.CheckPassword(def.Password); err != nil {
			t.Errorf("account %s: password check failed: %v", def.Email, err)
		}
	}
}

// TestSeedTestAccounts_MemberRecords verifies member records exist for non-admin accounts.
func TestSeedTestAccounts_MemberRecords(t *testing.T) {
	acctStore := newMemTestAcctStore()
	memberStore := &memTestMemberStore{}
	deps := TestAccountSeedDeps{AccountStore: acctStore, MemberStore: memberStore}

	if err := ExecuteSeedTestAccounts(context.Background(), deps); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create 3 member records (no member for admin)
	if len(memberStore.members) != 3 {
		t.Fatalf("expected 3 members, got %d", len(memberStore.members))
	}

	names := map[string]bool{}
	for _, m := range memberStore.members {
		names[m.Name] = true
		if m.Program != member.ProgramAdults {
			t.Errorf("member %s: expected program adults, got %s", m.Name, m.Program)
		}
		if m.Status != member.StatusActive {
			t.Errorf("member %s: expected status active, got %s", m.Name, m.Status)
		}
		if m.AccountID == "" {
			t.Errorf("member %s: missing AccountID", m.Name)
		}
	}

	for _, name := range []string{"Test Coach", "Test Member", "Test Trial"} {
		if !names[name] {
			t.Errorf("missing member record for %s", name)
		}
	}
}
