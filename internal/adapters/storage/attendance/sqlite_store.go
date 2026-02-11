package attendance

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"workshop/internal/adapters/storage"
	domain "workshop/internal/domain/attendance"
)

// SQLiteStore implements domain.AttendanceStore using SQLite.
type SQLiteStore struct {
	db storage.SQLDB
}

// NewSQLiteStore creates a new AttendanceStore.
func NewSQLiteStore(db storage.SQLDB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// GetByID retrieves a Attendance by its ID.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (s *SQLiteStore) GetByID(ctx context.Context, id string) (domain.Attendance, error) {
	query := "SELECT id, check_in_time, check_out_time, member_id, schedule_id, class_date, mat_hours FROM attendance WHERE id = ?"

	row := s.db.QueryRowContext(ctx, query, id)

	var entity domain.Attendance
	var checkInStr string
	var checkOutStr, scheduleID, classDate sql.NullString
	err := row.Scan(
		&entity.ID,
		&checkInStr,
		&checkOutStr,
		&entity.MemberID,
		&scheduleID,
		&classDate,
		&entity.MatHours,
	)
	if scheduleID.Valid {
		entity.ScheduleID = scheduleID.String
	}
	if classDate.Valid {
		entity.ClassDate = classDate.String
	}
	if err == nil {
		// Parse check-in time (required)
		entity.CheckInTime, err = parseStoredTime(checkInStr)
		if err != nil {
			return domain.Attendance{}, fmt.Errorf("failed to parse check_in_time: %w", err)
		}
		// Parse check-out time (optional)
		if checkOutStr.Valid {
			parsedTime, parseErr := parseStoredTime(checkOutStr.String)
			if parseErr != nil {
				return domain.Attendance{}, fmt.Errorf("failed to parse check_out_time: %w", parseErr)
			}
			entity.CheckOutTime = parsedTime
		}
	}
	if err == sql.ErrNoRows {
		return domain.Attendance{}, fmt.Errorf("attendance not found: %w", err)
	}
	return entity, err
}

// Save persists a Attendance to the database.
// PRE: entity has been validated
// POST: Entity is persisted (insert or update)
func (s *SQLiteStore) Save(ctx context.Context, entity domain.Attendance) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Upsert implementation
	fields := []string{"id", "check_in_time", "check_out_time", "member_id", "schedule_id", "class_date", "mat_hours"}
	placeholders := []string{"?", "?", "?", "?", "?", "?", "?"}
	updates := []string{"check_in_time=excluded.check_in_time", "check_out_time=excluded.check_out_time", "member_id=excluded.member_id", "schedule_id=excluded.schedule_id", "class_date=excluded.class_date", "mat_hours=excluded.mat_hours"}

	query := fmt.Sprintf(
		"INSERT INTO attendance (%s) VALUES (%s) ON CONFLICT(id) DO UPDATE SET %s",
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updates, ", "),
	)

	// Format check-out time: NULL if zero, RFC3339 otherwise
	var checkOutValue interface{}
	if !entity.CheckOutTime.IsZero() {
		checkOutValue = entity.CheckOutTime.Format(time.RFC3339Nano)
	}

	var scheduleIDVal, classDateVal interface{}
	if entity.ScheduleID != "" {
		scheduleIDVal = entity.ScheduleID
	}
	if entity.ClassDate != "" {
		classDateVal = entity.ClassDate
	}

	_, err = tx.ExecContext(ctx, query,
		entity.ID,
		entity.CheckInTime.Format(time.RFC3339Nano),
		checkOutValue,
		entity.MemberID,
		scheduleIDVal,
		classDateVal,
		entity.MatHours,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete removes a Attendance from the database.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM attendance WHERE id = ?", id)
	return err
}

