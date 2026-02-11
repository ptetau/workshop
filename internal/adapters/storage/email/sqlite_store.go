package email

import (
	"context"
	"database/sql"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/email"
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

// GetByID retrieves an Email by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Email, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, subject, body, sender_id, status, scheduled_at, sent_at,
		        created_at, updated_at, resend_message_id, template_version_id
		 FROM email WHERE id = ?`, id)
	return scanEmail(row)
}

// Save persists an Email to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, e domain.Email) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO email (id, subject, body, sender_id, status, scheduled_at, sent_at,
		                    created_at, updated_at, resend_message_id, template_version_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   subject=excluded.subject, body=excluded.body, sender_id=excluded.sender_id,
		   status=excluded.status, scheduled_at=excluded.scheduled_at, sent_at=excluded.sent_at,
		   created_at=excluded.created_at, updated_at=excluded.updated_at,
		   resend_message_id=excluded.resend_message_id, template_version_id=excluded.template_version_id`,
		e.ID, e.Subject, e.Body, e.SenderID, e.Status,
		nullTime(e.ScheduledAt), nullTime(e.SentAt),
		e.CreatedAt.Format(timeLayout), nullTime(e.UpdatedAt),
		nullStr(e.ResendMessageID), nullStr(e.TemplateVersionID))
	return err
}

