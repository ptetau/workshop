package message

import (
	"context"
	"database/sql"
	"time"

	domain "workshop/internal/domain/message"
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

// GetByID retrieves a Message by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Message, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, sender_id, receiver_id, subject, content, read_at, created_at
		 FROM message WHERE id = ?`, id)
	return scanMessage(row)
}

// Save persists a Message to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, m domain.Message) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO message (id, sender_id, receiver_id, subject, content, read_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   sender_id=excluded.sender_id, receiver_id=excluded.receiver_id,
		   subject=excluded.subject, content=excluded.content,
		   read_at=excluded.read_at, created_at=excluded.created_at`,
		m.ID, m.SenderID, m.ReceiverID, nullStr(m.Subject), m.Content,
		nullTime(m.ReadAt), m.CreatedAt.Format(timeLayout))
	return err
}

// Delete removes a Message from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM message WHERE id = ?`, id)
	return err
}

// ListByReceiverID retrieves Messages for a receiver.
// PRE: receiverID is non-empty
// POST: Returns messages for the given receiver
func (s *SQLiteStore) ListByReceiverID(ctx context.Context, receiverID string) ([]domain.Message, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, sender_id, receiver_id, subject, content, read_at, created_at
		 FROM message WHERE receiver_id = ? ORDER BY created_at DESC`, receiverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

// CountUnread counts unread messages for a receiver.
// PRE: receiverID is non-empty
// POST: Returns count of unread messages
func (s *SQLiteStore) CountUnread(ctx context.Context, receiverID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM message WHERE receiver_id = ? AND read_at IS NULL`, receiverID).Scan(&count)
	return count, err
}

func scanMessage(row *sql.Row) (domain.Message, error) {
	var m domain.Message
	var subject, readAt sql.NullString
	var createdAt string
	err := row.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &subject, &m.Content, &readAt, &createdAt)
	if err != nil {
		return domain.Message{}, err
	}
	m.CreatedAt, _ = time.Parse(timeLayout, createdAt)
	if subject.Valid {
		m.Subject = subject.String
	}
	if readAt.Valid {
		m.ReadAt, _ = time.Parse(timeLayout, readAt.String)
	}
	return m, nil
}

func scanMessages(rows *sql.Rows) ([]domain.Message, error) {
	var messages []domain.Message
	for rows.Next() {
		var m domain.Message
		var subject, readAt sql.NullString
		var createdAt string
		err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &subject, &m.Content, &readAt, &createdAt)
		if err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(timeLayout, createdAt)
		if subject.Valid {
			m.Subject = subject.String
		}
		if readAt.Valid {
			m.ReadAt, _ = time.Parse(timeLayout, readAt.String)
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t.Format(timeLayout)
}
