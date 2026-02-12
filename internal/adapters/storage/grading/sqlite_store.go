package grading

import (
	"context"
	"database/sql"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/grading"
)

const timeLayout = "2006-01-02T15:04:05Z07:00"

// RecordSQLiteStore implements RecordStore using SQLite.
type RecordSQLiteStore struct {
	db storage.SQLDB
}

// NewRecordSQLiteStore creates a new RecordSQLiteStore.
func NewRecordSQLiteStore(db storage.SQLDB) *RecordSQLiteStore {
	return &RecordSQLiteStore{db: db}
}

// GetByID retrieves a grading Record by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *RecordSQLiteStore) GetByID(ctx context.Context, id string) (domain.Record, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, belt, stripe, promoted_at, proposed_by, approved_by, method
		 FROM grading_record WHERE id = ?`, id)
	return scanRecord(row)
}

// Save persists a grading Record to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *RecordSQLiteStore) Save(ctx context.Context, r domain.Record) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO grading_record (id, member_id, belt, stripe, promoted_at, proposed_by, approved_by, method)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   member_id=excluded.member_id, belt=excluded.belt, stripe=excluded.stripe,
		   promoted_at=excluded.promoted_at, proposed_by=excluded.proposed_by,
		   approved_by=excluded.approved_by, method=excluded.method`,
		r.ID, r.MemberID, r.Belt, r.Stripe, r.PromotedAt.Format(timeLayout),
		nullStr(r.ProposedBy), nullStr(r.ApprovedBy), r.Method)
	return err
}

