package milestone

import (
	"context"
	"database/sql"
	"time"

	domain "workshop/internal/domain/milestone"
)

// MemberMilestoneSQLiteStore implements MemberMilestoneStore using SQLite.
type MemberMilestoneSQLiteStore struct {
	db *sql.DB
}

// NewMemberMilestoneSQLiteStore creates a new MemberMilestoneSQLiteStore.
func NewMemberMilestoneSQLiteStore(db *sql.DB) *MemberMilestoneSQLiteStore {
	return &MemberMilestoneSQLiteStore{db: db}
}

// Save persists a MemberMilestone to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or ignore if already exists)
func (s *MemberMilestoneSQLiteStore) Save(ctx context.Context, value domain.MemberMilestone) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO member_milestone (id, member_id, milestone_id, earned_at, notified)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(member_id, milestone_id) DO NOTHING`,
		value.ID, value.MemberID, value.MilestoneID, value.EarnedAt.Format(time.RFC3339), boolToInt(value.Notified))
	return err
}

// ListByMemberID retrieves all milestones earned by a member.
// PRE: memberID is non-empty
// POST: Returns earned milestones ordered by earned_at descending
func (s *MemberMilestoneSQLiteStore) ListByMemberID(ctx context.Context, memberID string) ([]domain.MemberMilestone, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, milestone_id, earned_at, notified
		 FROM member_milestone WHERE member_id = ? ORDER BY earned_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemberMilestones(rows)
}

// MarkNotified marks a member milestone as notified.
// PRE: id is non-empty
// POST: The notified flag is set to true
func (s *MemberMilestoneSQLiteStore) MarkNotified(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE member_milestone SET notified = 1 WHERE id = ?`, id)
	return err
}

// ListUnnotifiedByMemberID retrieves unnotified milestones for a member.
// PRE: memberID is non-empty
// POST: Returns unnotified milestones ordered by earned_at descending
func (s *MemberMilestoneSQLiteStore) ListUnnotifiedByMemberID(ctx context.Context, memberID string) ([]domain.MemberMilestone, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, milestone_id, earned_at, notified
		 FROM member_milestone WHERE member_id = ? AND notified = 0 ORDER BY earned_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemberMilestones(rows)
}

func scanMemberMilestones(rows *sql.Rows) ([]domain.MemberMilestone, error) {
	var results []domain.MemberMilestone
	for rows.Next() {
		var item domain.MemberMilestone
		var earnedAt string
		var notified int
		if err := rows.Scan(&item.ID, &item.MemberID, &item.MilestoneID, &earnedAt, &notified); err != nil {
			return nil, err
		}
		item.EarnedAt, _ = time.Parse(time.RFC3339, earnedAt)
		item.Notified = notified != 0
		results = append(results, item)
	}
	return results, rows.Err()
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
