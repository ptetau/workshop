package clip

import (
	"context"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/clip"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new SQLiteStore and ensures the table exists.
// PRE: db is a valid, open database connection
// POST: clips table exists; store is ready for use
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS clips (
		id TEXT PRIMARY KEY,
		theme_id TEXT NOT NULL,
		title TEXT NOT NULL,
		youtube_url TEXT NOT NULL,
		youtube_id TEXT NOT NULL DEFAULT '',
		start_seconds INTEGER NOT NULL DEFAULT 0,
		end_seconds INTEGER NOT NULL DEFAULT 0,
		notes TEXT NOT NULL DEFAULT '',
		created_by TEXT NOT NULL DEFAULT '',
		promoted INTEGER NOT NULL DEFAULT 0,
		promoted_by TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (theme_id) REFERENCES themes(id)
	)`)
	return &SQLiteStore{db: db}
}

// GetByID retrieves a clip by its ID.
// PRE: id is non-empty
// POST: returns the clip or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Clip, error) {
	var c domain.Clip
	var promoted int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, theme_id, title, youtube_url, youtube_id, start_seconds, end_seconds, notes, created_by, promoted, promoted_by, created_at FROM clips WHERE id = ?`, id,
	).Scan(&c.ID, &c.ThemeID, &c.Title, &c.YouTubeURL, &c.YouTubeID, &c.StartSeconds, &c.EndSeconds, &c.Notes, &c.CreatedBy, &promoted, &c.PromotedBy, &c.CreatedAt)
	c.Promoted = promoted == 1
	return c, err
}

// Save inserts or updates a clip.
// PRE: value has a non-empty ID
// POST: clip is persisted
func (s *SQLiteStore) Save(ctx context.Context, value domain.Clip) error {
	promoted := 0
	if value.Promoted {
		promoted = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO clips (id, theme_id, title, youtube_url, youtube_id, start_seconds, end_seconds, notes, created_by, promoted, promoted_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET theme_id=excluded.theme_id, title=excluded.title, youtube_url=excluded.youtube_url,
		 youtube_id=excluded.youtube_id, start_seconds=excluded.start_seconds, end_seconds=excluded.end_seconds,
		 notes=excluded.notes, promoted=excluded.promoted, promoted_by=excluded.promoted_by`,
		value.ID, value.ThemeID, value.Title, value.YouTubeURL, value.YouTubeID, value.StartSeconds, value.EndSeconds, value.Notes, value.CreatedBy, promoted, value.PromotedBy, value.CreatedAt,
	)
	return err
}

// Delete removes a clip by ID.
// PRE: id is non-empty
// POST: clip is removed from storage
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM clips WHERE id = ?`, id)
	return err
}

// ListByThemeID returns all clips for a theme ordered by creation time.
// PRE: themeID is non-empty
// POST: returns matching clips or empty slice
func (s *SQLiteStore) ListByThemeID(ctx context.Context, themeID string) ([]domain.Clip, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, theme_id, title, youtube_url, youtube_id, start_seconds, end_seconds, notes, created_by, promoted, promoted_by, created_at FROM clips WHERE theme_id = ? ORDER BY created_at DESC`, themeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Clip
	for rows.Next() {
		var c domain.Clip
		var promoted int
		if err := rows.Scan(&c.ID, &c.ThemeID, &c.Title, &c.YouTubeURL, &c.YouTubeID, &c.StartSeconds, &c.EndSeconds, &c.Notes, &c.CreatedBy, &promoted, &c.PromotedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Promoted = promoted == 1
		list = append(list, c)
	}
	return list, rows.Err()
}

// ListPromoted returns all promoted clips ordered by creation time.
// PRE: none
// POST: returns promoted clips or empty slice
func (s *SQLiteStore) ListPromoted(ctx context.Context) ([]domain.Clip, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, theme_id, title, youtube_url, youtube_id, start_seconds, end_seconds, notes, created_by, promoted, promoted_by, created_at FROM clips WHERE promoted = 1 ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Clip
	for rows.Next() {
		var c domain.Clip
		var promoted int
		if err := rows.Scan(&c.ID, &c.ThemeID, &c.Title, &c.YouTubeURL, &c.YouTubeID, &c.StartSeconds, &c.EndSeconds, &c.Notes, &c.CreatedBy, &promoted, &c.PromotedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Promoted = promoted == 1
		list = append(list, c)
	}
	return list, rows.Err()
}
