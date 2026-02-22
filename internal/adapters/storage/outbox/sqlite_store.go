package outbox

import (
	"context"
	"database/sql"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/outbox"
)

const (
	dateLayout = "2006-01-02T15:04:05.999999999Z07:00"
)

// SQLiteStore implements the outbox Store interface using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new outbox store.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves an outbox entry by its ID.
// PRE: id is non-empty
// POST: Returns the entry or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Entry, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, action_type, payload, status, attempts, max_attempts, last_attempted_at, created_at, external_id, error_message
		 FROM outbox WHERE id = ?`, id)
	return scanEntry(row)
}

// Save persists an outbox entry to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, e domain.Entry) error {
	lastAttemptedAt := ""
	if !e.LastAttemptedAt.IsZero() {
		lastAttemptedAt = e.LastAttemptedAt.Format(dateLayout)
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO outbox (id, action_type, payload, status, attempts, max_attempts, last_attempted_at, created_at, external_id, error_message)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   action_type=excluded.action_type, payload=excluded.payload, status=excluded.status,
		   attempts=excluded.attempts, max_attempts=excluded.max_attempts,
		   last_attempted_at=excluded.last_attempted_at, external_id=excluded.external_id,
		   error_message=excluded.error_message`,
		e.ID, e.ActionType, e.Payload, e.Status, e.Attempts, e.MaxAttempts,
		lastAttemptedAt, e.CreatedAt.Format(dateLayout), e.ExternalID, e.ErrorMessage)
	return err
}

// ListPending returns entries that need to be processed (pending or retrying).
// PRE: limit > 0
// POST: Returns up to limit entries ordered by created_at
func (s *SQLiteStore) ListPending(ctx context.Context, limit int) ([]domain.Entry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, action_type, payload, status, attempts, max_attempts, last_attempted_at, created_at, external_id, error_message
		 FROM outbox WHERE status IN (?, ?) ORDER BY created_at ASC LIMIT ?`,
		domain.StatusPending, domain.StatusRetrying, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntries(rows)
}

// ListFailed returns entries that have permanently failed.
// PRE: limit > 0
// POST: Returns up to limit failed entries ordered by last_attempted_at desc
func (s *SQLiteStore) ListFailed(ctx context.Context, limit int) ([]domain.Entry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, action_type, payload, status, attempts, max_attempts, last_attempted_at, created_at, external_id, error_message
		 FROM outbox WHERE status = ? AND attempts >= max_attempts ORDER BY last_attempted_at DESC LIMIT ?`,
		domain.StatusFailed, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntries(rows)
}

// ListByActionType returns entries filtered by action type and status.
// PRE: actionType is non-empty
// POST: Returns matching entries ordered by created_at
func (s *SQLiteStore) ListByActionType(ctx context.Context, actionType string, status string, limit int) ([]domain.Entry, error) {
	var rows *sql.Rows
	var err error
	if status != "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, action_type, payload, status, attempts, max_attempts, last_attempted_at, created_at, external_id, error_message
			 FROM outbox WHERE action_type = ? AND status = ? ORDER BY created_at ASC LIMIT ?`,
			actionType, status, limit)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, action_type, payload, status, attempts, max_attempts, last_attempted_at, created_at, external_id, error_message
			 FROM outbox WHERE action_type = ? ORDER BY created_at ASC LIMIT ?`,
			actionType, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntries(rows)
}

// Delete removes an outbox entry (only for abandoned/terminal entries).
// PRE: id is non-empty and entry is in terminal state
// POST: Entry is removed from database
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM outbox WHERE id = ?`, id)
	return err
}

// scanEntry scans a single row into an Entry.
func scanEntry(row *sql.Row) (domain.Entry, error) {
	var e domain.Entry
	var createdAt, lastAttemptedAt string
	err := row.Scan(&e.ID, &e.ActionType, &e.Payload, &e.Status, &e.Attempts, &e.MaxAttempts,
		&lastAttemptedAt, &createdAt, &e.ExternalID, &e.ErrorMessage)
	if err != nil {
		return domain.Entry{}, err
	}
	e.CreatedAt, _ = time.Parse(dateLayout, createdAt)
	if lastAttemptedAt != "" {
		e.LastAttemptedAt, _ = time.Parse(dateLayout, lastAttemptedAt)
	}
	return e, nil
}

// scanEntryFromRows scans a single row from Rows into an Entry.
func scanEntryFromRows(rows *sql.Rows) (domain.Entry, error) {
	var e domain.Entry
	var createdAt, lastAttemptedAt string
	err := rows.Scan(&e.ID, &e.ActionType, &e.Payload, &e.Status, &e.Attempts, &e.MaxAttempts,
		&lastAttemptedAt, &createdAt, &e.ExternalID, &e.ErrorMessage)
	if err != nil {
		return domain.Entry{}, err
	}
	e.CreatedAt, _ = time.Parse(dateLayout, createdAt)
	if lastAttemptedAt != "" {
		e.LastAttemptedAt, _ = time.Parse(dateLayout, lastAttemptedAt)
	}
	return e, nil
}

// scanEntries scans multiple rows into a slice of Entries.
func scanEntries(rows *sql.Rows) ([]domain.Entry, error) {
	var entries []domain.Entry
	for rows.Next() {
		e, err := scanEntryFromRows(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
