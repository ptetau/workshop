package audit

import (
	"context"
	"database/sql"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/audit"
)

const dateLayout = "2006-01-02T15:04:05.999999999Z07:00"

// SQLiteStore implements the audit Store interface using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new audit event store.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// Save persists an audit event.
// PRE: event is valid
// POST: Event is persisted
func (s *SQLiteStore) Save(ctx context.Context, event domain.Event) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_event (id, timestamp, category, action, severity, actor_id, actor_email, actor_role, resource_id, resource_type, description, ip_address, user_agent, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.Timestamp.Format(dateLayout), string(event.Category), string(event.Action),
		string(event.Severity), event.ActorID, event.ActorEmail, event.ActorRole,
		event.ResourceID, event.ResourceType, event.Description, event.IPAddress, event.UserAgent, event.Metadata)
	return err
}

// List returns audit events with optional filtering.
// PRE: limit > 0
// POST: Returns events ordered by timestamp desc
func (s *SQLiteStore) List(ctx context.Context, filter Filter, limit int) ([]domain.Event, error) {
	query := `SELECT id, timestamp, category, action, severity, actor_id, actor_email, actor_role, resource_id, resource_type, description, ip_address, user_agent, metadata FROM audit_event WHERE 1=1`
	args := []interface{}{}

	if filter.Category != nil {
		query += " AND category = ?"
		args = append(args, string(*filter.Category))
	}
	if filter.Action != nil {
		query += " AND action = ?"
		args = append(args, string(*filter.Action))
	}
	if filter.ActorID != nil {
		query += " AND actor_id = ?"
		args = append(args, *filter.ActorID)
	}
	if filter.Severity != nil {
		query += " AND severity = ?"
		args = append(args, string(*filter.Severity))
	}
	if filter.ResourceID != nil {
		query += " AND resource_id = ?"
		args = append(args, *filter.ResourceID)
	}
	if filter.FromDate != nil {
		query += " AND timestamp >= ?"
		args = append(args, *filter.FromDate)
	}
	if filter.ToDate != nil {
		query += " AND timestamp <= ?"
		args = append(args, *filter.ToDate)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

// GetByID retrieves a specific audit event.
// PRE: id is non-empty
// POST: Returns the event or error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Event, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, timestamp, category, action, severity, actor_id, actor_email, actor_role, resource_id, resource_type, description, ip_address, user_agent, metadata FROM audit_event WHERE id = ?`, id)
	return scanEvent(row)
}

// scanEvent scans a single row into an Event.
func scanEvent(row *sql.Row) (domain.Event, error) {
	var e domain.Event
	var timestamp string
	err := row.Scan(&e.ID, &timestamp, &e.Category, &e.Action, &e.Severity, &e.ActorID, &e.ActorEmail, &e.ActorRole, &e.ResourceID, &e.ResourceType, &e.Description, &e.IPAddress, &e.UserAgent, &e.Metadata)
	if err != nil {
		return domain.Event{}, err
	}
	e.Timestamp, _ = time.Parse(dateLayout, timestamp)
	return e, nil
}

// scanEventFromRows scans a single row from Rows into an Event.
func scanEventFromRows(rows *sql.Rows) (domain.Event, error) {
	var e domain.Event
	var timestamp string
	err := rows.Scan(&e.ID, &timestamp, &e.Category, &e.Action, &e.Severity, &e.ActorID, &e.ActorEmail, &e.ActorRole, &e.ResourceID, &e.ResourceType, &e.Description, &e.IPAddress, &e.UserAgent, &e.Metadata)
	if err != nil {
		return domain.Event{}, err
	}
	e.Timestamp, _ = time.Parse(dateLayout, timestamp)
	return e, nil
}

// scanEvents scans multiple rows into a slice of Events.
func scanEvents(rows *sql.Rows) ([]domain.Event, error) {
	var events []domain.Event
	for rows.Next() {
		e, err := scanEventFromRows(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
