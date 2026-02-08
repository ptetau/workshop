package account

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Role constants
const (
	RoleAdmin  = "admin"
	RoleCoach  = "coach"
	RoleMember = "member"
	RoleTrial  = "trial"
	RoleGuest  = "guest"
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
)

// Account holds state for the Account concept.
type Account struct {
	ID                     string
	Email                  string
	PasswordHash           string
	Role                   string
	CreatedAt              time.Time
	FailedLogins           int
	LockedUntil            time.Time
	PasswordChangeRequired bool
}

// Validate checks if the Account has valid data.
// PRE: Account struct is populated
// POST: Returns nil if valid, error otherwise
func (a *Account) Validate() error {
	if strings.TrimSpace(a.Email) == "" {
		return ErrEmptyEmail
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

func isValidRole(role string) bool {
	for _, r := range ValidRoles {
		if r == role {
			return true
		}
	}
	return false
}
