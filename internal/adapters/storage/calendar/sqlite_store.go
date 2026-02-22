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
	// Store end_date as start_date for single-day events so range queries can use
	// a simple overlap predicate: start_date <= to AND end_date >= from.
	endDate := e.StartDate.Format("2006-01-02")
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
	// Preserve domain semantic: EndDate zero => single-day.
	if endStr != "" && endStr != startStr {
		e.EndDate = parseDate(endStr)
	}
	return e, nil
}

// ListByDateRange returns events overlapping the given date range.
// PRE: from and to are valid date strings (YYYY-MM-DD)
// POST: returns events sorted by start_date ascending
func (s *SQLiteStore) ListByDateRange(ctx context.Context, from, to string) ([]domain.Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, type, description, location, start_date, end_date, registration_url, created_by, created_at
		 FROM calendar_event
		 WHERE start_date <= ? AND end_date >= ?
		 ORDER BY start_date ASC`, to, from,
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
		// Preserve domain semantic: EndDate zero => single-day.
		if endStr != "" && endStr != startStr {
			e.EndDate = parseDate(endStr)
		}
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

// SaveInterest inserts a competition interest record.
// PRE: ci.ID, ci.EventID, ci.MemberID are non-empty; ci.CreatedAt is valid.
// POST: The interest is persisted (idempotent on conflict).
// INVARIANT: Database connection is valid.
func (s *SQLiteStore) SaveInterest(ctx context.Context, ci domain.CompetitionInterest) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO competition_interest (id, event_id, member_id, created_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(event_id, member_id) DO NOTHING`,
		ci.ID, ci.EventID, ci.MemberID, ci.CreatedAt,
	)
	return err
}

// DeleteInterest removes a competition interest record.
// PRE: eventID and memberID are non-empty.
// POST: The interest record is removed if it existed.
// INVARIANT: Database connection is valid.
func (s *SQLiteStore) DeleteInterest(ctx context.Context, eventID, memberID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM competition_interest WHERE event_id = ? AND member_id = ?`,
		eventID, memberID,
	)
	return err
}

// GetInterestsByEvent returns all interest records for an event.
// PRE: eventID is non-empty.
// POST: Returns slice of interests (empty if none found), ordered by creation time.
// INVARIANT: Database connection is valid.
func (s *SQLiteStore) GetInterestsByEvent(ctx context.Context, eventID string) ([]domain.CompetitionInterest, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, event_id, member_id, created_at FROM competition_interest WHERE event_id = ?`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interests []domain.CompetitionInterest
	for rows.Next() {
		var ci domain.CompetitionInterest
		if err := rows.Scan(&ci.ID, &ci.EventID, &ci.MemberID, &ci.CreatedAt); err != nil {
			return nil, err
		}
		interests = append(interests, ci)
	}
	return interests, rows.Err()
}

// CountInterestsByEvent returns the number of interested members for an event.
// PRE: eventID is non-empty.
// POST: Returns count >= 0.
// INVARIANT: Database connection is valid.
func (s *SQLiteStore) CountInterestsByEvent(ctx context.Context, eventID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM competition_interest WHERE event_id = ?`,
		eventID,
	).Scan(&count)
	return count, err
}

// IsInterested checks if a member is interested in an event.
// PRE: eventID and memberID are non-empty.
// POST: Returns true if interest exists, false otherwise.
// INVARIANT: Database connection is valid.
func (s *SQLiteStore) IsInterested(ctx context.Context, eventID, memberID string) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM competition_interest WHERE event_id = ? AND member_id = ?`,
		eventID, memberID,
	).Scan(&exists)
	if err != nil {
		return false, nil
	}
	return true, nil
}
