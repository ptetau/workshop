package injury

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/injury"
)

// SQLiteStore implements domain.InjuryStore using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new InjuryStore.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a Injury by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Injury, error) {
	query := "SELECT id, body_part, description, member_id, reported_at FROM injury WHERE id = ?"

	row := s.db.QueryRowContext(ctx, query, id)

	var entity domain.Injury
	var reportedAtStr string
	err := row.Scan(
		&entity.ID,
		&entity.BodyPart,
		&entity.Description,
		&entity.MemberID,
		&reportedAtStr,
	)
	if err == nil {
		entity.ReportedAt, err = parseStoredTime(reportedAtStr)
		if err != nil {
			return domain.Injury{}, fmt.Errorf("failed to parse reported_at: %w", err)
		}
	}
	if err == sql.ErrNoRows {
		return domain.Injury{}, fmt.Errorf("injury not found: %w", err)
	}
	return entity, err
}

// Save persists a Injury to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, entity domain.Injury) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Upsert implementation
	fields := []string{"id", "body_part", "description", "member_id", "reported_at"}
	placeholders := []string{"?", "?", "?", "?", "?"}
	updates := []string{"id=excluded.id", "body_part=excluded.body_part", "description=excluded.description", "member_id=excluded.member_id", "reported_at=excluded.reported_at"}

	query := fmt.Sprintf(
		"INSERT INTO injury (%s) VALUES (%s) ON CONFLICT(id) DO UPDATE SET %s",
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updates, ", "),
	)

	_, err = tx.ExecContext(ctx, query,
		entity.ID,
		entity.BodyPart,
		entity.Description,
		entity.MemberID,
		entity.ReportedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete removes a Injury from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM injury WHERE id = ?", id)
	return err
}

// List retrieves a list of Injurys based on the filter.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (s *SQLiteStore) List(ctx context.Context, filter ListFilter) ([]domain.Injury, error) {
	query := "SELECT id, body_part, description, member_id, reported_at FROM injury LIMIT ? OFFSET ?"

	rows, err := s.db.QueryContext(ctx, query, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Injury
	for rows.Next() {
		var entity domain.Injury
		var reportedAtStr string
		if err := rows.Scan(
			&entity.ID,
			&entity.BodyPart,
			&entity.Description,
			&entity.MemberID,
			&reportedAtStr,
		); err != nil {
			return nil, err
		}
		entity.ReportedAt, err = parseStoredTime(reportedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse reported_at: %w", err)
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
