package observation

import (
	"context"
	"database/sql"
	"time"

	domain "workshop/internal/domain/observation"
)

const timeLayout = "2006-01-02T15:04:05Z07:00"

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLiteStore.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves an Observation by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Observation, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, author_id, content, created_at, updated_at
		 FROM coach_observation WHERE id = ?`, id)
	return scanObservation(row)
}

// Save persists an Observation to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, o domain.Observation) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO coach_observation (id, member_id, author_id, content, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   member_id=excluded.member_id, author_id=excluded.author_id,
		   content=excluded.content, created_at=excluded.created_at, updated_at=excluded.updated_at`,
		o.ID, o.MemberID, o.AuthorID, o.Content,
		o.CreatedAt.Format(timeLayout), nullTime(o.UpdatedAt))
	return err
}

// Delete removes an Observation from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM coach_observation WHERE id = ?`, id)
	return err
}

// ListByMemberID retrieves Observations for a member.
// PRE: memberID is non-empty
// POST: Returns observations for the given member
func (s *SQLiteStore) ListByMemberID(ctx context.Context, memberID string) ([]domain.Observation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, author_id, content, created_at, updated_at
		 FROM coach_observation WHERE member_id = ? ORDER BY created_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var observations []domain.Observation
	for rows.Next() {
		var o domain.Observation
		var createdAt string
		var updatedAt sql.NullString
		err := rows.Scan(&o.ID, &o.MemberID, &o.AuthorID, &o.Content, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		o.CreatedAt, _ = time.Parse(timeLayout, createdAt)
		if updatedAt.Valid {
			o.UpdatedAt, _ = time.Parse(timeLayout, updatedAt.String)
		}
		observations = append(observations, o)
	}
	return observations, rows.Err()
}

func scanObservation(row *sql.Row) (domain.Observation, error) {
	var o domain.Observation
	var createdAt string
	var updatedAt sql.NullString
	err := row.Scan(&o.ID, &o.MemberID, &o.AuthorID, &o.Content, &createdAt, &updatedAt)
	if err != nil {
		return domain.Observation{}, err
	}
	o.CreatedAt, _ = time.Parse(timeLayout, createdAt)
	if updatedAt.Valid {
		o.UpdatedAt, _ = time.Parse(timeLayout, updatedAt.String)
	}
	return o, nil
}

func nullTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t.Format(timeLayout)
}
