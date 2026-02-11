package schedule

import (
	"context"
	"database/sql"
	"fmt"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/schedule"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new ScheduleStore.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a Schedule by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Schedule, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, class_type_id, day, start_time, end_time FROM schedule WHERE id = ?", id)
	var entity domain.Schedule
	err := row.Scan(&entity.ID, &entity.ClassTypeID, &entity.Day, &entity.StartTime, &entity.EndTime)
	if err == sql.ErrNoRows {
		return domain.Schedule{}, fmt.Errorf("schedule not found: %w", err)
	}
	return entity, err
}

// Save persists a Schedule to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, entity domain.Schedule) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO schedule (id, class_type_id, day, start_time, end_time) VALUES (?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET class_type_id=excluded.class_type_id, day=excluded.day, start_time=excluded.start_time, end_time=excluded.end_time",
		entity.ID, entity.ClassTypeID, entity.Day, entity.StartTime, entity.EndTime,
	)
	return err
}

// Delete removes a Schedule from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM schedule WHERE id = ?", id)
	return err
}

// List retrieves all Schedules.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (s *SQLiteStore) List(ctx context.Context) ([]domain.Schedule, error) {
	return s.querySchedules(ctx, "SELECT id, class_type_id, day, start_time, end_time FROM schedule ORDER BY day, start_time")
}

// ListByDay retrieves Schedules for a specific day.
// PRE: day is a valid weekday
// POST: Returns schedules for the given day
func (s *SQLiteStore) ListByDay(ctx context.Context, day string) ([]domain.Schedule, error) {
	return s.querySchedules(ctx, "SELECT id, class_type_id, day, start_time, end_time FROM schedule WHERE day = ? ORDER BY start_time", day)
}

// ListByClassTypeID retrieves Schedules for a specific class type.
// PRE: classTypeID is non-empty
// POST: Returns schedules for the given class type
func (s *SQLiteStore) ListByClassTypeID(ctx context.Context, classTypeID string) ([]domain.Schedule, error) {
	return s.querySchedules(ctx, "SELECT id, class_type_id, day, start_time, end_time FROM schedule WHERE class_type_id = ? ORDER BY day, start_time", classTypeID)
}

func (s *SQLiteStore) querySchedules(ctx context.Context, query string, args ...interface{}) ([]domain.Schedule, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Schedule
	for rows.Next() {
		var entity domain.Schedule
		if err := rows.Scan(&entity.ID, &entity.ClassTypeID, &entity.Day, &entity.StartTime, &entity.EndTime); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return results, nil
}
