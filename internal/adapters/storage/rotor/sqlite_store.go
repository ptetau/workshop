package rotor

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/rotor"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new SQLiteStore.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

const timeFormat = time.RFC3339

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(timeFormat, s)
	return t
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(timeFormat)
}

// --- Rotor CRUD ---

// SaveRotor inserts or updates a rotor.
// PRE: r is a valid Rotor
// POST: rotor is persisted
func (s *SQLiteStore) SaveRotor(ctx context.Context, r domain.Rotor) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO rotor (id, class_type_id, name, version, status, preview_on, created_by, created_at, activated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   class_type_id=excluded.class_type_id, name=excluded.name, version=excluded.version,
		   status=excluded.status, preview_on=excluded.preview_on, created_by=excluded.created_by,
		   created_at=excluded.created_at, activated_at=excluded.activated_at`,
		r.ID, r.ClassTypeID, r.Name, r.Version, r.Status,
		boolToInt(r.PreviewOn), r.CreatedBy, formatTime(r.CreatedAt), formatTime(r.ActivatedAt))
	return err
}

// GetRotor retrieves a rotor by ID.
// PRE: id is non-empty
// POST: returns the rotor or error if not found
func (s *SQLiteStore) GetRotor(ctx context.Context, id string) (domain.Rotor, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, class_type_id, name, version, status, preview_on, created_by, created_at, activated_at
		 FROM rotor WHERE id = ?`, id)
	return scanRotor(row)
}

// ListRotorsByClassType returns all rotors for a class type, ordered by version desc.
// PRE: classTypeID is non-empty
// POST: returns rotors or empty slice
func (s *SQLiteStore) ListRotorsByClassType(ctx context.Context, classTypeID string) ([]domain.Rotor, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, class_type_id, name, version, status, preview_on, created_by, created_at, activated_at
		 FROM rotor WHERE class_type_id = ? ORDER BY version DESC`, classTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Rotor
	for rows.Next() {
		r, err := scanRotorRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

// GetActiveRotor returns the active rotor for a class type.
// PRE: classTypeID is non-empty
// POST: returns the active rotor or error if none
func (s *SQLiteStore) GetActiveRotor(ctx context.Context, classTypeID string) (domain.Rotor, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, class_type_id, name, version, status, preview_on, created_by, created_at, activated_at
		 FROM rotor WHERE class_type_id = ? AND status = 'active' LIMIT 1`, classTypeID)
	return scanRotor(row)
}

// DeleteRotor deletes a rotor by ID (cascades to themes, topics, schedules).
// PRE: id is non-empty
// POST: rotor and all children are deleted
func (s *SQLiteStore) DeleteRotor(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM rotor WHERE id = ?`, id)
	return err
}

func scanRotor(row *sql.Row) (domain.Rotor, error) {
	var r domain.Rotor
	var previewOn int
	var createdAt, activatedAt string
	err := row.Scan(&r.ID, &r.ClassTypeID, &r.Name, &r.Version, &r.Status,
		&previewOn, &r.CreatedBy, &createdAt, &activatedAt)
	if err != nil {
		return domain.Rotor{}, err
	}
	r.PreviewOn = previewOn == 1
	r.CreatedAt = parseTime(createdAt)
	r.ActivatedAt = parseTime(activatedAt)
	return r, nil
}

func scanRotorRows(rows *sql.Rows) (domain.Rotor, error) {
	var r domain.Rotor
	var previewOn int
	var createdAt, activatedAt string
	err := rows.Scan(&r.ID, &r.ClassTypeID, &r.Name, &r.Version, &r.Status,
		&previewOn, &r.CreatedBy, &createdAt, &activatedAt)
	if err != nil {
		return domain.Rotor{}, err
	}
	r.PreviewOn = previewOn == 1
	r.CreatedAt = parseTime(createdAt)
	r.ActivatedAt = parseTime(activatedAt)
	return r, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// --- RotorTheme CRUD ---

// SaveRotorTheme inserts or updates a theme.
// PRE: t is a valid RotorTheme
// POST: theme is persisted
func (s *SQLiteStore) SaveRotorTheme(ctx context.Context, t domain.RotorTheme) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO rotor_theme (id, rotor_id, name, position, hidden)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   rotor_id=excluded.rotor_id, name=excluded.name, position=excluded.position, hidden=excluded.hidden`,
		t.ID, t.RotorID, t.Name, t.Position, boolToInt(t.Hidden))
	return err
}

