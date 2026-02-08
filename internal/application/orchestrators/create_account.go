package orchestrators

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"workshop/internal/domain/account"

	"github.com/google/uuid"
)

// AccountStoreForCreate defines the store interface needed by CreateAccount.
type AccountStoreForCreate interface {
	GetByEmail(ctx context.Context, email string) (account.Account, error)
	Save(ctx context.Context, a account.Account) error
	Count(ctx context.Context) (int, error)
}

// CreateAccountInput carries input for the orchestrator.
type CreateAccountInput struct {
	Email                  string
	Password               string
	Role                   string
	PasswordChangeRequired bool
}

// CreateAccountDeps holds dependencies for CreateAccount.
type CreateAccountDeps struct {
	AccountStore AccountStoreForCreate
}

var ErrEmailAlreadyExists = errors.New("an account with this email already exists")

// ExecuteCreateAccount coordinates account creation.
// PRE: Valid email, password >= 12 chars, valid role
// POST: Account created with hashed password
// INVARIANT: Email must be unique
func ExecuteCreateAccount(ctx context.Context, input CreateAccountInput, deps CreateAccountDeps) (string, error) {
	if input.Email == "" {
		return "", errors.New("email cannot be empty")
	}
	if input.Password == "" {
		return "", errors.New("password cannot be empty")
	}
	if input.Role == "" {
		return "", errors.New("role cannot be empty")
	}

	// Check if email already exists
	_, err := deps.AccountStore.GetByEmail(ctx, input.Email)
	if err == nil {
		return "", ErrEmailAlreadyExists
	}

	acct := account.Account{
		ID:                     uuid.New().String(),
		Email:                  input.Email,
		Role:                   input.Role,
		CreatedAt:              time.Now(),
		PasswordChangeRequired: input.PasswordChangeRequired,
	}

	// Validate domain rules
	if err := acct.Validate(); err != nil {
		return "", err
	}

	// Set password (handles hashing and length validation)
	if err := acct.SetPassword(input.Password); err != nil {
		return "", err
	}

	// Save to store
	if err := deps.AccountStore.Save(ctx, acct); err != nil {
		return "", err
	}

	slog.Info("auth_event", "event", "account_created", "email", input.Email, "role", input.Role)

	return acct.ID, nil
}

// ExecuteSeedAdmin creates a default admin account if no accounts exist.
// PRE: Database is initialized
// POST: Admin account created if count == 0
func ExecuteSeedAdmin(ctx context.Context, deps CreateAccountDeps, email, password string) error {
	count, err := deps.AccountStore.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Accounts already exist, skip seeding
	}

	_, err = ExecuteCreateAccount(ctx, CreateAccountInput{
		Email:                  email,
		Password:               password,
		Role:                   account.RoleAdmin,
		PasswordChangeRequired: true,
	}, deps)
	if err != nil {
		return err
	}

	slog.Info("auth_event", "event", "admin_seeded", "email", email)
	return nil
}
