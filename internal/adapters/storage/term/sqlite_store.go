package term

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	domain "workshop/internal/domain/term"
)

const dateFormat = "2006-01-02"

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new TermStore.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a Term by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Term, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, name, start_date, end_date FROM term WHERE id = ?", id)
	var entity domain.Term
	var startStr, endStr string
	err := row.Scan(&entity.ID, &entity.Name, &startStr, &endStr)
	if err == sql.ErrNoRows {
		return domain.Term{}, fmt.Errorf("term not found: %w", err)
	}
	if err != nil {
		return domain.Term{}, err
	}
	entity.StartDate, _ = time.Parse(dateFormat, startStr)
	entity.EndDate, _ = time.Parse(dateFormat, endStr)
	return entity, nil
}

// Save persists a Term to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, entity domain.Term) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO term (id, name, start_date, end_date) VALUES (?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET name=excluded.name, start_date=excluded.start_date, end_date=excluded.end_date",
		entity.ID, entity.Name, entity.StartDate.Format(dateFormat), entity.EndDate.Format(dateFormat),
	)
	return err
}

// Delete removes a Term from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM term WHERE id = ?", id)
	return err
}

// List retrieves all Terms ordered by start date.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (s *SQLiteStore) List(ctx context.Context) ([]domain.Term, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, start_date, end_date FROM term ORDER BY start_date")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Term
	for rows.Next() {
		var entity domain.Term
		var startStr, endStr string
		if err := rows.Scan(&entity.ID, &entity.Name, &startStr, &endStr); err != nil {
			return nil, err
		}
		entity.StartDate, _ = time.Parse(dateFormat, startStr)
		entity.EndDate, _ = time.Parse(dateFormat, endStr)
		results = append(results, entity)
	}
	return results, nil
}