// Delete removes an Email from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed (CASCADE deletes recipients)
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM email WHERE id = ?`, id)
	return err
}

// List retrieves emails matching the filter.
// PRE: none
// POST: Returns matching emails sorted by created_at DESC
func (s *SQLiteStore) List(ctx context.Context, filter ListFilter) ([]domain.Email, error) {
	query := `SELECT id, subject, body, sender_id, status, scheduled_at, sent_at,
	                 created_at, updated_at, resend_message_id, template_version_id
	          FROM email WHERE 1=1`
	var args []interface{}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.SenderID != "" {
		query += " AND sender_id = ?"
		args = append(args, filter.SenderID)
	}
	if filter.Search != "" {
		query += " AND (subject LIKE ? OR body LIKE ?)"
		like := "%" + filter.Search + "%"
		args = append(args, like, like)
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmails(rows)
}

// SaveRecipients saves the recipient list for an email, replacing any existing.
// PRE: emailID exists
// POST: Recipients are persisted
func (s *SQLiteStore) SaveRecipients(ctx context.Context, emailID string, recipients []domain.Recipient) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM email_recipient WHERE email_id = ?`, emailID); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO email_recipient (email_id, member_id, member_name, member_email, delivery_status)
		 VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range recipients {
		if _, err := stmt.ExecContext(ctx, r.EmailID, r.MemberID, r.MemberName, r.MemberEmail, r.DeliveryStatus); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetRecipients retrieves all recipients for an email.
// PRE: emailID is non-empty
// POST: Returns recipient list
func (s *SQLiteStore) GetRecipients(ctx context.Context, emailID string) ([]domain.Recipient, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT email_id, member_id, member_name, member_email, delivery_status
		 FROM email_recipient WHERE email_id = ?`, emailID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []domain.Recipient
	for rows.Next() {
		var r domain.Recipient
		if err := rows.Scan(&r.EmailID, &r.MemberID, &r.MemberName, &r.MemberEmail, &r.DeliveryStatus); err != nil {
			return nil, err
		}
		recipients = append(recipients, r)
	}
	return recipients, rows.Err()
}

// ListByRecipientMemberID retrieves emails sent to a specific member.
// PRE: memberID is non-empty
// POST: Returns emails for the member's inbox, sorted by created_at DESC
func (s *SQLiteStore) ListByRecipientMemberID(ctx context.Context, memberID string) ([]domain.Email, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT e.id, e.subject, e.body, e.sender_id, e.status, e.scheduled_at, e.sent_at,
		        e.created_at, e.updated_at, e.resend_message_id, e.template_version_id
		 FROM email e
		 JOIN email_recipient er ON e.id = er.email_id
		 WHERE er.member_id = ? AND e.status = 'sent'
		 ORDER BY e.sent_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmails(rows)
}

func scanEmail(row *sql.Row) (domain.Email, error) {
	var e domain.Email
	var scheduledAt, sentAt, updatedAt, resendID, templateID sql.NullString
	var createdAt string
	err := row.Scan(&e.ID, &e.Subject, &e.Body, &e.SenderID, &e.Status,
		&scheduledAt, &sentAt, &createdAt, &updatedAt, &resendID, &templateID)
	if err != nil {
		return domain.Email{}, err
	}
	e.CreatedAt, _ = time.Parse(timeLayout, createdAt)
	if scheduledAt.Valid {
		e.ScheduledAt, _ = time.Parse(timeLayout, scheduledAt.String)
	}
	if sentAt.Valid {
		e.SentAt, _ = time.Parse(timeLayout, sentAt.String)
	}
	if updatedAt.Valid {
		e.UpdatedAt, _ = time.Parse(timeLayout, updatedAt.String)
	}
	if resendID.Valid {
		e.ResendMessageID = resendID.String
	}
	if templateID.Valid {
		e.TemplateVersionID = templateID.String
	}
	return e, nil
}

func scanEmails(rows *sql.Rows) ([]domain.Email, error) {
	var emails []domain.Email
	for rows.Next() {
		var e domain.Email
		var scheduledAt, sentAt, updatedAt, resendID, templateID sql.NullString
		var createdAt string
		err := rows.Scan(&e.ID, &e.Subject, &e.Body, &e.SenderID, &e.Status,
			&scheduledAt, &sentAt, &createdAt, &updatedAt, &resendID, &templateID)
		if err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(timeLayout, createdAt)
		if scheduledAt.Valid {
			e.ScheduledAt, _ = time.Parse(timeLayout, scheduledAt.String)
		}
		if sentAt.Valid {
			e.SentAt, _ = time.Parse(timeLayout, sentAt.String)
		}
		if updatedAt.Valid {
			e.UpdatedAt, _ = time.Parse(timeLayout, updatedAt.String)
		}
		if resendID.Valid {
			e.ResendMessageID = resendID.String
		}
		if templateID.Valid {
			e.TemplateVersionID = templateID.String
		}
		emails = append(emails, e)
	}
	return emails, rows.Err()
}

// SaveTemplate saves a new template version and deactivates the previous active one.
// PRE: t has a valid ID and CreatedAt
// POST: Template persisted; previous active template deactivated
func (s *SQLiteStore) SaveTemplate(ctx context.Context, t domain.EmailTemplate) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Deactivate all existing templates
	if _, err := tx.ExecContext(ctx, `UPDATE email_template SET active = 0 WHERE active = 1`); err != nil {
		return err
	}

	active := 0
	if t.Active {
		active = 1
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO email_template (id, header, footer, created_at, active) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET header=excluded.header, footer=excluded.footer, active=excluded.active`,
		t.ID, t.Header, t.Footer, t.CreatedAt.Format(timeLayout), active)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetActiveTemplate retrieves the currently active email template.
// PRE: none
// POST: Returns the active template or error if none exists
func (s *SQLiteStore) GetActiveTemplate(ctx context.Context) (domain.EmailTemplate, error) {
	var t domain.EmailTemplate
	var createdStr string
	var active int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, header, footer, created_at, active FROM email_template WHERE active = 1 LIMIT 1`).
		Scan(&t.ID, &t.Header, &t.Footer, &createdStr, &active)
	if err != nil {
		return domain.EmailTemplate{}, err
	}
	t.CreatedAt, _ = time.Parse(timeLayout, createdStr)
	t.Active = active == 1
	return t, nil
}

// GetTemplateByID retrieves a specific template version by ID.
// PRE: id is non-empty
// POST: Returns the template or error
func (s *SQLiteStore) GetTemplateByID(ctx context.Context, id string) (domain.EmailTemplate, error) {
	var t domain.EmailTemplate
	var createdStr string
	var active int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, header, footer, created_at, active FROM email_template WHERE id = ?`, id).
		Scan(&t.ID, &t.Header, &t.Footer, &createdStr, &active)
	if err != nil {
		return domain.EmailTemplate{}, err
	}
	t.CreatedAt, _ = time.Parse(timeLayout, createdStr)
	t.Active = active == 1
	return t, nil
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
