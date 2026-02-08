package orchestrators

import (
	"context"
	"errors"
	"log/slog"

	"workshop/internal/domain/account"
)

// AccountStoreForLogin defines the store interface needed by Login.
type AccountStoreForLogin interface {
	GetByEmail(ctx context.Context, email string) (account.Account, error)
	Save(ctx context.Context, a account.Account) error
}

// LoginInput carries input for the login orchestrator.
type LoginInput struct {
	Email    string
	Password string
}

// LoginResult carries the result of a successful login.
type LoginResult struct {
	AccountID              string
	Email                  string
	Role                   string
	PasswordChangeRequired bool
}

// LoginDeps holds dependencies for Login.
type LoginDeps struct {
	AccountStore AccountStoreForLogin
}

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is locked due to too many failed attempts")
	ErrPendingActivation  = errors.New("account is pending activation — check your email for the activation link")
)

// ExecuteLogin validates credentials and returns account info for session creation.
// PRE: Valid email and password provided
// POST: Returns account info on success, records failed login on failure
// INVARIANT: Account must not be locked
func ExecuteLogin(ctx context.Context, input LoginInput, deps LoginDeps) (LoginResult, error) {
	if input.Email == "" || input.Password == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	acct, err := deps.AccountStore.GetByEmail(ctx, input.Email)
	if err != nil {
		slog.Info("auth_event", "event", "login_failed", "email", input.Email, "reason", "not_found")
		return LoginResult{}, ErrInvalidCredentials
	}

	// Check if account is pending activation
	if acct.IsPendingActivation() {
		slog.Info("auth_event", "event", "login_blocked", "email", input.Email, "reason", "pending_activation")
		return LoginResult{}, ErrPendingActivation
	}

	// Check if account is locked
	if acct.IsLocked() {
		slog.Info("auth_event", "event", "login_blocked", "email", input.Email, "reason", "locked")
		return LoginResult{}, ErrAccountLocked
	}

	// Verify password
	if err := acct.CheckPassword(input.Password); err != nil {
		acct.RecordFailedLogin()
		_ = deps.AccountStore.Save(ctx, acct)
		slog.Info("auth_event", "event", "login_failed", "email", input.Email, "reason", "wrong_password", "failed_logins", acct.FailedLogins)
		return LoginResult{}, ErrInvalidCredentials
	}

	// Successful login — reset failed attempts
	acct.ResetFailedLogins()
	_ = deps.AccountStore.Save(ctx, acct)

	slog.Info("auth_event", "event", "login_success", "email", input.Email, "role", acct.Role)

	return LoginResult{
		AccountID:              acct.ID,
		Email:                  acct.Email,
		Role:                   acct.Role,
		PasswordChangeRequired: acct.PasswordChangeRequired,
	}, nil
}