// ListByMemberID retrieves grading Records for a member.
// PRE: memberID is non-empty
// POST: Returns records for the given member
func (s *RecordSQLiteStore) ListByMemberID(ctx context.Context, memberID string) ([]domain.Record, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, belt, stripe, promoted_at, proposed_by, approved_by, method
		 FROM grading_record WHERE member_id = ? ORDER BY promoted_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []domain.Record
	for rows.Next() {
		var r domain.Record
		var promotedAt string
		var proposedBy, approvedBy sql.NullString
		if err := rows.Scan(&r.ID, &r.MemberID, &r.Belt, &r.Stripe, &promotedAt, &proposedBy, &approvedBy, &r.Method); err != nil {
			return nil, err
		}
		r.PromotedAt, _ = time.Parse(timeLayout, promotedAt)
		if proposedBy.Valid {
			r.ProposedBy = proposedBy.String
		}
		if approvedBy.Valid {
			r.ApprovedBy = approvedBy.String
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func scanRecord(row *sql.Row) (domain.Record, error) {
	var r domain.Record
	var promotedAt string
	var proposedBy, approvedBy sql.NullString
	err := row.Scan(&r.ID, &r.MemberID, &r.Belt, &r.Stripe, &promotedAt, &proposedBy, &approvedBy, &r.Method)
	if err != nil {
		return domain.Record{}, err
	}
	r.PromotedAt, _ = time.Parse(timeLayout, promotedAt)
	if proposedBy.Valid {
		r.ProposedBy = proposedBy.String
	}
	if approvedBy.Valid {
		r.ApprovedBy = approvedBy.String
	}
	return r, nil
}

// ConfigSQLiteStore implements ConfigStore using SQLite.
type ConfigSQLiteStore struct {
	db storage.SQLDB
}

// NewConfigSQLiteStore creates a new ConfigSQLiteStore.
func NewConfigSQLiteStore(db storage.SQLDB) *ConfigSQLiteStore {
	return &ConfigSQLiteStore{db: db}
}

// GetByID retrieves a grading Config by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *ConfigSQLiteStore) GetByID(ctx context.Context, id string) (domain.Config, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, program, belt, flight_time_hours, attendance_pct, stripe_count
		 FROM grading_config WHERE id = ?`, id)
	var c domain.Config
	err := row.Scan(&c.ID, &c.Program, &c.Belt, &c.FlightTimeHours, &c.AttendancePct, &c.StripeCount)
	return c, err
}

// Save persists a grading Config to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *ConfigSQLiteStore) Save(ctx context.Context, c domain.Config) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO grading_config (id, program, belt, flight_time_hours, attendance_pct, stripe_count)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   program=excluded.program, belt=excluded.belt,
		   flight_time_hours=excluded.flight_time_hours, attendance_pct=excluded.attendance_pct,
		   stripe_count=excluded.stripe_count`,
		c.ID, c.Program, c.Belt, c.FlightTimeHours, c.AttendancePct, c.StripeCount)
	return err
}

// GetByProgramAndBelt retrieves a Config by program and belt.
// PRE: program and belt are non-empty
// POST: Returns the matching config or error
func (s *ConfigSQLiteStore) GetByProgramAndBelt(ctx context.Context, program, belt string) (domain.Config, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, program, belt, flight_time_hours, attendance_pct, stripe_count
		 FROM grading_config WHERE program = ? AND belt = ?`, program, belt)
	var c domain.Config
	err := row.Scan(&c.ID, &c.Program, &c.Belt, &c.FlightTimeHours, &c.AttendancePct, &c.StripeCount)
	return c, err
}

// List retrieves all grading Configs.
// PRE: none
// POST: Returns all configs ordered by program and belt
func (s *ConfigSQLiteStore) List(ctx context.Context) ([]domain.Config, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, program, belt, flight_time_hours, attendance_pct, stripe_count
		 FROM grading_config ORDER BY program, belt`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.Config
	for rows.Next() {
		var c domain.Config
		if err := rows.Scan(&c.ID, &c.Program, &c.Belt, &c.FlightTimeHours, &c.AttendancePct, &c.StripeCount); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

// ProposalSQLiteStore implements ProposalStore using SQLite.
type ProposalSQLiteStore struct {
	db storage.SQLDB
}

// NewProposalSQLiteStore creates a new ProposalSQLiteStore.
func NewProposalSQLiteStore(db storage.SQLDB) *ProposalSQLiteStore {
	return &ProposalSQLiteStore{db: db}
}

// GetByID retrieves a grading Proposal by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *ProposalSQLiteStore) GetByID(ctx context.Context, id string) (domain.Proposal, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, target_belt, notes, proposed_by, approved_by, status, created_at, decided_at
		 FROM grading_proposal WHERE id = ?`, id)
	return scanProposal(row)
}

// Save persists a grading Proposal to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *ProposalSQLiteStore) Save(ctx context.Context, p domain.Proposal) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO grading_proposal (id, member_id, target_belt, notes, proposed_by, approved_by, status, created_at, decided_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   member_id=excluded.member_id, target_belt=excluded.target_belt, notes=excluded.notes,
		   proposed_by=excluded.proposed_by, approved_by=excluded.approved_by, status=excluded.status,
		   created_at=excluded.created_at, decided_at=excluded.decided_at`,
		p.ID, p.MemberID, p.TargetBelt, nullStr(p.Notes), p.ProposedBy,
		nullStr(p.ApprovedBy), p.Status, p.CreatedAt.Format(timeLayout), nullTime(p.DecidedAt))
	return err
}

// ListPending retrieves all pending grading Proposals.
// PRE: none
// POST: Returns pending proposals ordered by creation time
func (s *ProposalSQLiteStore) ListPending(ctx context.Context) ([]domain.Proposal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, target_belt, notes, proposed_by, approved_by, status, created_at, decided_at
		 FROM grading_proposal WHERE status = 'pending' ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProposals(rows)
}

// ListByMemberID retrieves grading Proposals for a member.
// PRE: memberID is non-empty
// POST: Returns proposals for the given member
func (s *ProposalSQLiteStore) ListByMemberID(ctx context.Context, memberID string) ([]domain.Proposal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, target_belt, notes, proposed_by, approved_by, status, created_at, decided_at
		 FROM grading_proposal WHERE member_id = ? ORDER BY created_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProposals(rows)
}

func scanProposal(row *sql.Row) (domain.Proposal, error) {
	var p domain.Proposal
	var notes, approvedBy, decidedAt sql.NullString
	var createdAt string
	err := row.Scan(&p.ID, &p.MemberID, &p.TargetBelt, &notes, &p.ProposedBy, &approvedBy, &p.Status, &createdAt, &decidedAt)
	if err != nil {
		return domain.Proposal{}, err
	}
	p.CreatedAt, _ = time.Parse(timeLayout, createdAt)
	if notes.Valid {
		p.Notes = notes.String
	}
	if approvedBy.Valid {
		p.ApprovedBy = approvedBy.String
	}
	if decidedAt.Valid {
		p.DecidedAt, _ = time.Parse(timeLayout, decidedAt.String)
	}
	return p, nil
}

func scanProposals(rows *sql.Rows) ([]domain.Proposal, error) {
	var proposals []domain.Proposal
	for rows.Next() {
		var p domain.Proposal
		var notes, approvedBy, decidedAt sql.NullString
		var createdAt string
		err := rows.Scan(&p.ID, &p.MemberID, &p.TargetBelt, &notes, &p.ProposedBy, &approvedBy, &p.Status, &createdAt, &decidedAt)
		if err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse(timeLayout, createdAt)
		if notes.Valid {
			p.Notes = notes.String
		}
		if approvedBy.Valid {
			p.ApprovedBy = approvedBy.String
		}
		if decidedAt.Valid {
			p.DecidedAt, _ = time.Parse(timeLayout, decidedAt.String)
		}
		proposals = append(proposals, p)
	}
	return proposals, rows.Err()
}

// --- MemberConfigSQLiteStore ---

// MemberConfigSQLiteStore implements MemberConfigStore using SQLite.
type MemberConfigSQLiteStore struct {
	db storage.SQLDB
}

// NewMemberConfigSQLiteStore creates a new MemberConfigSQLiteStore.
func NewMemberConfigSQLiteStore(db storage.SQLDB) *MemberConfigSQLiteStore {
	return &MemberConfigSQLiteStore{db: db}
}

// Save persists a MemberConfig to the database (upsert on member_id+belt).
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *MemberConfigSQLiteStore) Save(ctx context.Context, mc domain.MemberConfig) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO grading_member_config (id, member_id, belt, flight_time_hours, attendance_pct)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(member_id, belt) DO UPDATE SET
		   flight_time_hours=excluded.flight_time_hours,
		   attendance_pct=excluded.attendance_pct`,
		mc.ID, mc.MemberID, mc.Belt, mc.FlightTimeHours, mc.AttendancePct)
	return err
}

// GetByMemberAndBelt retrieves a MemberConfig for a specific member and belt.
// PRE: memberID and belt are non-empty
// POST: Returns the config or sql.ErrNoRows
func (s *MemberConfigSQLiteStore) GetByMemberAndBelt(ctx context.Context, memberID, belt string) (domain.MemberConfig, error) {
	var mc domain.MemberConfig
	err := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, belt, flight_time_hours, attendance_pct
		 FROM grading_member_config WHERE member_id = ? AND belt = ?`, memberID, belt).
		Scan(&mc.ID, &mc.MemberID, &mc.Belt, &mc.FlightTimeHours, &mc.AttendancePct)
	return mc, err
}

// ListByMemberID retrieves all MemberConfig entries for a member.
// PRE: memberID is non-empty
// POST: Returns configs for the given member
func (s *MemberConfigSQLiteStore) ListByMemberID(ctx context.Context, memberID string) ([]domain.MemberConfig, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, belt, flight_time_hours, attendance_pct
		 FROM grading_member_config WHERE member_id = ?`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var configs []domain.MemberConfig
	for rows.Next() {
		var mc domain.MemberConfig
		if err := rows.Scan(&mc.ID, &mc.MemberID, &mc.Belt, &mc.FlightTimeHours, &mc.AttendancePct); err != nil {
			return nil, err
		}
		configs = append(configs, mc)
	}
	return configs, rows.Err()
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
