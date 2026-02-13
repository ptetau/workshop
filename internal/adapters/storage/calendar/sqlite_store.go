package calendar

import (
	"context"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/calendar"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new SQLiteStore.
// PRE: db is a valid, open database connection with migrations applied
// POST: store is ready for use
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// Save inserts or updates a calendar event.
// PRE: e is a valid Event (Validate() returns nil)
// POST: event is persisted
func (s *SQLiteStore) Save(ctx context.Context, e domain.Event) error {
	endDate := ""
	if !e.EndDate.IsZero() {
		endDate = e.EndDate.Format("2006-01-02")
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO calendar_event (id, title, type, description, location, start_date, end_date, registration_url, created_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   title=excluded.title, type=excluded.type, description=excluded.description,
		   location=excluded.location, start_date=excluded.start_date, end_date=excluded.end_date,
		   registration_url=excluded.registration_url`,
		e.ID, e.Title, e.Type, e.Description, e.Location,
		e.StartDate.Format("2006-01-02"), endDate,
		e.RegistrationURL, e.CreatedBy, e.CreatedAt,
	)
	return err
}

// GetByID retrieves a calendar event by ID.
// PRE: id is non-empty
// POST: returns the event or error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Event, error) {
	var e domain.Event
	var startStr, endStr string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, type, description, location, start_date, end_date, registration_url, created_by, created_at
		 FROM calendar_event WHERE id = ?`, id,
	).Scan(&e.ID, &e.Title, &e.Type, &e.Description, &e.Location,
		&startStr, &endStr, &e.RegistrationURL, &e.CreatedBy, &e.CreatedAt)
	if err != nil {
		return e, err
	}
	e.StartDate = parseDate(startStr)
	e.EndDate = parseDate(endStr)
	return e, nil
}

// ListByDateRange returns events overlapping the given date range.
// PRE: from and to are valid date strings (YYYY-MM-DD)
// POST: returns events sorted by start_date ascending
func (s *SQLiteStore) ListByDateRange(ctx context.Context, from, to string) ([]domain.Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, type, description, location, start_date, end_date, registration_url, created_by, created_at
		 FROM calendar_event
		 WHERE start_date <= ? AND (end_date >= ? OR (end_date = '' AND start_date >= ?))
		 ORDER BY start_date ASC`, to, from, from,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var e domain.Event
		var startStr, endStr string
		if err := rows.Scan(&e.ID, &e.Title, &e.Type, &e.Description, &e.Location,
			&startStr, &endStr, &e.RegistrationURL, &e.CreatedBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.StartDate = parseDate(startStr)
		e.EndDate = parseDate(endStr)
		events = append(events, e)
	}
	return events, rows.Err()
}

// Delete removes a calendar event by ID.
// PRE: id is non-empty
// POST: event is removed from storage
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM calendar_event WHERE id = ?`, id)
	return err
}

func parseDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse("2006-01-02", s)
	return t
}
