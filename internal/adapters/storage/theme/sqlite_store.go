package theme

import (
	"context"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/theme"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new SQLiteStore and ensures the table exists.
// PRE: db is a valid, open database connection
// POST: themes table exists; store is ready for use
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS themes (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		program TEXT NOT NULL,
		start_date DATETIME NOT NULL,
		end_date DATETIME NOT NULL,
		created_by TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	return &SQLiteStore{db: db}
}

// GetByID retrieves a theme by its ID.
// PRE: id is non-empty
// POST: returns the theme or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Theme, error) {
	var t domain.Theme
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, program, start_date, end_date, created_by, created_at FROM themes WHERE id = ?`, id,
	).Scan(&t.ID, &t.Name, &t.Description, &t.Program, &t.StartDate, &t.EndDate, &t.CreatedBy, &t.CreatedAt)
	return t, err
}

// Save inserts or updates a theme.
// PRE: value has a non-empty ID
// POST: theme is persisted
func (s *SQLiteStore) Save(ctx context.Context, value domain.Theme) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO themes (id, name, description, program, start_date, end_date, created_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET name=excluded.name, description=excluded.description, program=excluded.program,
		 start_date=excluded.start_date, end_date=excluded.end_date, created_by=excluded.created_by`,
		value.ID, value.Name, value.Description, value.Program, value.StartDate, value.EndDate, value.CreatedBy, value.CreatedAt,
	)
	return err
}

// Delete removes a theme by ID.
// PRE: id is non-empty
// POST: theme is removed from storage
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM themes WHERE id = ?`, id)
	return err
}

// List returns all themes ordered by start date descending.
// PRE: none
// POST: returns all themes or empty slice
func (s *SQLiteStore) List(ctx context.Context) ([]domain.Theme, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description, program, start_date, end_date, created_by, created_at FROM themes ORDER BY start_date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Theme
	for rows.Next() {
		var t domain.Theme
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Program, &t.StartDate, &t.EndDate, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

// ListByProgram returns themes for a specific program ordered by start date descending.
// PRE: program is non-empty
// POST: returns matching themes or empty slice
func (s *SQLiteStore) ListByProgram(ctx context.Context, program string) ([]domain.Theme, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description, program, start_date, end_date, created_by, created_at FROM themes WHERE program = ? ORDER BY start_date DESC`, program)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Theme
	for rows.Next() {
		var t domain.Theme
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Program, &t.StartDate, &t.EndDate, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}