// List retrieves a list of Attendances based on the filter.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (s *SQLiteStore) List(ctx context.Context, filter ListFilter) ([]domain.Attendance, error) {
	query := "SELECT id, check_in_time, check_out_time, member_id, schedule_id, class_date, mat_hours FROM attendance LIMIT ? OFFSET ?"

	rows, err := s.db.QueryContext(ctx, query, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Attendance
	for rows.Next() {
		var entity domain.Attendance
		var checkInStr string
		var checkOutStr, scheduleID, classDate sql.NullString
		if err := rows.Scan(
			&entity.ID,
			&checkInStr,
			&checkOutStr,
			&entity.MemberID,
			&scheduleID,
			&classDate,
			&entity.MatHours,
		); err != nil {
			return nil, err
		}
		if scheduleID.Valid {
			entity.ScheduleID = scheduleID.String
		}
		if classDate.Valid {
			entity.ClassDate = classDate.String
		}
		// Parse check-in time (required)
		entity.CheckInTime, err = parseStoredTime(checkInStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse check_in_time: %w", err)
		}
		// Parse check-out time (optional)
		if checkOutStr.Valid {
			parsedTime, parseErr := parseStoredTime(checkOutStr.String)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse check_out_time: %w", parseErr)
			}
			entity.CheckOutTime = parsedTime
		}
		results = append(results, entity)
	}
	return results, nil
}

// ListByMemberID retrieves all attendance records for a given member, ordered by check-in time descending.
// PRE: memberID is non-empty
// POST: Returns records for the given member
func (s *SQLiteStore) ListByMemberID(ctx context.Context, memberID string) ([]domain.Attendance, error) {
	query := "SELECT id, check_in_time, check_out_time, member_id, schedule_id, class_date, mat_hours FROM attendance WHERE member_id = ? ORDER BY check_in_time DESC"

	rows, err := s.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Attendance
	for rows.Next() {
		var entity domain.Attendance
		var checkInStr string
		var checkOutStr, scheduleID, classDate sql.NullString
		if err := rows.Scan(
			&entity.ID,
			&checkInStr,
			&checkOutStr,
			&entity.MemberID,
			&scheduleID,
			&classDate,
			&entity.MatHours,
		); err != nil {
			return nil, err
		}
		if scheduleID.Valid {
			entity.ScheduleID = scheduleID.String
		}
		if classDate.Valid {
			entity.ClassDate = classDate.String
		}
		entity.CheckInTime, err = parseStoredTime(checkInStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse check_in_time: %w", err)
		}
		if checkOutStr.Valid {
			parsedTime, parseErr := parseStoredTime(checkOutStr.String)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse check_out_time: %w", parseErr)
			}
			entity.CheckOutTime = parsedTime
		}
		results = append(results, entity)
	}
	return results, rows.Err()
}

// ListByMemberIDAndDate retrieves attendance records for a member on a specific date.
// PRE: memberID is non-empty, date is YYYY-MM-DD format
// POST: Returns records matching memberID and date, ordered by check-in time desc
func (s *SQLiteStore) ListByMemberIDAndDate(ctx context.Context, memberID string, date string) ([]domain.Attendance, error) {
	query := `SELECT id, check_in_time, check_out_time, member_id, schedule_id, class_date, mat_hours
		FROM attendance
		WHERE member_id = ? AND SUBSTR(check_in_time, 1, 10) = ?
		ORDER BY check_in_time DESC`

	rows, err := s.db.QueryContext(ctx, query, memberID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Attendance
	for rows.Next() {
		var entity domain.Attendance
		var checkInStr string
		var checkOutStr, scheduleID, classDate sql.NullString
		if err := rows.Scan(
			&entity.ID,
			&checkInStr,
			&checkOutStr,
			&entity.MemberID,
			&scheduleID,
			&classDate,
			&entity.MatHours,
		); err != nil {
			return nil, err
		}
		if scheduleID.Valid {
			entity.ScheduleID = scheduleID.String
		}
		if classDate.Valid {
			entity.ClassDate = classDate.String
		}
		entity.CheckInTime, err = parseStoredTime(checkInStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse check_in_time: %w", err)
		}
		if checkOutStr.Valid {
			parsedTime, parseErr := parseStoredTime(checkOutStr.String)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse check_out_time: %w", parseErr)
			}
			entity.CheckOutTime = parsedTime
		}
		results = append(results, entity)
	}
	return results, rows.Err()
}

