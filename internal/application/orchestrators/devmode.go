package orchestrators

import (
	"errors"

	"workshop/internal/domain/account"
)

// DevMode errors
var (
	ErrDevModeNotAdmin         = errors.New("only admins can use devmode impersonation")
	ErrDevModeInvalidRole      = errors.New("target role is not valid")
	ErrDevModeNotImpersonating = errors.New("not currently impersonating")
	ErrDevModeAlreadyAdmin     = errors.New("already operating as admin")
)

// DevModeImpersonateInput carries input for the impersonate orchestrator.
type DevModeImpersonateInput struct {
	TargetRole    string
	CurrentRole   string
	AccountID     string
	Email         string
	RealAccountID string // non-empty if already impersonating
	RealRole      string // non-empty if already impersonating
	RealEmail     string // non-empty if already impersonating
}

// DevModeImpersonateResult carries the updated session fields.
type DevModeImpersonateResult struct {
	Role          string
	RealAccountID string
	RealEmail     string
	RealRole      string
}

// ExecuteDevModeImpersonate validates the impersonation request and returns updated session fields.
// PRE: Caller must be a real admin (either directly or via RealRole if already impersonating).
// POST: Returns new session fields with the target role and preserved admin identity.
func ExecuteDevModeImpersonate(input DevModeImpersonateInput) (DevModeImpersonateResult, error) {
	// Determine the real identity
	realRole := input.CurrentRole
	realAccountID := input.AccountID
	realEmail := input.Email
	if input.RealRole != "" {
		// Already impersonating — use the stashed real identity
		realRole = input.RealRole
		realAccountID = input.RealAccountID
		realEmail = input.RealEmail
	}

	// Only real admins can impersonate
	if realRole != account.RoleAdmin {
		return DevModeImpersonateResult{}, ErrDevModeNotAdmin
	}

	// Validate target role against allowlist
	if !isValidTargetRole(input.TargetRole) {
		return DevModeImpersonateResult{}, ErrDevModeInvalidRole
	}

	// If switching back to admin, clear impersonation
	if input.TargetRole == account.RoleAdmin {
		return DevModeImpersonateResult{
			Role:          account.RoleAdmin,
			RealAccountID: "",
			RealEmail:     "",
			RealRole:      "",
		}, nil
	}

	return DevModeImpersonateResult{
		Role:          input.TargetRole,
		RealAccountID: realAccountID,
		RealEmail:     realEmail,
		RealRole:      account.RoleAdmin,
	}, nil
}

// DevModeRestoreInput carries input for the restore orchestrator.
type DevModeRestoreInput struct {
	CurrentRole   string
	RealAccountID string
	RealEmail     string
	RealRole      string
}

// DevModeRestoreResult carries the restored session fields.
type DevModeRestoreResult struct {
	AccountID string
	Email     string
	Role      string
}

// ExecuteDevModeRestore validates the restore request and returns the original admin session fields.
// PRE: Caller must be currently impersonating with a real admin identity.
// POST: Returns original admin session fields; impersonation fields should be cleared.
func ExecuteDevModeRestore(input DevModeRestoreInput) (DevModeRestoreResult, error) {
	if input.RealRole == "" {
		return DevModeRestoreResult{}, ErrDevModeNotImpersonating
	}
	if input.RealRole != account.RoleAdmin {
		return DevModeRestoreResult{}, ErrDevModeNotAdmin
	}

	return DevModeRestoreResult{
		AccountID: input.RealAccountID,
		Email:     input.RealEmail,
		Role:      account.RoleAdmin,
	}, nil
}

func isValidTargetRole(role string) bool {
	for _, r := range account.ValidRoles {
		if r == role {
			return true
		}
	}
	return false
}
