package orchestrators

import (
	"testing"

	"workshop/internal/domain/account"
)

// TestExecuteDevModeImpersonate_ValidRole tests impersonation with a valid target role.
func TestExecuteDevModeImpersonate_ValidRole(t *testing.T) {
	input := DevModeImpersonateInput{
		TargetRole:  account.RoleCoach,
		CurrentRole: account.RoleAdmin,
		AccountID:   "admin-001",
		Email:       "admin@workshop.co.nz",
	}

	result, err := ExecuteDevModeImpersonate(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Role != account.RoleCoach {
		t.Errorf("expected role %q, got %q", account.RoleCoach, result.Role)
	}
	if result.RealAccountID != "admin-001" {
		t.Errorf("expected RealAccountID %q, got %q", "admin-001", result.RealAccountID)
	}
	if result.RealRole != account.RoleAdmin {
		t.Errorf("expected RealRole %q, got %q", account.RoleAdmin, result.RealRole)
	}
	if result.RealEmail != "admin@workshop.co.nz" {
		t.Errorf("expected RealEmail %q, got %q", "admin@workshop.co.nz", result.RealEmail)
	}
}

// TestExecuteDevModeImpersonate_InvalidRole tests impersonation with an invalid target role.
func TestExecuteDevModeImpersonate_InvalidRole(t *testing.T) {
	input := DevModeImpersonateInput{
		TargetRole:  "superuser",
		CurrentRole: account.RoleAdmin,
		AccountID:   "admin-001",
		Email:       "admin@workshop.co.nz",
	}

	_, err := ExecuteDevModeImpersonate(input)
	if err != ErrDevModeInvalidRole {
		t.Errorf("expected ErrDevModeInvalidRole, got %v", err)
	}
}

// TestExecuteDevModeImpersonate_NonAdminCaller tests that non-admin callers are rejected.
func TestExecuteDevModeImpersonate_NonAdminCaller(t *testing.T) {
	input := DevModeImpersonateInput{
		TargetRole:  account.RoleCoach,
		CurrentRole: account.RoleMember,
		AccountID:   "member-001",
		Email:       "member@workshop.co.nz",
	}

	_, err := ExecuteDevModeImpersonate(input)
	if err != ErrDevModeNotAdmin {
		t.Errorf("expected ErrDevModeNotAdmin, got %v", err)
	}
}

// TestExecuteDevModeImpersonate_SwitchBackToAdmin tests switching back to admin from impersonation.
func TestExecuteDevModeImpersonate_SwitchBackToAdmin(t *testing.T) {
	input := DevModeImpersonateInput{
		TargetRole:    account.RoleAdmin,
		CurrentRole:   account.RoleCoach,
		AccountID:     "admin-001",
		Email:         "admin@workshop.co.nz",
		RealAccountID: "admin-001",
		RealEmail:     "admin@workshop.co.nz",
		RealRole:      account.RoleAdmin,
	}

	result, err := ExecuteDevModeImpersonate(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Role != account.RoleAdmin {
		t.Errorf("expected role %q, got %q", account.RoleAdmin, result.Role)
	}
	if result.RealRole != "" {
		t.Errorf("expected empty RealRole after restore, got %q", result.RealRole)
	}
	if result.RealAccountID != "" {
		t.Errorf("expected empty RealAccountID after restore, got %q", result.RealAccountID)
	}
}

// TestExecuteDevModeImpersonate_AlreadyImpersonating_SwitchRole tests switching roles while already impersonating.
func TestExecuteDevModeImpersonate_AlreadyImpersonating_SwitchRole(t *testing.T) {
	input := DevModeImpersonateInput{
		TargetRole:    account.RoleMember,
		CurrentRole:   account.RoleCoach,
		AccountID:     "admin-001",
		Email:         "admin@workshop.co.nz",
		RealAccountID: "admin-001",
		RealEmail:     "admin@workshop.co.nz",
		RealRole:      account.RoleAdmin,
	}

	result, err := ExecuteDevModeImpersonate(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Role != account.RoleMember {
		t.Errorf("expected role %q, got %q", account.RoleMember, result.Role)
	}
	if result.RealRole != account.RoleAdmin {
		t.Errorf("expected RealRole %q, got %q", account.RoleAdmin, result.RealRole)
	}
}

// TestExecuteDevModeRestore_Success tests successful restore from impersonation.
func TestExecuteDevModeRestore_Success(t *testing.T) {
	input := DevModeRestoreInput{
		CurrentRole:   account.RoleCoach,
		RealAccountID: "admin-001",
		RealEmail:     "admin@workshop.co.nz",
		RealRole:      account.RoleAdmin,
	}

	result, err := ExecuteDevModeRestore(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Role != account.RoleAdmin {
		t.Errorf("expected role %q, got %q", account.RoleAdmin, result.Role)
	}
	if result.AccountID != "admin-001" {
		t.Errorf("expected AccountID %q, got %q", "admin-001", result.AccountID)
	}
	if result.Email != "admin@workshop.co.nz" {
		t.Errorf("expected Email %q, got %q", "admin@workshop.co.nz", result.Email)
	}
}

// TestExecuteDevModeRestore_NotImpersonating tests restore when not impersonating.
func TestExecuteDevModeRestore_NotImpersonating(t *testing.T) {
	input := DevModeRestoreInput{
		CurrentRole: account.RoleAdmin,
	}

	_, err := ExecuteDevModeRestore(input)
	if err != ErrDevModeNotImpersonating {
		t.Errorf("expected ErrDevModeNotImpersonating, got %v", err)
	}
}

// TestExecuteDevModeRestore_NonAdminRealRole tests restore when real role is not admin.
func TestExecuteDevModeRestore_NonAdminRealRole(t *testing.T) {
	input := DevModeRestoreInput{
		CurrentRole:   account.RoleMember,
		RealAccountID: "coach-001",
		RealEmail:     "coach@workshop.co.nz",
		RealRole:      account.RoleCoach,
	}

	_, err := ExecuteDevModeRestore(input)
	if err != ErrDevModeNotAdmin {
		t.Errorf("expected ErrDevModeNotAdmin, got %v", err)
	}
}
