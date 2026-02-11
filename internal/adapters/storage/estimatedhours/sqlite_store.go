package estimatedhours

import (
	"context"
	"database/sql"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/estimatedhours"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new SQLiteStore.
// PRE: db is a valid database connection
// POST: returns a new SQLiteStore instance
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

const timeFormat = time.RFC3339

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(timeFormat)
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(timeFormat, s)
	return t
}

// Save inserts or updates an estimated hours entry.
// PRE: e is a valid EstimatedHours
// POST: entry is persisted
func (s *SQLiteStore) Save(ctx context.Context, e domain.EstimatedHours) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO estimated_hours (id, member_id, start_date, end_date, weekly_hours, total_hours, source, status, note, created_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   member_id=excluded.member_id, start_date=excluded.start_date, end_date=excluded.end_date,
		   weekly_hours=excluded.weekly_hours, total_hours=excluded.total_hours, source=excluded.source,
		   status=excluded.status, note=excluded.note, created_by=excluded.created_by, created_at=excluded.created_at`,
		e.ID, e.MemberID, e.StartDate, e.EndDate, e.WeeklyHours, e.TotalHours,
		e.Source, e.Status, e.Note, e.CreatedBy, formatTime(e.CreatedAt))
	return err
}

// GetByID retrieves an estimated hours entry by ID.
// PRE: id is non-empty
// POST: returns the entry or error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.EstimatedHours, error) {
	var e domain.EstimatedHours
	var createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, start_date, end_date, weekly_hours, total_hours, source, status, note, created_by, created_at
		 FROM estimated_hours WHERE id = ?`, id).
		Scan(&e.ID, &e.MemberID, &e.StartDate, &e.EndDate, &e.WeeklyHours, &e.TotalHours,
			&e.Source, &e.Status, &e.Note, &e.CreatedBy, &createdAt)
	if err != nil {
		return domain.EstimatedHours{}, err
	}
	e.CreatedAt = parseTime(createdAt)
	return e, nil
}

// ListByMemberID returns all estimated hours entries for a member, ordered by start date desc.
// PRE: memberID is non-empty
// POST: returns entries or empty slice
func (s *SQLiteStore) ListByMemberID(ctx context.Context, memberID string) ([]domain.EstimatedHours, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, start_date, end_date, weekly_hours, total_hours, source, status, note, created_by, created_at
		 FROM estimated_hours WHERE member_id = ? ORDER BY start_date DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.EstimatedHours
	for rows.Next() {
		var e domain.EstimatedHours
		var createdAt string
		if err := rows.Scan(&e.ID, &e.MemberID, &e.StartDate, &e.EndDate, &e.WeeklyHours, &e.TotalHours,
			&e.Source, &e.Status, &e.Note, &e.CreatedBy, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt = parseTime(createdAt)
		result = append(result, e)
	}
	return result, rows.Err()
}

// Delete removes an estimated hours entry by ID.
// PRE: id is non-empty
// POST: entry is deleted
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM estimated_hours WHERE id = ?`, id)
	return err
}

// SumApprovedByMemberID returns the total approved estimated hours for a member.
// PRE: memberID is non-empty
// POST: returns the sum of total_hours where status = 'approved'
func (s *SQLiteStore) SumApprovedByMemberID(ctx context.Context, memberID string) (float64, error) {
	var total sql.NullFloat64
	err := s.db.QueryRowContext(ctx,
		`SELECT SUM(total_hours) FROM estimated_hours WHERE member_id = ? AND status = 'approved'`,
		memberID).Scan(&total)
	if err != nil {
		return 0, err
	}
	if !total.Valid {
		return 0, nil
	}
	return total.Float64, nil
}

// Verify interface compliance at compile time.
var _ Store = (*SQLiteStore)(nil)
