package clip

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/clip"
)

// SQLiteComparisonStore implements ComparisonStore using SQLite.
type SQLiteComparisonStore struct {
	db storage.SQLDB
}

// NewSQLiteComparisonStore creates a new SQLiteComparisonStore.
func NewSQLiteComparisonStore(db storage.SQLDB) *SQLiteComparisonStore {
	db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS comparison_sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL DEFAULT '',
		clip_ids TEXT NOT NULL DEFAULT '[]',
		created_by TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS research_notes (
		id TEXT PRIMARY KEY,
		comparison_session_id TEXT NOT NULL UNIQUE,
		content TEXT NOT NULL DEFAULT '',
		created_by TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (comparison_session_id) REFERENCES comparison_sessions(id) ON DELETE CASCADE
	)`)
	db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS idx_comparison_created_by ON comparison_sessions(created_by)`)
	return &SQLiteComparisonStore{db: db}
}

// SaveSession inserts or updates a comparison session.
// PRE: session has valid ID
// POST: session persisted; error if database operation fails
func (s *SQLiteComparisonStore) SaveSession(ctx context.Context, session domain.ComparisonSession) error {
	clipIDsJSON, err := json.Marshal(session.ClipIDs)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO comparison_sessions (id, name, clip_ids, created_by, created_at) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET name=excluded.name, clip_ids=excluded.clip_ids`,
		session.ID, session.Name, string(clipIDsJSON), session.CreatedBy, session.CreatedAt,
	)
	return err
}

// GetSessionByID retrieves a comparison session by ID.
// PRE: id is non-empty
// POST: returns session if found; error if not found or database fails
func (s *SQLiteComparisonStore) GetSessionByID(ctx context.Context, id string) (domain.ComparisonSession, error) {
	var session domain.ComparisonSession
	var clipIDsJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, clip_ids, created_by, created_at FROM comparison_sessions WHERE id = ?`, id,
	).Scan(&session.ID, &session.Name, &clipIDsJSON, &session.CreatedBy, &session.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ComparisonSession{}, errors.New("comparison session not found")
	}
	if err != nil {
		return domain.ComparisonSession{}, err
	}
	if err := json.Unmarshal([]byte(clipIDsJSON), &session.ClipIDs); err != nil {
		return domain.ComparisonSession{}, err
	}
	return session, nil
}

// ListSessionsByUser retrieves all comparison sessions created by a user.
// PRE: userID is non-empty
// POST: returns sessions ordered by created_at desc; error if database fails
func (s *SQLiteComparisonStore) ListSessionsByUser(ctx context.Context, userID string) ([]domain.ComparisonSession, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, clip_ids, created_by, created_at FROM comparison_sessions WHERE created_by = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.ComparisonSession
	for rows.Next() {
		var session domain.ComparisonSession
		var clipIDsJSON string
		if err := rows.Scan(&session.ID, &session.Name, &clipIDsJSON, &session.CreatedBy, &session.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(clipIDsJSON), &session.ClipIDs); err != nil {
			return nil, err
		}
		list = append(list, session)
	}
	return list, rows.Err()
}

// DeleteSession removes a comparison session.
// PRE: id is non-empty
// POST: session removed; error if database operation fails
func (s *SQLiteComparisonStore) DeleteSession(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM comparison_sessions WHERE id = ?`, id)
	return err
}

// SaveResearchNote inserts or updates a research note.
// PRE: note has valid ID and ComparisonSessionID
// POST: note persisted; error if database operation fails
func (s *SQLiteComparisonStore) SaveResearchNote(ctx context.Context, note domain.ResearchNote) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO research_notes (id, comparison_session_id, content, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(comparison_session_id) DO UPDATE SET content=excluded.content, updated_at=excluded.updated_at`,
		note.ID, note.ComparisonSessionID, note.Content, note.CreatedBy, note.CreatedAt, note.UpdatedAt,
	)
	return err
}

// GetResearchNoteBySession retrieves a research note by comparison session ID.
// PRE: sessionID is non-empty
// POST: returns note if found; error if not found or database fails
func (s *SQLiteComparisonStore) GetResearchNoteBySession(ctx context.Context, sessionID string) (domain.ResearchNote, error) {
	var note domain.ResearchNote
	err := s.db.QueryRowContext(ctx,
		`SELECT id, comparison_session_id, content, created_by, created_at, updated_at FROM research_notes WHERE comparison_session_id = ?`, sessionID,
	).Scan(&note.ID, &note.ComparisonSessionID, &note.Content, &note.CreatedBy, &note.CreatedAt, &note.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ResearchNote{}, errors.New("research note not found")
	}
	return note, err
}

// DeleteResearchNote removes a research note.
// PRE: id is non-empty
// POST: note removed; error if database operation fails
func (s *SQLiteComparisonStore) DeleteResearchNote(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM research_notes WHERE id = ?`, id)
	return err
}