// ListThemesByRotor returns all themes for a rotor, ordered by position.
// PRE: rotorID is non-empty
// POST: returns themes or empty slice
func (s *SQLiteStore) ListThemesByRotor(ctx context.Context, rotorID string) ([]domain.RotorTheme, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, rotor_id, name, position, hidden FROM rotor_theme WHERE rotor_id = ? ORDER BY position`, rotorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.RotorTheme
	for rows.Next() {
		var t domain.RotorTheme
		var hidden int
		if err := rows.Scan(&t.ID, &t.RotorID, &t.Name, &t.Position, &hidden); err != nil {
			return nil, err
		}
		t.Hidden = hidden == 1
		result = append(result, t)
	}
	return result, rows.Err()
}

// DeleteRotorTheme deletes a theme by ID (cascades to topics).
// PRE: id is non-empty
// POST: theme and all children are deleted
func (s *SQLiteStore) DeleteRotorTheme(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM rotor_theme WHERE id = ?`, id)
	return err
}

// --- Topic CRUD ---

// SaveTopic inserts or updates a topic.
// PRE: t is a valid Topic
// POST: topic is persisted
func (s *SQLiteStore) SaveTopic(ctx context.Context, t domain.Topic) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO topic (id, rotor_theme_id, name, description, duration_weeks, position, last_covered)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   rotor_theme_id=excluded.rotor_theme_id, name=excluded.name, description=excluded.description,
		   duration_weeks=excluded.duration_weeks, position=excluded.position, last_covered=excluded.last_covered`,
		t.ID, t.RotorThemeID, t.Name, t.Description, t.DurationWeeks, t.Position, formatTime(t.LastCovered))
	return err
}

// GetTopic retrieves a topic by ID.
// PRE: id is non-empty
// POST: returns the topic or error if not found
func (s *SQLiteStore) GetTopic(ctx context.Context, id string) (domain.Topic, error) {
	var t domain.Topic
	var lastCovered string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, rotor_theme_id, name, description, duration_weeks, position, last_covered
		 FROM topic WHERE id = ?`, id).
		Scan(&t.ID, &t.RotorThemeID, &t.Name, &t.Description, &t.DurationWeeks, &t.Position, &lastCovered)
	if err != nil {
		return domain.Topic{}, err
	}
	t.LastCovered = parseTime(lastCovered)
	return t, nil
}

// ListTopicsByTheme returns all topics for a theme, ordered by position.
// PRE: rotorThemeID is non-empty
// POST: returns topics or empty slice
func (s *SQLiteStore) ListTopicsByTheme(ctx context.Context, rotorThemeID string) ([]domain.Topic, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, rotor_theme_id, name, description, duration_weeks, position, last_covered
		 FROM topic WHERE rotor_theme_id = ? ORDER BY position`, rotorThemeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Topic
	for rows.Next() {
		var t domain.Topic
		var lastCovered string
		if err := rows.Scan(&t.ID, &t.RotorThemeID, &t.Name, &t.Description, &t.DurationWeeks, &t.Position, &lastCovered); err != nil {
			return nil, err
		}
		t.LastCovered = parseTime(lastCovered)
		result = append(result, t)
	}
	return result, rows.Err()
}

// DeleteTopic deletes a topic by ID.
// PRE: id is non-empty
// POST: topic is deleted
func (s *SQLiteStore) DeleteTopic(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM topic WHERE id = ?`, id)
	return err
}

