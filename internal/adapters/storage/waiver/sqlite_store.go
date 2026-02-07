package waiver

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	domain "workshop/internal/domain/waiver"
)

// SQLiteStore implements domain.WaiverStore using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new WaiverStore.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a Waiver by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Waiver, error) {
	query := "SELECT id, accepted_terms, ip_address, member_id, signed_at FROM waiver WHERE id = ?"

	row := s.db.QueryRowContext(ctx, query, id)

	var entity domain.Waiver
	var signedAtStr string
	err := row.Scan(
		&entity.ID,
		&entity.AcceptedTerms,
		&entity.IPAddress,
		&entity.MemberID,
		&signedAtStr,
	)
	if err == nil {
		entity.SignedAt, err = parseStoredTime(signedAtStr)
		if err != nil {
			return domain.Waiver{}, fmt.Errorf("failed to parse signed_at: %w", err)
		}
	}
	if err == sql.ErrNoRows {
		return domain.Waiver{}, fmt.Errorf("waiver not found: %w", err)
	}
	return entity, err
}

// Save persists a Waiver to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, entity domain.Waiver) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Upsert implementation
	fields := []string{ "id","accepted_terms","ip_address","member_id","signed_at", }
	placeholders := []string{ "?","?","?","?","?", }
	updates := []string{ "id=excluded.id","accepted_terms=excluded.accepted_terms","ip_address=excluded.ip_address","member_id=excluded.member_id","signed_at=excluded.signed_at", }

	query := fmt.Sprintf(
		"INSERT INTO waiver (%s) VALUES (%s) ON CONFLICT(id) DO UPDATE SET %s",
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updates, ", "),
	)

	_, err = tx.ExecContext(ctx, query,
		entity.ID,
		entity.AcceptedTerms,
		entity.IPAddress,
		entity.MemberID,
		entity.SignedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete removes a Waiver from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM waiver WHERE id = ?", id)
	return err
}

// List retrieves a list of Waivers based on the filter.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (s *SQLiteStore) List(ctx context.Context, filter ListFilter) ([]domain.Waiver, error) {
	query := "SELECT id, accepted_terms, ip_address, member_id, signed_at FROM waiver LIMIT ? OFFSET ?"

	rows, err := s.db.QueryContext(ctx, query, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Waiver
	for rows.Next() {
		var entity domain.Waiver
		var signedAtStr string
		if err := rows.Scan(
			&entity.ID,
			&entity.AcceptedTerms,
			&entity.IPAddress,
			&entity.MemberID,
			&signedAtStr,
		); err != nil {
			return nil, err
		}
		entity.SignedAt, err = parseStoredTime(signedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse signed_at: %w", err)
		}
		results = append(results, entity)
	}
	return results, nil
}

func parseStoredTime(value string) (time.Time, error) {
	if idx := strings.Index(value, " m="); idx != -1 {
		value = value[:idx]
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.9999999-07:00",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %q", value)
}
