package program

import (
	"context"
	"database/sql"
	"fmt"

	domain "workshop/internal/domain/program"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new ProgramStore.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a Program by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Program, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, name, type FROM program WHERE id = ?", id)
	var entity domain.Program
	err := row.Scan(&entity.ID, &entity.Name, &entity.Type)
	if err == sql.ErrNoRows {
		return domain.Program{}, fmt.Errorf("program not found: %w", err)
	}
	return entity, err
}

// Save persists a Program to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, entity domain.Program) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO program (id, name, type) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET name=excluded.name, type=excluded.type",
		entity.ID, entity.Name, entity.Type,
	)
	return err
}

// Delete removes a Program from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM program WHERE id = ?", id)
	return err
}

// List retrieves all Programs.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (s *SQLiteStore) List(ctx context.Context) ([]domain.Program, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, type FROM program ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Program
	for rows.Next() {
		var entity domain.Program
		if err := rows.Scan(&entity.ID, &entity.Name, &entity.Type); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return results, nil
}