// ReorderTopics updates the position of topics within a theme.
// PRE: rotorThemeID is non-empty, topicIDs contains all topic IDs for the theme
// POST: positions are updated to match the order of topicIDs
func (s *SQLiteStore) ReorderTopics(ctx context.Context, rotorThemeID string, topicIDs []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range topicIDs {
		_, err := tx.ExecContext(ctx,
			`UPDATE topic SET position = ? WHERE id = ? AND rotor_theme_id = ?`,
			i, id, rotorThemeID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// --- TopicSchedule ---

// SaveTopicSchedule inserts or updates a schedule entry.
// PRE: sched is a valid TopicSchedule
// POST: schedule is persisted
func (s *SQLiteStore) SaveTopicSchedule(ctx context.Context, sched domain.TopicSchedule) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO topic_schedule (id, topic_id, rotor_theme_id, start_date, end_date, status)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   topic_id=excluded.topic_id, rotor_theme_id=excluded.rotor_theme_id,
		   start_date=excluded.start_date, end_date=excluded.end_date, status=excluded.status`,
		sched.ID, sched.TopicID, sched.RotorThemeID,
		formatTime(sched.StartDate), formatTime(sched.EndDate), sched.Status)
	return err
}

// GetActiveScheduleForTheme returns the currently active schedule for a theme.
// PRE: rotorThemeID is non-empty
// POST: returns the active schedule or error if none
func (s *SQLiteStore) GetActiveScheduleForTheme(ctx context.Context, rotorThemeID string) (domain.TopicSchedule, error) {
	var sched domain.TopicSchedule
	var startDate, endDate string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, topic_id, rotor_theme_id, start_date, end_date, status
		 FROM topic_schedule WHERE rotor_theme_id = ? AND status = 'active' LIMIT 1`,
		rotorThemeID).
		Scan(&sched.ID, &sched.TopicID, &sched.RotorThemeID, &startDate, &endDate, &sched.Status)
	if err != nil {
		return domain.TopicSchedule{}, err
	}
	sched.StartDate = parseTime(startDate)
	sched.EndDate = parseTime(endDate)
	return sched, nil
}

// ListSchedulesByTheme returns all schedules for a theme, ordered by start date desc.
// PRE: rotorThemeID is non-empty
// POST: returns schedules or empty slice
func (s *SQLiteStore) ListSchedulesByTheme(ctx context.Context, rotorThemeID string) ([]domain.TopicSchedule, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, topic_id, rotor_theme_id, start_date, end_date, status
		 FROM topic_schedule WHERE rotor_theme_id = ? ORDER BY start_date DESC`, rotorThemeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.TopicSchedule
	for rows.Next() {
		var sched domain.TopicSchedule
		var startDate, endDate string
		if err := rows.Scan(&sched.ID, &sched.TopicID, &sched.RotorThemeID, &startDate, &endDate, &sched.Status); err != nil {
			return nil, err
		}
		sched.StartDate = parseTime(startDate)
		sched.EndDate = parseTime(endDate)
		result = append(result, sched)
	}
	return result, rows.Err()
}

// --- Votes ---

// SaveVote inserts a vote.
// PRE: v is a valid Vote
// POST: vote is persisted (fails if duplicate topic_id+account_id)
func (s *SQLiteStore) SaveVote(ctx context.Context, v domain.Vote) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO vote (id, topic_id, account_id, created_at)
		 VALUES (?, ?, ?, ?)`,
		v.ID, v.TopicID, v.AccountID, formatTime(v.CreatedAt))
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint") {
		return domain.ErrAlreadyVoted
	}
	return err
}

// CountVotesForTopic returns the number of votes for a topic.
// PRE: topicID is non-empty
// POST: returns the vote count
func (s *SQLiteStore) CountVotesForTopic(ctx context.Context, topicID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM vote WHERE topic_id = ?`, topicID).Scan(&count)
	return count, err
}

// HasVoted checks if an account has voted for a topic.
// PRE: topicID and accountID are non-empty
// POST: returns true if a vote exists
func (s *SQLiteStore) HasVoted(ctx context.Context, topicID, accountID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM vote WHERE topic_id = ? AND account_id = ?`,
		topicID, accountID).Scan(&count)
	return count > 0, err
}

// DeleteVotesForTopic deletes all votes for a topic.
// PRE: topicID is non-empty
// POST: all votes for the topic are deleted
func (s *SQLiteStore) DeleteVotesForTopic(ctx context.Context, topicID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM vote WHERE topic_id = ?`, topicID)
	return err
}

// Verify interface compliance at compile time.
var _ Store = (*SQLiteStore)(nil)

// Ensure unused import is used
var _ = fmt.Sprintf
