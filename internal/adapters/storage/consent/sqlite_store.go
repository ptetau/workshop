package consent

import (
	"context"
	"database/sql"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/consent"
)

const dateLayout = "2006-01-02T15:04:05.999999999Z07:00"

// SQLiteStore implements the consent Store interface using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new consent store.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// Save persists a consent record.
// PRE: consent is valid
// POST: Consent is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, c domain.Consent) error {
	revokedAt := ""
	if c.RevokedAt != nil {
		revokedAt = c.RevokedAt.Format(dateLayout)
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO consent (id, member_id, type, granted, granted_at, revoked_at, source, ip_address, user_agent, version)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   granted=excluded.granted,
		   revoked_at=excluded.revoked_at`,
		c.ID, c.MemberID, string(c.Type), c.Granted, c.GrantedAt.Format(dateLayout),
		revokedAt, c.Source, c.IPAddress, c.UserAgent, c.Version)
	return err
}

// GetByMemberID retrieves all consent records for a member.
// PRE: memberID is non-empty
// POST: Returns all consent records for the member
func (s *SQLiteStore) GetByMemberID(ctx context.Context, memberID string) ([]domain.Consent, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, type, granted, granted_at, revoked_at, source, ip_address, user_agent, version
		 FROM consent WHERE member_id = ? ORDER BY granted_at DESC`,
		memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanConsents(rows)
}

// GetByType retrieves a specific consent type for a member.
// PRE: memberID and consentType are non-empty
// POST: Returns the consent or error if not found
func (s *SQLiteStore) GetByType(ctx context.Context, memberID string, consentType domain.Type) (domain.Consent, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, type, granted, granted_at, revoked_at, source, ip_address, user_agent, version
		 FROM consent WHERE member_id = ? AND type = ? ORDER BY granted_at DESC LIMIT 1`,
		memberID, string(consentType))
	return scanConsent(row)
}

// HasValidConsent checks if member has valid consent of a specific type.
// PRE: memberID and consentType are non-empty
// POST: Returns true if consent exists, granted, and not revoked
func (s *SQLiteStore) HasValidConsent(ctx context.Context, memberID string, consentType domain.Type) (bool, error) {
	c, err := s.GetByType(ctx, memberID, consentType)
	if err != nil {
		return false, nil // No consent record means no consent
	}
	return c.IsValid(), nil
}

// scanConsent scans a single row into a Consent.
func scanConsent(row *sql.Row) (domain.Consent, error) {
	var c domain.Consent
	var grantedAt, revokedAt string
	err := row.Scan(&c.ID, &c.MemberID, &c.Type, &c.Granted, &grantedAt, &revokedAt, &c.Source, &c.IPAddress, &c.UserAgent, &c.Version)
	if err != nil {
		return domain.Consent{}, err
	}
	c.GrantedAt, _ = time.Parse(dateLayout, grantedAt)
	if revokedAt != "" {
		t, _ := time.Parse(dateLayout, revokedAt)
		c.RevokedAt = &t
	}
	return c, nil
}

// scanConsentFromRows scans a single row from Rows into a Consent.
func scanConsentFromRows(rows *sql.Rows) (domain.Consent, error) {
	var c domain.Consent
	var grantedAt, revokedAt string
	err := rows.Scan(&c.ID, &c.MemberID, &c.Type, &c.Granted, &grantedAt, &revokedAt, &c.Source, &c.IPAddress, &c.UserAgent, &c.Version)
	if err != nil {
		return domain.Consent{}, err
	}
	c.GrantedAt, _ = time.Parse(dateLayout, grantedAt)
	if revokedAt != "" {
		t, _ := time.Parse(dateLayout, revokedAt)
		c.RevokedAt = &t
	}
	return c, nil
}

// scanConsents scans multiple rows into a slice of Consents.
func scanConsents(rows *sql.Rows) ([]domain.Consent, error) {
	var consents []domain.Consent
	for rows.Next() {
		c, err := scanConsentFromRows(rows)
		if err != nil {
			return nil, err
		}
		consents = append(consents, c)
	}
	return consents, rows.Err()
}
