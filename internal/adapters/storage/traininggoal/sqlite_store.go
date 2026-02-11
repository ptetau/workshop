package traininggoal

import (
	"context"
	"database/sql"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/traininggoal"
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

// GetByID retrieves a TrainingGoal by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.TrainingGoal, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, target, period, created_at, active FROM training_goal WHERE id = ?`, id)
	return scanGoal(row)
}

// Save persists a TrainingGoal to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, g domain.TrainingGoal) error {
	active := 0
	if g.Active {
		active = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO training_goal (id, member_id, target, period, created_at, active)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   member_id=excluded.member_id, target=excluded.target, period=excluded.period,
		   created_at=excluded.created_at, active=excluded.active`,
		g.ID, g.MemberID, g.Target, g.Period, g.CreatedAt.Format(timeLayout), active)
	return err
}

// Delete removes a TrainingGoal from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM training_goal WHERE id = ?`, id)
	return err
}

// GetActiveByMemberID retrieves the active TrainingGoal for a member.
// PRE: memberID is non-empty
// POST: Returns the active goal or error if none
func (s *SQLiteStore) GetActiveByMemberID(ctx context.Context, memberID string) (domain.TrainingGoal, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, target, period, created_at, active
		 FROM training_goal WHERE member_id = ? AND active = 1 LIMIT 1`, memberID)
	return scanGoal(row)
}

// ListByMemberID retrieves TrainingGoals for a member.
// PRE: memberID is non-empty
// POST: Returns goals for the given member
func (s *SQLiteStore) ListByMemberID(ctx context.Context, memberID string) ([]domain.TrainingGoal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, target, period, created_at, active
		 FROM training_goal WHERE member_id = ? ORDER BY created_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []domain.TrainingGoal
	for rows.Next() {
		var g domain.TrainingGoal
		var createdAt string
		var active int
		if err := rows.Scan(&g.ID, &g.MemberID, &g.Target, &g.Period, &createdAt, &active); err != nil {
			return nil, err
		}
		g.CreatedAt, _ = time.Parse(timeLayout, createdAt)
		g.Active = active == 1
		goals = append(goals, g)
	}
	return goals, rows.Err()
}

func scanGoal(row *sql.Row) (domain.TrainingGoal, error) {
	var g domain.TrainingGoal
	var createdAt string
	var active int
	err := row.Scan(&g.ID, &g.MemberID, &g.Target, &g.Period, &createdAt, &active)
	if err != nil {
		return domain.TrainingGoal{}, err
	}
	g.CreatedAt, _ = time.Parse(timeLayout, createdAt)
	g.Active = active == 1
	return g, nil
}
