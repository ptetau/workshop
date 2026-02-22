package personalgoal

import (
	"context"
	"database/sql"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/personalgoal"
)

const timeLayout = "2006-01-02T15:04:05Z07:00"
const dateLayout = "2006-01-02"

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new SQLiteStore.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a PersonalGoal by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.PersonalGoal, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, member_id, title, description, target, unit, type, start_date, end_date, color, progress, created_at, updated_at
		 FROM personal_goal WHERE id = ?`, id)
	return scanGoal(row)
}

// Save persists a PersonalGoal to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, g domain.PersonalGoal) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO personal_goal (id, member_id, title, description, target, unit, type, start_date, end_date, color, progress, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   member_id=excluded.member_id, title=excluded.title, description=excluded.description,
		   target=excluded.target, unit=excluded.unit, type=excluded.type, start_date=excluded.start_date,
		   end_date=excluded.end_date, color=excluded.color, progress=excluded.progress,
		   created_at=excluded.created_at, updated_at=excluded.updated_at`,
		g.ID, g.MemberID, g.Title, g.Description, g.Target, g.Unit, g.Type,
		g.StartDate.Format(dateLayout), g.EndDate.Format(dateLayout), g.Color, g.Progress,
		g.CreatedAt.Format(timeLayout), g.UpdatedAt.Format(timeLayout))
	return err
}

// Delete removes a PersonalGoal from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM personal_goal WHERE id = ?`, id)
	return err
}

// ListByMemberID retrieves PersonalGoals for a member.
// PRE: memberID is non-empty
// POST: Returns goals for the given member, ordered by start date desc
func (s *SQLiteStore) ListByMemberID(ctx context.Context, memberID string) ([]domain.PersonalGoal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, title, description, target, unit, type, start_date, end_date, color, progress, created_at, updated_at
		 FROM personal_goal WHERE member_id = ? ORDER BY start_date DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanGoals(rows)
}

// ListByDateRange retrieves PersonalGoals that overlap with the given date range.
// PRE: memberID is non-empty, from and to are valid dates (YYYY-MM-DD)
// POST: Returns goals that overlap with the date range
func (s *SQLiteStore) ListByDateRange(ctx context.Context, memberID string, from, to string) ([]domain.PersonalGoal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, member_id, title, description, target, unit, type, start_date, end_date, color, progress, created_at, updated_at
		 FROM personal_goal WHERE member_id = ?
		 AND (start_date <= ? AND end_date >= ?)`,
		memberID, to, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanGoals(rows)
}

func scanGoal(row *sql.Row) (domain.PersonalGoal, error) {
	var g domain.PersonalGoal
	var createdAt, updatedAt string
	var startDate, endDate string
	err := row.Scan(&g.ID, &g.MemberID, &g.Title, &g.Description, &g.Target, &g.Unit, &g.Type,
		&startDate, &endDate, &g.Color, &g.Progress, &createdAt, &updatedAt)
	if err != nil {
		return domain.PersonalGoal{}, err
	}
	g.StartDate, _ = time.Parse(dateLayout, startDate)
	g.EndDate, _ = time.Parse(dateLayout, endDate)
	g.CreatedAt, _ = time.Parse(timeLayout, createdAt)
	g.UpdatedAt, _ = time.Parse(timeLayout, updatedAt)
	return g, nil
}

func scanGoals(rows *sql.Rows) ([]domain.PersonalGoal, error) {
	var goals []domain.PersonalGoal
	for rows.Next() {
		var g domain.PersonalGoal
		var createdAt, updatedAt string
		var startDate, endDate string
		if err := rows.Scan(&g.ID, &g.MemberID, &g.Title, &g.Description, &g.Target, &g.Unit, &g.Type,
			&startDate, &endDate, &g.Color, &g.Progress, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		g.StartDate, _ = time.Parse(dateLayout, startDate)
		g.EndDate, _ = time.Parse(dateLayout, endDate)
		g.CreatedAt, _ = time.Parse(timeLayout, createdAt)
		g.UpdatedAt, _ = time.Parse(timeLayout, updatedAt)
		goals = append(goals, g)
	}
	return goals, rows.Err()
}

// Ensure interface compliance at compile time.
var _ Store = (*SQLiteStore)(nil)