// ListDistinctMemberIDsByScheduleAndDate returns distinct member IDs who attended a specific class session.
// PRE: scheduleID and classDate are non-empty
// POST: Returns distinct member IDs for the given session
func (s *SQLiteStore) ListDistinctMemberIDsByScheduleAndDate(ctx context.Context, scheduleID string, classDate string) ([]string, error) {
	query := `SELECT DISTINCT member_id FROM attendance WHERE schedule_id = ? AND class_date = ?`
	rows, err := s.db.QueryContext(ctx, query, scheduleID, classDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ListDistinctMemberIDsByScheduleIDsSince returns distinct member IDs who attended any of the given schedules since a date.
// PRE: scheduleIDs is non-empty, since is YYYY-MM-DD format
// POST: Returns distinct member IDs matching the criteria
func (s *SQLiteStore) ListDistinctMemberIDsByScheduleIDsSince(ctx context.Context, scheduleIDs []string, since string) ([]string, error) {
	if len(scheduleIDs) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(scheduleIDs))
	args := make([]any, len(scheduleIDs)+1)
	for i, id := range scheduleIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	args[len(scheduleIDs)] = since

	query := fmt.Sprintf(
		`SELECT DISTINCT member_id FROM attendance WHERE schedule_id IN (%s) AND class_date >= ?`,
		strings.Join(placeholders, ","),
	)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ListByMemberIDAndDateRange retrieves attendance records for a member within a date range.
// PRE: memberID is non-empty, startDate and endDate are YYYY-MM-DD format
// POST: Returns records where check_in_time falls within the range (inclusive)
func (s *SQLiteStore) ListByMemberIDAndDateRange(ctx context.Context, memberID string, startDate string, endDate string) ([]domain.Attendance, error) {
	query := `SELECT id, check_in_time, check_out_time, member_id, schedule_id, class_date, mat_hours
		FROM attendance
		WHERE member_id = ? AND SUBSTR(check_in_time, 1, 10) >= ? AND SUBSTR(check_in_time, 1, 10) <= ?
		ORDER BY check_in_time DESC`

	rows, err := s.db.QueryContext(ctx, query, memberID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Attendance
	for rows.Next() {
		var entity domain.Attendance
		var checkInStr string
		var checkOutStr, scheduleID, classDate sql.NullString
		if err := rows.Scan(
			&entity.ID,
			&checkInStr,
			&checkOutStr,
			&entity.MemberID,
			&scheduleID,
			&classDate,
			&entity.MatHours,
		); err != nil {
			return nil, err
		}
		if scheduleID.Valid {
			entity.ScheduleID = scheduleID.String
		}
		if classDate.Valid {
			entity.ClassDate = classDate.String
		}
		entity.CheckInTime, err = parseStoredTime(checkInStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse check_in_time: %w", err)
		}
		if checkOutStr.Valid {
			parsedTime, parseErr := parseStoredTime(checkOutStr.String)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse check_out_time: %w", parseErr)
			}
			entity.CheckOutTime = parsedTime
		}
		results = append(results, entity)
	}
	return results, rows.Err()
}

// DeleteByMemberIDAndDateRange deletes attendance records for a member within a date range.
// PRE: memberID is non-empty, startDate and endDate are YYYY-MM-DD format
// POST: Returns the number of deleted records
func (s *SQLiteStore) DeleteByMemberIDAndDateRange(ctx context.Context, memberID string, startDate string, endDate string) (int, error) {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM attendance WHERE member_id = ? AND SUBSTR(check_in_time, 1, 10) >= ? AND SUBSTR(check_in_time, 1, 10) <= ?`,
		memberID, startDate, endDate)
	if err != nil {
		return 0, err
	}
	n, err := result.RowsAffected()
	return int(n), err
}

// SumMatHoursByMemberID returns the total mat hours for a member.
// Uses checkout-based duration where available, else defaults to 1.5h per session.
// PRE: memberID is non-empty
// POST: Returns total hours (>=0)
func (s *SQLiteStore) SumMatHoursByMemberID(ctx context.Context, memberID string) (float64, error) {
	var total sql.NullFloat64
	err := s.db.QueryRowContext(ctx,
		`SELECT SUM(
			CASE
				WHEN check_out_time IS NOT NULL AND check_out_time != ''
				THEN (julianday(check_out_time) - julianday(check_in_time)) * 24.0
				ELSE 1.5
			END
		) FROM attendance WHERE member_id = ?`, memberID).Scan(&total)
	if err != nil {
		return 0, err
	}
	if !total.Valid {
		return 0, nil
	}
	return total.Float64, nil
}

func parseStoredTime(value string) (time.Time, error) {
	if idx := strings.Index(value, " m="); idx != -1 {
		value = value[:idx]
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.9999999-07:00",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %q", value)
}
