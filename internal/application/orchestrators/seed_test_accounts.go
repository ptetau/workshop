package orchestrators

import (
	"context"
	"fmt"
	"log/slog"

	"workshop/internal/domain/account"
	"workshop/internal/domain/member"

	"github.com/google/uuid"
)

// TestAccountSeedDeps holds stores needed for test account seeding.
type TestAccountSeedDeps struct {
	AccountStore testAcctAccountStore
	MemberStore  testAcctMemberStore
}

type testAcctAccountStore interface {
	Save(ctx context.Context, a account.Account) error
	GetByEmail(ctx context.Context, email string) (account.Account, error)
}

type testAcctMemberStore interface {
	Save(ctx context.Context, m member.Member) error
}

// testAccountDef defines a single test account to seed.
type testAccountDef struct {
	Email      string
	Password   string
	Role       string
	MemberName string
}

// testAccounts returns the list of test accounts to seed.
func testAccounts() []testAccountDef {
	return []testAccountDef{
		{
			Email:      "info+admin@workshopjiujitsu.co.nz",
			Password:   "Umami+admin!",
			Role:       account.RoleAdmin,
			MemberName: "", // admin doesn't need a member record
		},
		{
			Email:      "info+coach@workshopjiujitsu.co.nz",
			Password:   "Umami+coach!",
			Role:       account.RoleCoach,
			MemberName: "Test Coach",
		},
		{
			Email:      "info+member@workshopjiujitsu.co.nz",
			Password:   "Umami+coach!",
			Role:       account.RoleMember,
			MemberName: "Test Member",
		},
		{
			Email:      "info+trial@workshopjiujitsu.co.nz",
			Password:   "Umami+coach!",
			Role:       account.RoleTrial,
			MemberName: "Test Trial",
		},
	}
}

// ExecuteSeedTestAccounts creates test accounts for each role if they don't already exist.
// It is idempotent — skips accounts that already exist (checked by email).
// PRE: Database is migrated, admin seed has run.
// POST: 4 test accounts exist with correct roles; 3 member records for non-admin accounts.
func ExecuteSeedTestAccounts(ctx context.Context, deps TestAccountSeedDeps) error {
	created := 0
	for _, def := range testAccounts() {
		// Check if account already exists
		_, err := deps.AccountStore.GetByEmail(ctx, def.Email)
		if err == nil {
			continue // already exists
		}

		acct := account.Account{
			ID:    uuid.New().String(),
			Email: def.Email,
			Role:  def.Role,
		}
		if err := acct.SetPassword(def.Password); err != nil {
			return fmt.Errorf("seed test account %s: set password: %w", def.Email, err)
		}
		if err := deps.AccountStore.Save(ctx, acct); err != nil {
			return fmt.Errorf("seed test account %s: save: %w", def.Email, err)
		}

		// Create member record for non-admin accounts
		if def.MemberName != "" {
			m := member.Member{
				ID:        uuid.New().String(),
				AccountID: acct.ID,
				Name:      def.MemberName,
				Email:     def.Email,
				Program:   member.ProgramAdults,
				Status:    member.StatusActive,
			}
			if err := deps.MemberStore.Save(ctx, m); err != nil {
				return fmt.Errorf("seed test member %s: save: %w", def.MemberName, err)
			}
		}

		created++
		slog.Info("seed_event", "event", "test_account_created", "email", def.Email, "role", def.Role)
	}

	if created > 0 {
		slog.Info("seed_event", "event", "test_accounts_seeded", "created", created)
	}
	return nil
}
