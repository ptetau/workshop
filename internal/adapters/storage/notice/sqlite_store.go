package notice

import (
	"context"
	"database/sql"
	"time"

	domain "workshop/internal/domain/notice"
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

// GetByID retrieves a notice by ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Notice, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, type, status, title, content, created_by, published_by, target_id, created_at, published_at
		 FROM notice WHERE id = ?`, id)
	return scanNotice(row)
}

// Save inserts or updates a notice.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, n domain.Notice) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO notice (id, type, status, title, content, created_by, published_by, target_id, created_at, published_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   type=excluded.type, status=excluded.status, title=excluded.title, content=excluded.content,
		   created_by=excluded.created_by, published_by=excluded.published_by, target_id=excluded.target_id,
		   created_at=excluded.created_at, published_at=excluded.published_at`,
		n.ID, n.Type, n.Status, n.Title, n.Content, n.CreatedBy,
		nullableString(n.PublishedBy), nullableString(n.TargetID),
		n.CreatedAt.Format(timeLayout), nullableTime(n.PublishedAt))
	return err
}

// Delete removes a notice by ID.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM notice WHERE id = ?`, id)
	return err
}

// List returns notices matching the filter.
// PRE: filter has valid parameters
// POST: Returns matching notices
func (s *SQLiteStore) List(ctx context.Context, filter ListFilter) ([]domain.Notice, error) {
	query := `SELECT id, type, status, title, content, created_by, published_by, target_id, created_at, published_at FROM notice WHERE 1=1`
	args := []any{}

	if filter.Type != "" {
		query += ` AND type = ?`
		args = append(args, filter.Type)
	}
	if filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, filter.Status)
	}
	query += ` ORDER BY created_at DESC`
	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNotices(rows)
}

// ListPublished returns all published notices of a given type.
// PRE: noticeType is valid
// POST: Returns published notices of the given type
func (s *SQLiteStore) ListPublished(ctx context.Context, noticeType string) ([]domain.Notice, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, type, status, title, content, created_by, published_by, target_id, created_at, published_at
		 FROM notice WHERE status = 'published' AND type = ? ORDER BY published_at DESC`, noticeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotices(rows)
}

// scanNotice scans a single row into a Notice.
func scanNotice(row *sql.Row) (domain.Notice, error) {
	var n domain.Notice
	var publishedBy, targetID, publishedAt sql.NullString
	var createdAt string

	err := row.Scan(&n.ID, &n.Type, &n.Status, &n.Title, &n.Content, &n.CreatedBy,
		&publishedBy, &targetID, &createdAt, &publishedAt)
	if err != nil {
		return domain.Notice{}, err
	}

	n.CreatedAt, _ = time.Parse(timeLayout, createdAt)
	if publishedBy.Valid {
		n.PublishedBy = publishedBy.String
	}
	if targetID.Valid {
		n.TargetID = targetID.String
	}
	if publishedAt.Valid {
		n.PublishedAt, _ = time.Parse(timeLayout, publishedAt.String)
	}
	return n, nil
}

// scanNotices scans multiple rows into a slice of Notices.
func scanNotices(rows *sql.Rows) ([]domain.Notice, error) {
	var notices []domain.Notice
	for rows.Next() {
		var n domain.Notice
		var publishedBy, targetID, publishedAt sql.NullString
		var createdAt string

		err := rows.Scan(&n.ID, &n.Type, &n.Status, &n.Title, &n.Content, &n.CreatedBy,
			&publishedBy, &targetID, &createdAt, &publishedAt)
		if err != nil {
			return nil, err
		}

		n.CreatedAt, _ = time.Parse(timeLayout, createdAt)
		if publishedBy.Valid {
			n.PublishedBy = publishedBy.String
		}
		if targetID.Valid {
			n.TargetID = targetID.String
		}
		if publishedAt.Valid {
			n.PublishedAt, _ = time.Parse(timeLayout, publishedAt.String)
		}
		notices = append(notices, n)
	}
	return notices, rows.Err()
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.Format(timeLayout)
}
