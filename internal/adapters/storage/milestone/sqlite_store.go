package milestone

import (
	"context"
	"database/sql"

	domain "workshop/internal/domain/milestone"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLiteStore.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a Milestone by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Milestone, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, metric, threshold, badge_icon FROM milestone WHERE id = ?`, id)
	var m domain.Milestone
	var badgeIcon sql.NullString
	err := row.Scan(&m.ID, &m.Name, &m.Metric, &m.Threshold, &badgeIcon)
	if err != nil {
		return domain.Milestone{}, err
	}
	if badgeIcon.Valid {
		m.BadgeIcon = badgeIcon.String
	}
	return m, nil
}

// Save persists a Milestone to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, m domain.Milestone) error {
	var icon interface{}
	if m.BadgeIcon != "" {
		icon = m.BadgeIcon
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO milestone (id, name, metric, threshold, badge_icon)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   name=excluded.name, metric=excluded.metric, threshold=excluded.threshold, badge_icon=excluded.badge_icon`,
		m.ID, m.Name, m.Metric, m.Threshold, icon)
	return err
}

// Delete removes a Milestone from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM milestone WHERE id = ?`, id)
	return err
}

// List retrieves all Milestones.
// PRE: none
// POST: Returns all milestones ordered by metric and threshold
func (s *SQLiteStore) List(ctx context.Context) ([]domain.Milestone, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, metric, threshold, badge_icon FROM milestone ORDER BY metric, threshold`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var milestones []domain.Milestone
	for rows.Next() {
		var m domain.Milestone
		var badgeIcon sql.NullString
		if err := rows.Scan(&m.ID, &m.Name, &m.Metric, &m.Threshold, &badgeIcon); err != nil {
			return nil, err
		}
		if badgeIcon.Valid {
			m.BadgeIcon = badgeIcon.String
		}
		milestones = append(milestones, m)
	}
	return milestones, rows.Err()
}
