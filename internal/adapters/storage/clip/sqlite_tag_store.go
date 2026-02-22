package clip

import (
	"context"
	"database/sql"
	"errors"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/clip"
)

// SQLiteTagStore implements TagStore using SQLite.
type SQLiteTagStore struct {
	db storage.SQLDB
}

// NewSQLiteTagStore creates a new SQLiteTagStore and ensures tables exist.
func NewSQLiteTagStore(db storage.SQLDB) *SQLiteTagStore {
	db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS clip_tags (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		created_by TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS clip_tag_associations (
		clip_id TEXT NOT NULL,
		tag_id TEXT NOT NULL,
		created_by TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (clip_id, tag_id),
		FOREIGN KEY (clip_id) REFERENCES clips(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES clip_tags(id) ON DELETE CASCADE
	)`)
	db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS idx_clip_tag_clip ON clip_tag_associations(clip_id)`)
	db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS idx_clip_tag_tag ON clip_tag_associations(tag_id)`)
	return &SQLiteTagStore{db: db}
}

// SaveTag inserts or updates a tag.
// PRE: tag has valid ID and Name
// POST: tag persisted to database; error if database operation fails
func (s *SQLiteTagStore) SaveTag(ctx context.Context, tag domain.Tag) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO clip_tags (id, name, created_by, created_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET name=excluded.name`,
		tag.ID, tag.Name, tag.CreatedBy, tag.CreatedAt,
	)
	return err
}

// GetTagByID retrieves a tag by ID.
// PRE: id is non-empty
// POST: returns tag if found; error if not found or database fails
func (s *SQLiteTagStore) GetTagByID(ctx context.Context, id string) (domain.Tag, error) {
	var t domain.Tag
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, created_by, created_at FROM clip_tags WHERE id = ?`, id,
	).Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Tag{}, errors.New("tag not found")
	}
	return t, err
}

// GetTagByName retrieves a tag by name (case-insensitive).
// PRE: name is non-empty
// POST: returns tag if found; error if not found or database fails
func (s *SQLiteTagStore) GetTagByName(ctx context.Context, name string) (domain.Tag, error) {
	var t domain.Tag
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, created_by, created_at FROM clip_tags WHERE LOWER(name) = LOWER(?)`, name,
	).Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Tag{}, errors.New("tag not found")
	}
	return t, err
}

// ListTags returns all tags ordered by name.
// PRE: none
// POST: returns all tags; error if database fails
func (s *SQLiteTagStore) ListTags(ctx context.Context) ([]domain.Tag, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, created_by, created_at FROM clip_tags ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

// DeleteTag removes a tag and all its associations.
// PRE: id is non-empty
// POST: tag and associations removed; error if database fails
func (s *SQLiteTagStore) DeleteTag(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM clip_tags WHERE id = ?`, id)
	return err
}

// AddTagToClip associates a tag with a clip.
// PRE: clipTag has valid ClipID and TagID
// POST: association persisted; error if database operation fails
func (s *SQLiteTagStore) AddTagToClip(ctx context.Context, clipTag domain.ClipTag) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO clip_tag_associations (clip_id, tag_id, created_by, created_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(clip_id, tag_id) DO NOTHING`,
		clipTag.ClipID, clipTag.TagID, clipTag.CreatedBy, clipTag.CreatedAt,
	)
	return err
}

// RemoveTagFromClip removes a tag association from a clip.
// PRE: clipID and tagID are non-empty
// POST: association removed; error if database operation fails
func (s *SQLiteTagStore) RemoveTagFromClip(ctx context.Context, clipID, tagID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM clip_tag_associations WHERE clip_id = ? AND tag_id = ?`,
		clipID, tagID,
	)
	return err
}

// GetTagsForClip returns all tags associated with a clip.
// PRE: clipID is non-empty
// POST: returns tags for clip; error if database fails
func (s *SQLiteTagStore) GetTagsForClip(ctx context.Context, clipID string) ([]domain.Tag, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT t.id, t.name, t.created_by, t.created_at 
		 FROM clip_tags t
		 JOIN clip_tag_associations a ON t.id = a.tag_id
		 WHERE a.clip_id = ?
		 ORDER BY t.name ASC`, clipID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

// GetClipsForTag returns all clip IDs associated with a tag.
// PRE: tagID is non-empty
// POST: returns clip IDs; error if database fails
func (s *SQLiteTagStore) GetClipsForTag(ctx context.Context, tagID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT clip_id FROM clip_tag_associations WHERE tag_id = ?`, tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []string
	for rows.Next() {
		var clipID string
		if err := rows.Scan(&clipID); err != nil {
			return nil, err
		}
		list = append(list, clipID)
	}
	return list, rows.Err()
}

// SearchClipsByTags returns clips that have ALL of the specified tags (AND search).
// PRE: none (empty tagIDs returns nil, nil)
// POST: returns clips matching all tags; error if database fails
func (s *SQLiteTagStore) SearchClipsByTags(ctx context.Context, tagIDs []string) ([]domain.Clip, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}
	// Build query for AND search: clip must have all specified tags
	query := `SELECT c.id, c.theme_id, c.title, c.youtube_url, c.youtube_id, c.start_seconds, c.end_seconds, c.notes, c.created_by, c.promoted, c.promoted_by, c.created_at
			  FROM clips c
			  JOIN clip_tag_associations a ON c.id = a.clip_id
			  WHERE a.tag_id IN (`
	params := make([]interface{}, len(tagIDs))
	for i, id := range tagIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		params[i] = id
	}
	query += `)
			  GROUP BY c.id
			  HAVING COUNT(DISTINCT a.tag_id) = ?
			  ORDER BY c.created_at DESC`
	params = append(params, len(tagIDs))

	rows, err := s.db.QueryContext(ctx, query, params...)
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
