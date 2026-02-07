package classtype

import (
	"context"
	"database/sql"
	"fmt"

	domain "workshop/internal/domain/classtype"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new ClassTypeStore.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a ClassType by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.ClassType, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, program_id, name FROM class_type WHERE id = ?", id)
	var entity domain.ClassType
	err := row.Scan(&entity.ID, &entity.ProgramID, &entity.Name)
	if err == sql.ErrNoRows {
		return domain.ClassType{}, fmt.Errorf("class type not found: %w", err)
	}
	return entity, err
}

// Save persists a ClassType to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, entity domain.ClassType) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO class_type (id, program_id, name) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET program_id=excluded.program_id, name=excluded.name",
		entity.ID, entity.ProgramID, entity.Name,
	)
	return err
}

// Delete removes a ClassType from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM class_type WHERE id = ?", id)
	return err
}

// List retrieves all ClassTypes.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (s *SQLiteStore) List(ctx context.Context) ([]domain.ClassType, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, program_id, name FROM class_type ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.ClassType
	for rows.Next() {
		var entity domain.ClassType
		if err := rows.Scan(&entity.ID, &entity.ProgramID, &entity.Name); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return results, nil
}

// ListByProgramID retrieves ClassTypes for a specific program.
// PRE: programID is non-empty
// POST: Returns class types for the given program
func (s *SQLiteStore) ListByProgramID(ctx context.Context, programID string) ([]domain.ClassType, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, program_id, name FROM class_type WHERE program_id = ? ORDER BY name", programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.ClassType
	for rows.Next() {
		var entity domain.ClassType
		if err := rows.Scan(&entity.ID, &entity.ProgramID, &entity.Name); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return results, nil
}
