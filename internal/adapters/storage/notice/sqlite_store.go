package notice

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/notice"
)

const timeLayout = "2006-01-02T15:04:05Z07:00"

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new SQLiteStore.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

const noticeColumns = `id, type, status, title, content, created_by, published_by, target_id,
		author_name, show_author, color, pinned, pinned_at, visible_from, visible_until,
		created_at, updated_at, published_at`

// GetByID retrieves a notice by ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Notice, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+noticeColumns+` FROM notice WHERE id = ?`, id)
	return scanNotice(row)
}

// Save inserts or updates a notice.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, n domain.Notice) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO notice (id, type, status, title, content, created_by, published_by, target_id,
		   author_name, show_author, color, pinned, pinned_at, visible_from, visible_until,
		   created_at, updated_at, published_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   type=excluded.type, status=excluded.status, title=excluded.title, content=excluded.content,
		   created_by=excluded.created_by, published_by=excluded.published_by, target_id=excluded.target_id,
		   author_name=excluded.author_name, show_author=excluded.show_author, color=excluded.color,
		   pinned=excluded.pinned, pinned_at=excluded.pinned_at, visible_from=excluded.visible_from,
		   visible_until=excluded.visible_until, created_at=excluded.created_at, updated_at=excluded.updated_at,
		   published_at=excluded.published_at`,
		n.ID, n.Type, n.Status, n.Title, n.Content, n.CreatedBy,
		nullableString(n.PublishedBy), nullableString(n.TargetID),
		n.AuthorName, boolToInt(n.ShowAuthor), n.Color, boolToInt(n.Pinned),
		nullableTime(n.PinnedAt), nullableTime(n.VisibleFrom), nullableTime(n.VisibleUntil),
		n.CreatedAt.Format(timeLayout), nullableTime(n.UpdatedAt), nullableTime(n.PublishedAt))
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
// POST: Returns matching notices ordered by pinned first (most recently pinned), then by created_at DESC
func (s *SQLiteStore) List(ctx context.Context, filter ListFilter) ([]domain.Notice, error) {
	query := `SELECT ` + noticeColumns + ` FROM notice WHERE 1=1`
	args := []any{}

	if filter.Type != "" {
		query += ` AND type = ?`
		args = append(args, filter.Type)
	}
	if filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, filter.Status)
	}
	query += ` ORDER BY pinned DESC, pinned_at DESC, created_at DESC`
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

// ListPublished returns all published notices of a given type that are currently visible.
// PRE: noticeType is valid, now is the current time
// POST: Returns published, visible notices ordered by pinned first then published_at DESC
func (s *SQLiteStore) ListPublished(ctx context.Context, noticeType string, now time.Time) ([]domain.Notice, error) {
	nowStr := now.UTC().Format(timeLayout)
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+noticeColumns+`
		 FROM notice WHERE status = 'published' AND type = ?
		 AND (visible_from IS NULL OR visible_from <= ?)
		 AND (visible_until IS NULL OR visible_until >= ?)
		 ORDER BY pinned DESC, pinned_at DESC, published_at DESC`,
		noticeType, nowStr, nowStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotices(rows)
}

// scannedRow holds the raw scanned values from a notice row before conversion.
type scannedRow struct {
	publishedBy  sql.NullString
	targetID     sql.NullString
	showAuthor   int
	pinned       int
	pinnedAt     sql.NullString
	visibleFrom  sql.NullString
	visibleUntil sql.NullString
	createdAt    string
	updatedAt    sql.NullString
	publishedAt  sql.NullString
}

// scanNotice scans a single row into a Notice.
func scanNotice(row *sql.Row) (domain.Notice, error) {
	var n domain.Notice
	var s scannedRow

	err := row.Scan(&n.ID, &n.Type, &n.Status, &n.Title, &n.Content, &n.CreatedBy,
		&s.publishedBy, &s.targetID,
		&n.AuthorName, &s.showAuthor, &n.Color, &s.pinned,
		&s.pinnedAt, &s.visibleFrom, &s.visibleUntil,
		&s.createdAt, &s.updatedAt, &s.publishedAt)
	if err != nil {
		return domain.Notice{}, err
	}

	applyScanned(&n, &s)
	return n, nil
}

// scanNotices scans multiple rows into a slice of Notices.
func scanNotices(rows *sql.Rows) ([]domain.Notice, error) {
	var notices []domain.Notice
	for rows.Next() {
		var n domain.Notice
		var s scannedRow

		err := rows.Scan(&n.ID, &n.Type, &n.Status, &n.Title, &n.Content, &n.CreatedBy,
			&s.publishedBy, &s.targetID,
			&n.AuthorName, &s.showAuthor, &n.Color, &s.pinned,
			&s.pinnedAt, &s.visibleFrom, &s.visibleUntil,
			&s.createdAt, &s.updatedAt, &s.publishedAt)
		if err != nil {
			return nil, err
		}

		applyScanned(&n, &s)
		notices = append(notices, n)
	}
	return notices, rows.Err()
}

// applyScanned converts raw scanned values into the Notice domain fields.
func applyScanned(n *domain.Notice, s *scannedRow) {
	n.ShowAuthor = s.showAuthor != 0
	n.Pinned = s.pinned != 0
	if s.publishedBy.Valid {
		n.PublishedBy = s.publishedBy.String
	}
	if s.targetID.Valid {
		n.TargetID = s.targetID.String
	}
	n.CreatedAt = parseTime(s.createdAt, "created_at", n.ID)
	n.PinnedAt = parseNullableTime(s.pinnedAt, "pinned_at", n.ID)
	n.VisibleFrom = parseNullableTime(s.visibleFrom, "visible_from", n.ID)
	n.VisibleUntil = parseNullableTime(s.visibleUntil, "visible_until", n.ID)
	n.UpdatedAt = parseNullableTime(s.updatedAt, "updated_at", n.ID)
	n.PublishedAt = parseNullableTime(s.publishedAt, "published_at", n.ID)
}

// parseTime parses a time string, logging a warning on failure.
func parseTime(raw, field, noticeID string) time.Time {
	t, err := time.Parse(timeLayout, raw)
	if err != nil {
		slog.Warn("notice: failed to parse time", "field", field, "notice_id", noticeID, "raw", raw, "error", err)
	}
	return t
}

// parseNullableTime parses a nullable time string, logging a warning on failure.
func parseNullableTime(ns sql.NullString, field, noticeID string) time.Time {
	if !ns.Valid {
		return time.Time{}
	}
	return parseTime(ns.String, field, noticeID)
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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
