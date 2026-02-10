package account

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Max length constants for user-editable fields.
const (
	MaxEmailLength = 254
)

// Role constants
const (
	RoleAdmin  = "admin"
	RoleCoach  = "coach"
	RoleMember = "member"
	RoleTrial  = "trial"
	RoleGuest  = "guest"
)

// Account status constants
const (
	StatusActive            = "active"
	StatusPendingActivation = "pending_activation"
)

// ValidRoles contains all valid role values.
var ValidRoles = []string{RoleAdmin, RoleCoach, RoleMember, RoleTrial, RoleGuest}

// Domain errors
var (
	ErrInvalidEmail     = errors.New("email must contain '@'")
	ErrEmptyEmail       = errors.New("email cannot be empty")
	ErrInvalidRole      = errors.New("role must be one of: admin, coach, member, trial, guest")
	ErrEmptyPassword    = errors.New("password cannot be empty")
	ErrPasswordTooShort = errors.New("password must be at least 12 characters")
	ErrWrongPassword    = errors.New("incorrect password")
	ErrTokenExpired     = errors.New("activation link has expired")
	ErrTokenInvalid     = errors.New("activation token is invalid")
	ErrAlreadyActivated = errors.New("account is already activated")
	ErrNotPending       = errors.New("account is not pending activation")
)

// Account holds state for the Account concept.
type Account struct {
	ID                     string
	Email                  string
	PasswordHash           string
	Role                   string
	Status                 string // active, pending_activation
	CreatedAt              time.Time
	FailedLogins           int
	LockedUntil            time.Time
	PasswordChangeRequired bool
}

// ActivationToken represents a time-limited token for account activation.
type ActivationToken struct {
	ID        string
	AccountID string
	Token     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

// Validate checks if the Account has valid data.
// PRE: Account struct is populated
// POST: Returns nil if valid, error otherwise
func (a *Account) Validate() error {
	if strings.TrimSpace(a.Email) == "" {
		return ErrEmptyEmail
	}
	if len(a.Email) > MaxEmailLength {
		return errors.New("email cannot exceed 254 characters")
	}
	if !strings.Contains(a.Email, "@") {
		return ErrInvalidEmail
	}
	if !isValidRole(a.Role) {
		return ErrInvalidRole
	}
	return nil
}

// SetPassword hashes and stores a password using bcrypt with cost 12.
// PRE: plaintext is non-empty and >= 12 characters
// POST: PasswordHash is set to bcrypt hash
func (a *Account) SetPassword(plaintext string) error {
	if plaintext == "" {
		return ErrEmptyPassword
	}
	if len(plaintext) < 12 {
		return ErrPasswordTooShort
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}
	a.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies a plaintext password against the stored hash.
// PRE: PasswordHash is set
// INVARIANT: Account fields are not mutated
func (a *Account) CheckPassword(plaintext string) error {
	if a.PasswordHash == "" {
		return ErrWrongPassword
	}
	err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(plaintext))
	if err != nil {
		return ErrWrongPassword
	}
	return nil
}

// IsLocked returns true if the account is currently locked out.
// INVARIANT: Account fields are not mutated
func (a *Account) IsLocked() bool {
	if a.LockedUntil.IsZero() {
		return false
	}
	return time.Now().Before(a.LockedUntil)
}

// RecordFailedLogin increments the failed login counter and locks the account after 5 failures.
// PRE: Account exists
// POST: FailedLogins incremented; LockedUntil set if >= 5 failures
func (a *Account) RecordFailedLogin() {
	a.FailedLogins++
	if a.FailedLogins >= 5 {
		a.LockedUntil = time.Now().Add(15 * time.Minute)
	}
}

// ResetFailedLogins clears the failed login counter and lock.
// PRE: Account exists
// POST: FailedLogins is 0, LockedUntil is zero
func (a *Account) ResetFailedLogins() {
	a.FailedLogins = 0
	a.LockedUntil = time.Time{}
}

// IsAdmin returns true if the account has admin role.
// INVARIANT: Account fields are not mutated
func (a *Account) IsAdmin() bool {
	return a.Role == RoleAdmin
}

// IsCoachOrAdmin returns true if the account has coach or admin role.
// INVARIANT: Account fields are not mutated
func (a *Account) IsCoachOrAdmin() bool {
	return a.Role == RoleAdmin || a.Role == RoleCoach
}

// IsPendingActivation returns true if the account is pending activation.
// INVARIANT: Account fields are not mutated
func (a *Account) IsPendingActivation() bool {
	return a.Status == StatusPendingActivation
}

// Activate transitions the account from pending to active.
// PRE: Account is in pending_activation status
// POST: Status is set to active
func (a *Account) Activate() error {
	if a.Status == StatusActive {
		return ErrAlreadyActivated
	}
	if a.Status != StatusPendingActivation {
		return ErrNotPending
	}
	a.Status = StatusActive
	return nil
}

// IsExpired returns true if the activation token has expired.
// INVARIANT: Token fields are not mutated
func (t *ActivationToken) IsExpired(now time.Time) bool {
	return now.After(t.ExpiresAt)
}

// Invalidate marks the token as used.
// PRE: Token exists
// POST: Used is set to true
func (t *ActivationToken) Invalidate() {
	t.Used = true
}

func isValidRole(role string) bool {
	for _, r := range ValidRoles {
		if r == role {
			return true
		}
	}
	return false
}
