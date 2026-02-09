package account

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	domain "workshop/internal/domain/account"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new AccountStore.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves an Account by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Account, error) {
	query := "SELECT id, email, password_hash, role, status, created_at, failed_logins, locked_until, password_change_required FROM account WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, id)

	entity, err := scanAccount(row.Scan)
	if err == sql.ErrNoRows {
		return domain.Account{}, fmt.Errorf("account not found: %w", err)
	}
	return entity, err
}

// GetByEmail retrieves an Account by email.
// PRE: email is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByEmail(ctx context.Context, email string) (domain.Account, error) {
	query := "SELECT id, email, password_hash, role, status, created_at, failed_logins, locked_until, password_change_required FROM account WHERE email = ?"
	row := s.db.QueryRowContext(ctx, query, email)

	entity, err := scanAccount(row.Scan)
	if err == sql.ErrNoRows {
		return domain.Account{}, fmt.Errorf("account not found: %w", err)
	}
	return entity, err
}

// Save persists an Account to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, entity domain.Account) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fields := []string{"id", "email", "password_hash", "role", "status", "created_at", "failed_logins", "locked_until", "password_change_required"}
	placeholders := []string{"?", "?", "?", "?", "?", "?", "?", "?", "?"}
	updates := []string{
		"email=excluded.email",
		"password_hash=excluded.password_hash",
		"role=excluded.role",
		"status=excluded.status",
		"failed_logins=excluded.failed_logins",
		"locked_until=excluded.locked_until",
		"password_change_required=excluded.password_change_required",
	}

	query := fmt.Sprintf(
		"INSERT INTO account (%s) VALUES (%s) ON CONFLICT(id) DO UPDATE SET %s",
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updates, ", "),
	)

	var lockedUntil interface{}
	if !entity.LockedUntil.IsZero() {
		lockedUntil = entity.LockedUntil.Format("2006-01-02T15:04:05.999999999Z07:00")
	}

	passwordChangeRequired := 0
	if entity.PasswordChangeRequired {
		passwordChangeRequired = 1
	}

	status := entity.Status
	if status == "" {
		status = domain.StatusActive
	}

	_, err = tx.ExecContext(ctx, query,
		entity.ID,
		entity.Email,
		entity.PasswordHash,
		entity.Role,
		status,
		entity.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
		entity.FailedLogins,
		lockedUntil,
		passwordChangeRequired,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete removes an Account from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM account WHERE id = ?", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// List retrieves Accounts based on the filter.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (s *SQLiteStore) List(ctx context.Context, filter ListFilter) ([]domain.Account, error) {
	var queryBuilder strings.Builder
	var args []interface{}

	queryBuilder.WriteString("SELECT id, email, password_hash, role, status, created_at, failed_logins, locked_until, password_change_required FROM account")

	if filter.Role != "" {
		queryBuilder.WriteString(" WHERE role = ?")
		args = append(args, filter.Role)
	}

	queryBuilder.WriteString(" ORDER BY created_at DESC LIMIT ? OFFSET ?")
	args = append(args, filter.Limit, filter.Offset)

	rows, err := s.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Account
	for rows.Next() {
		entity, err := scanAccount(rows.Scan)
		if err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return results, nil
}

// Count returns the total number of accounts.
// PRE: none
// POST: Returns total account count
func (s *SQLiteStore) Count(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM account").Scan(&count)
	return count, err
}

// scanAccount extracts an Account from a row scanner function.
func scanAccount(scan func(dest ...interface{}) error) (domain.Account, error) {
	var entity domain.Account
	var createdAt string
	var lockedUntil sql.NullString
	var passwordChangeRequired int
	var status sql.NullString
	err := scan(
		&entity.ID,
		&entity.Email,
		&entity.PasswordHash,
		&entity.Role,
		&status,
		&createdAt,
		&entity.FailedLogins,
		&lockedUntil,
		&passwordChangeRequired,
	)
	if err != nil {
		return domain.Account{}, err
	}
	entity.CreatedAt, _ = parseTime(createdAt)
	if lockedUntil.Valid && lockedUntil.String != "" {
		entity.LockedUntil, _ = parseTime(lockedUntil.String)
	}
	entity.PasswordChangeRequired = passwordChangeRequired != 0
	if status.Valid && status.String != "" {
		entity.Status = status.String
	} else {
		entity.Status = domain.StatusActive
	}
	return entity, nil
}

// SaveActivationToken persists an activation token.
// PRE: token has valid ID, AccountID, Token, ExpiresAt
// POST: Token is persisted in database
func (s *SQLiteStore) SaveActivationToken(ctx context.Context, token domain.ActivationToken) error {
	used := 0
	if token.Used {
		used = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO activation_token (id, account_id, token, expires_at, used, created_at) VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET used=excluded.used`,
		token.ID, token.AccountID, token.Token,
		token.ExpiresAt.Format(time.RFC3339),
		used,
		token.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// GetActivationTokenByToken retrieves an activation token by its token string.
// PRE: token is non-empty
// POST: Returns the token or error if not found
func (s *SQLiteStore) GetActivationTokenByToken(ctx context.Context, token string) (domain.ActivationToken, error) {
	var t domain.ActivationToken
	var expiresStr, createdStr string
	var used int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, account_id, token, expires_at, used, created_at FROM activation_token WHERE token = ?`, token).
		Scan(&t.ID, &t.AccountID, &t.Token, &expiresStr, &used, &createdStr)
	if err != nil {
		return domain.ActivationToken{}, err
	}
	t.ExpiresAt, _ = parseTime(expiresStr)
	t.CreatedAt, _ = parseTime(createdStr)
	t.Used = used != 0
	return t, nil
}

// InvalidateTokensForAccount marks all tokens for an account as used.
// PRE: accountID is non-empty
// POST: All tokens for the account have used=1
func (s *SQLiteStore) InvalidateTokensForAccount(ctx context.Context, accountID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE activation_token SET used = 1 WHERE account_id = ? AND used = 0`, accountID)
	return err
}

func parseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %s", s)
}
