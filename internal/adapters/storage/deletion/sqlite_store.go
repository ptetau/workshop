package deletion

import (
	"context"
	"database/sql"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/deletion"
)

const dateLayout = "2006-01-02T15:04:05.999999999Z07:00"

// SQLiteStore implements the deletion Store interface using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new deletion request store.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a deletion request by its ID.
// PRE: id is non-empty
// POST: Returns the request or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Request, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, email, status, requested_at, grace_period_end, confirmed_at, processed_at, cancelled_at, ip_address, user_agent
		 FROM deletion_request WHERE id = ?`, id)
	return scanRequest(row)
}

// GetByMemberID retrieves the most recent deletion request for a member.
// PRE: memberID is non-empty
// POST: Returns the request or an error if not found
func (s *SQLiteStore) GetByMemberID(ctx context.Context, memberID string) (domain.Request, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, email, status, requested_at, grace_period_end, confirmed_at, processed_at, cancelled_at, ip_address, user_agent
		 FROM deletion_request WHERE member_id = ? ORDER BY requested_at DESC LIMIT 1`, memberID)
	return scanRequest(row)
}

// Save persists a deletion request to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, r domain.Request) error {
	confirmedAt := ""
	if r.ConfirmedAt != nil {
		confirmedAt = r.ConfirmedAt.Format(dateLayout)
	}
	processedAt := ""
	if r.ProcessedAt != nil {
		processedAt = r.ProcessedAt.Format(dateLayout)
	}
	cancelledAt := ""
	if r.CancelledAt != nil {
		cancelledAt = r.CancelledAt.Format(dateLayout)
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO deletion_request (id, member_id, email, status, requested_at, grace_period_end, confirmed_at, processed_at, cancelled_at, ip_address, user_agent)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   status=excluded.status,
		   confirmed_at=excluded.confirmed_at,
		   processed_at=excluded.processed_at,
		   cancelled_at=excluded.cancelled_at`,
		r.ID, r.MemberID, r.Email, r.Status, r.RequestedAt.Format(dateLayout),
		r.GracePeriodEnd.Format(dateLayout), confirmedAt, processedAt, cancelledAt,
		r.IPAddress, r.UserAgent)
	return err
}

// ListPending returns deletion requests that need processing (confirmed, grace period ended).
// PRE: limit > 0
// POST: Returns up to limit entries ordered by grace_period_end
func (s *SQLiteStore) ListPending(ctx context.Context, limit int) ([]domain.Request, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, email, status, requested_at, grace_period_end, confirmed_at, processed_at, cancelled_at, ip_address, user_agent
		 FROM deletion_request WHERE status = ? AND grace_period_end <= ? ORDER BY grace_period_end ASC LIMIT ?`,
		domain.StatusConfirmed, time.Now().Format(dateLayout), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRequests(rows)
}

// ListByStatus returns deletion requests filtered by status.
// PRE: status is non-empty, limit > 0
// POST: Returns matching entries ordered by requested_at desc
func (s *SQLiteStore) ListByStatus(ctx context.Context, status string, limit int) ([]domain.Request, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, email, status, requested_at, grace_period_end, confirmed_at, processed_at, cancelled_at, ip_address, user_agent
		 FROM deletion_request WHERE status = ? ORDER BY requested_at DESC LIMIT ?`,
		status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRequests(rows)
}

// scanRequest scans a single row into a Request.
func scanRequest(row *sql.Row) (domain.Request, error) {
	var r domain.Request
	var requestedAt, gracePeriodEnd, confirmedAt, processedAt, cancelledAt string
	err := row.Scan(&r.ID, &r.MemberID, &r.Email, &r.Status, &requestedAt, &gracePeriodEnd, &confirmedAt, &processedAt, &cancelledAt, &r.IPAddress, &r.UserAgent)
	if err != nil {
		return domain.Request{}, err
	}
	r.RequestedAt, _ = time.Parse(dateLayout, requestedAt)
	r.GracePeriodEnd, _ = time.Parse(dateLayout, gracePeriodEnd)
	if confirmedAt != "" {
		t, _ := time.Parse(dateLayout, confirmedAt)
		r.ConfirmedAt = &t
	}
	if processedAt != "" {
		t, _ := time.Parse(dateLayout, processedAt)
		r.ProcessedAt = &t
	}
	if cancelledAt != "" {
		t, _ := time.Parse(dateLayout, cancelledAt)
		r.CancelledAt = &t
	}
	return r, nil
}

// scanRequestFromRows scans a single row from Rows into a Request.
func scanRequestFromRows(rows *sql.Rows) (domain.Request, error) {
	var r domain.Request
	var requestedAt, gracePeriodEnd, confirmedAt, processedAt, cancelledAt string
	err := rows.Scan(&r.ID, &r.MemberID, &r.Email, &r.Status, &requestedAt, &gracePeriodEnd, &confirmedAt, &processedAt, &cancelledAt, &r.IPAddress, &r.UserAgent)
	if err != nil {
		return domain.Request{}, err
	}
	r.RequestedAt, _ = time.Parse(dateLayout, requestedAt)
	r.GracePeriodEnd, _ = time.Parse(dateLayout, gracePeriodEnd)
	if confirmedAt != "" {
		t, _ := time.Parse(dateLayout, confirmedAt)
		r.ConfirmedAt = &t
	}
	if processedAt != "" {
		t, _ := time.Parse(dateLayout, processedAt)
		r.ProcessedAt = &t
	}
	if cancelledAt != "" {
		t, _ := time.Parse(dateLayout, cancelledAt)
		r.CancelledAt = &t
	}
	return r, nil
}

// scanRequests scans multiple rows into a slice of Requests.
func scanRequests(rows *sql.Rows) ([]domain.Request, error) {
	var requests []domain.Request
	for rows.Next() {
		r, err := scanRequestFromRows(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, rows.Err()
}
