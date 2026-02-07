package member

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	domain "workshop/internal/domain/member"
)

// SQLiteStore implements domain.MemberStore using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new MemberStore.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a Member by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Member, error) {
	query := "SELECT id, account_id, email, fee, frequency, name, program, status FROM member WHERE id = ?"

	row := s.db.QueryRowContext(ctx, query, id)

	var entity domain.Member
	var accountID sql.NullString
	err := row.Scan(
		&entity.ID,
		&accountID,
		&entity.Email,
		&entity.Fee,
		&entity.Frequency,
		&entity.Name,
		&entity.Program,
		&entity.Status,
	)
	if accountID.Valid {
		entity.AccountID = accountID.String
	}
	if err == sql.ErrNoRows {
		return domain.Member{}, fmt.Errorf("member not found: %w", err)
	}
	return entity, err
}

// GetByEmail retrieves a Member by email.
// PRE: email is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByEmail(ctx context.Context, email string) (domain.Member, error) {
	query := "SELECT id, account_id, email, fee, frequency, name, program, status FROM member WHERE email = ?"

	row := s.db.QueryRowContext(ctx, query, email)

	var entity domain.Member
	var accountID sql.NullString
	err := row.Scan(
		&entity.ID,
		&accountID,
		&entity.Email,
		&entity.Fee,
		&entity.Frequency,
		&entity.Name,
		&entity.Program,
		&entity.Status,
	)
	if accountID.Valid {
		entity.AccountID = accountID.String
	}
	if err == sql.ErrNoRows {
		return domain.Member{}, fmt.Errorf("member not found: %w", err)
	}
	return entity, err
}

// Save persists a Member to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, entity domain.Member) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Upsert implementation
	fields := []string{"id", "account_id", "email", "fee", "frequency", "name", "program", "status"}
	placeholders := []string{"?", "?", "?", "?", "?", "?", "?", "?"}
	updates := []string{"account_id=excluded.account_id", "email=excluded.email", "fee=excluded.fee", "frequency=excluded.frequency", "name=excluded.name", "program=excluded.program", "status=excluded.status"}

	query := fmt.Sprintf(
		"INSERT INTO member (%s) VALUES (%s) ON CONFLICT(id) DO UPDATE SET %s",
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updates, ", "),
	)

	var accountID interface{}
	if entity.AccountID != "" {
		accountID = entity.AccountID
	}

	_, err = tx.ExecContext(ctx, query,
		entity.ID,
		accountID,
		entity.Email,
		entity.Fee,
		entity.Frequency,
		entity.Name,
		entity.Program,
		entity.Status,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete removes a Member from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM member WHERE id = ?", id)
	return err
}

// SearchByName finds members whose name matches the query (case-insensitive LIKE).
// PRE: query is non-empty, limit > 0
// POST: Returns matching members ordered by name
func (s *SQLiteStore) SearchByName(ctx context.Context, query string, limit int) ([]domain.Member, error) {
	q := "SELECT id, account_id, email, fee, frequency, name, program, status FROM member WHERE name LIKE ? AND status != 'archived' ORDER BY name LIMIT ?"
	rows, err := s.db.QueryContext(ctx, q, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Member
	for rows.Next() {
		var entity domain.Member
		var accountID sql.NullString
		if err := rows.Scan(
			&entity.ID,
			&accountID,
			&entity.Email,
			&entity.Fee,
			&entity.Frequency,
			&entity.Name,
			&entity.Program,
			&entity.Status,
		); err != nil {
			return nil, err
		}
		if accountID.Valid {
			entity.AccountID = accountID.String
		}
		results = append(results, entity)
	}
	return results, nil
}

// List retrieves a list of Members based on the filter.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (s *SQLiteStore) List(ctx context.Context, filter ListFilter) ([]domain.Member, error) {
	query := "SELECT id, account_id, email, fee, frequency, name, program, status FROM member LIMIT ? OFFSET ?"

	rows, err := s.db.QueryContext(ctx, query, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Member
	for rows.Next() {
		var entity domain.Member
		var accountID sql.NullString
		if err := rows.Scan(
			&entity.ID,
			&accountID,
			&entity.Email,
			&entity.Fee,
			&entity.Frequency,
			&entity.Name,
			&entity.Program,
			&entity.Status,
		); err != nil {
			return nil, err
		}
		if accountID.Valid {
			entity.AccountID = accountID.String
		}
		results = append(results, entity)
	}
	return results, nil
}
