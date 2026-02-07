package orchestrators

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"workshop/internal/domain/account"
	"workshop/internal/domain/kiosk"

	"github.com/google/uuid"
)

// KioskAccountStore defines the store interface needed by kiosk orchestrators.
type KioskAccountStore interface {
	GetByID(ctx context.Context, id string) (account.Account, error)
}

// LaunchKioskInput carries input for launching kiosk mode.
type LaunchKioskInput struct {
	AccountID string
}

// LaunchKioskDeps holds dependencies for LaunchKiosk.
type LaunchKioskDeps struct {
	AccountStore KioskAccountStore
}

// ExecuteLaunchKiosk creates a new kiosk session tied to the launching account.
// PRE: AccountID must be non-empty and belong to an admin or coach
// POST: Returns a new active kiosk Session
func ExecuteLaunchKiosk(ctx context.Context, input LaunchKioskInput, deps LaunchKioskDeps) (kiosk.Session, error) {
	if input.AccountID == "" {
		return kiosk.Session{}, errors.New("account ID is required")
	}

	acct, err := deps.AccountStore.GetByID(ctx, input.AccountID)
	if err != nil {
		return kiosk.Session{}, errors.New("account not found")
	}

	if !acct.IsCoachOrAdmin() {
		return kiosk.Session{}, errors.New("only admin or coach can launch kiosk mode")
	}

	session := kiosk.Session{
		ID:        uuid.New().String(),
		AccountID: input.AccountID,
		StartedAt: time.Now(),
	}

	if err := session.Validate(); err != nil {
		return kiosk.Session{}, err
	}

	slog.Info("kiosk_event", "event", "kiosk_launched", "account_id", input.AccountID)
	return session, nil
}

// ExitKioskInput carries input for exiting kiosk mode.
type ExitKioskInput struct {
	AccountID string
	Password  string
}

// ExitKioskDeps holds dependencies for ExitKiosk.
type ExitKioskDeps struct {
	AccountStore KioskAccountStore
}

// ExecuteExitKiosk verifies the password to exit kiosk mode.
// PRE: AccountID and Password must be non-empty
// POST: Returns nil if password is correct, error otherwise
func ExecuteExitKiosk(ctx context.Context, input ExitKioskInput, deps ExitKioskDeps) error {
	if input.AccountID == "" {
		return errors.New("account ID is required")
	}
	if input.Password == "" {
		return errors.New("password is required to exit kiosk mode")
	}

	acct, err := deps.AccountStore.GetByID(ctx, input.AccountID)
	if err != nil {
		return errors.New("account not found")
	}

	if err := acct.CheckPassword(input.Password); err != nil {
		return errors.New("invalid password")
	}

	slog.Info("kiosk_event", "event", "kiosk_exited", "account_id", input.AccountID)
	return nil
}
