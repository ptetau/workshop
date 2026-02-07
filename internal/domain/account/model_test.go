package account_test

import (
	"testing"
	"time"

	"workshop/internal/domain/account"
)

// TestAccount_Validate tests validation of Account.
func TestAccount_Validate(t *testing.T) {
	tests := []struct {
		name    string
		account account.Account
		wantErr bool
	}{
		{
			name: "valid admin account",
			account: account.Account{
				ID:    "1",
				Email: "admin@workshop.co.nz",
				Role:  account.RoleAdmin,
			},
			wantErr: false,
		},
		{
			name: "valid coach account",
			account: account.Account{
				ID:    "2",
				Email: "coach@workshop.co.nz",
				Role:  account.RoleCoach,
			},
			wantErr: false,
		},
		{
			name: "valid member account",
			account: account.Account{
				ID:    "3",
				Email: "member@workshop.co.nz",
				Role:  account.RoleMember,
			},
			wantErr: false,
		},
		{
			name: "valid trial account",
			account: account.Account{
				ID:    "4",
				Email: "trial@workshop.co.nz",
				Role:  account.RoleTrial,
			},
			wantErr: false,
		},
		{
			name: "valid guest account",
			account: account.Account{
				ID:    "5",
				Email: "guest@workshop.co.nz",
				Role:  account.RoleGuest,
			},
			wantErr: false,
		},
		{
			name: "empty email",
			account: account.Account{
				ID:   "6",
				Role: account.RoleAdmin,
			},
			wantErr: true,
		},
		{
			name: "invalid email no at sign",
			account: account.Account{
				ID:    "7",
				Email: "not-an-email",
				Role:  account.RoleAdmin,
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			account: account.Account{
				ID:    "8",
				Email: "user@workshop.co.nz",
				Role:  "superadmin",
			},
			wantErr: true,
		},
		{
			name: "empty role",
			account: account.Account{
				ID:    "9",
				Email: "user@workshop.co.nz",
				Role:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.account.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Account.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAccount_SetPassword tests the SetPassword method.
func TestAccount_SetPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "securepassword123", false},
		{"exactly 12 chars", "123456789012", false},
		{"empty password", "", true},
		{"too short", "short", true},
		{"11 chars", "12345678901", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &account.Account{}
			err := a.SetPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && a.PasswordHash == "" {
				t.Error("SetPassword() should set PasswordHash")
			}
			if err == nil && a.PasswordHash == tt.password {
				t.Error("SetPassword() should hash the password, not store plaintext")
			}
		})
	}
}

// TestAccount_CheckPassword tests the CheckPassword method.
func TestAccount_CheckPassword(t *testing.T) {
	a := &account.Account{}
	if err := a.SetPassword("securepassword123"); err != nil {
		t.Fatalf("SetPassword() failed: %v", err)
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"correct password", "securepassword123", false},
		{"wrong password", "wrongpassword123", true},
		{"empty password", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := a.CheckPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAccount_CheckPassword_NoHash tests CheckPassword with no hash set.
func TestAccount_CheckPassword_NoHash(t *testing.T) {
	a := &account.Account{}
	err := a.CheckPassword("anypassword1234")
	if err == nil {
		t.Error("CheckPassword() should fail when no hash is set")
	}
}

// TestAccount_IsLocked tests the IsLocked method.
func TestAccount_IsLocked(t *testing.T) {
	t.Run("not locked", func(t *testing.T) {
		a := &account.Account{}
		if a.IsLocked() {
			t.Error("new account should not be locked")
		}
	})

	t.Run("locked", func(t *testing.T) {
		a := &account.Account{
			LockedUntil: time.Now().Add(10 * time.Minute),
		}
		if !a.IsLocked() {
			t.Error("account with future LockedUntil should be locked")
		}
	})

	t.Run("lock expired", func(t *testing.T) {
		a := &account.Account{
			LockedUntil: time.Now().Add(-1 * time.Minute),
		}
		if a.IsLocked() {
			t.Error("account with past LockedUntil should not be locked")
		}
	})
}

// TestAccount_RecordFailedLogin tests the RecordFailedLogin method.
func TestAccount_RecordFailedLogin(t *testing.T) {
	a := &account.Account{}

	// First 4 failures should not lock
	for i := 0; i < 4; i++ {
		a.RecordFailedLogin()
		if a.IsLocked() {
			t.Errorf("account should not be locked after %d failures", i+1)
		}
	}

	// 5th failure should lock
	a.RecordFailedLogin()
	if !a.IsLocked() {
		t.Error("account should be locked after 5 failures")
	}
	if a.FailedLogins != 5 {
		t.Errorf("FailedLogins = %d, want 5", a.FailedLogins)
	}
}

// TestAccount_ResetFailedLogins tests the ResetFailedLogins method.
func TestAccount_ResetFailedLogins(t *testing.T) {
	a := &account.Account{
		FailedLogins: 5,
		LockedUntil:  time.Now().Add(15 * time.Minute),
	}

	a.ResetFailedLogins()

	if a.FailedLogins != 0 {
		t.Errorf("FailedLogins = %d, want 0", a.FailedLogins)
	}
	if a.IsLocked() {
		t.Error("account should not be locked after reset")
	}
}

// TestAccount_RoleChecks tests IsAdmin and IsCoachOrAdmin.
func TestAccount_RoleChecks(t *testing.T) {
	tests := []struct {
		role           string
		isAdmin        bool
		isCoachOrAdmin bool
	}{
		{account.RoleAdmin, true, true},
		{account.RoleCoach, false, true},
		{account.RoleMember, false, false},
		{account.RoleTrial, false, false},
		{account.RoleGuest, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			a := &account.Account{Role: tt.role}
			if a.IsAdmin() != tt.isAdmin {
				t.Errorf("IsAdmin() = %v, want %v", a.IsAdmin(), tt.isAdmin)
			}
			if a.IsCoachOrAdmin() != tt.isCoachOrAdmin {
				t.Errorf("IsCoachOrAdmin() = %v, want %v", a.IsCoachOrAdmin(), tt.isCoachOrAdmin)
			}
		})
	}
}
