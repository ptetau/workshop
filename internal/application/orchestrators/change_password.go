package orchestrators

import (
	"context"
	"errors"
	"log/slog"
	"workshop/internal/domain/account"
)

// ChangePasswordInput carries input for the change-password orchestrator.
type ChangePasswordInput struct {
	AccountID       string
	CurrentPassword string
	NewPassword     string
}

// AccountStoreForChangePassword defines the store interface needed by ChangePassword.
type AccountStoreForChangePassword interface {
	GetByID(ctx context.Context, id string) (account.Account, error)
	Save(ctx context.Context, a account.Account) error
}

// ChangePasswordDeps holds dependencies for ChangePassword.
type ChangePasswordDeps struct {
	AccountStore AccountStoreForChangePassword
}

var (
	ErrCurrentPasswordWrong = errors.New("current password is incorrect")
	ErrNewPasswordSame      = errors.New("new password must be different from current password")
)

// ExecuteChangePassword validates the current password and updates to the new one.
// PRE: AccountID is valid, both passwords are non-empty
// POST: Password is updated, PasswordChangeRequired is cleared
func ExecuteChangePassword(ctx context.Context, input ChangePasswordInput, deps ChangePasswordDeps) error {
	if input.AccountID == "" || input.CurrentPassword == "" || input.NewPassword == "" {
		return errors.New("all fields are required")
	}

	acct, err := deps.AccountStore.GetByID(ctx, input.AccountID)
	if err != nil {
		return errors.New("account not found")
	}

	// Verify current password
	if err := acct.CheckPassword(input.CurrentPassword); err != nil {
		return ErrCurrentPasswordWrong
	}

	// Ensure new password is different
	if input.CurrentPassword == input.NewPassword {
		return ErrNewPasswordSame
	}

	// Set new password (validates length, hashes)
	if err := acct.SetPassword(input.NewPassword); err != nil {
		return err
	}

	// Clear the forced change flag
	acct.PasswordChangeRequired = false

	if err := deps.AccountStore.Save(ctx, acct); err != nil {
		return err
	}

	slog.Info("auth_event", "event", "password_changed", "account_id", input.AccountID)
	return nil
}
