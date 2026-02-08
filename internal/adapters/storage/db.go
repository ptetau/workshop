package storage

import (
	"database/sql"
	"fmt"
)

// InitDB initializes the database schema.
// PRE: db is a valid database connection
// POST: All tables are created, WAL mode enabled
func InitDB(db *sql.DB) error {
	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}
	// Enable foreign key enforcement
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS account (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL DEFAULT '',
		role TEXT NOT NULL,
		created_at TEXT NOT NULL,
		failed_logins INTEGER NOT NULL DEFAULT 0,
		locked_until TEXT,
		password_change_required INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS member (
		id TEXT PRIMARY KEY,
		account_id TEXT,
		email TEXT NOT NULL UNIQUE,
		fee INTEGER,
		frequency TEXT,
		name TEXT NOT NULL,
		program TEXT NOT NULL,
		status TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS program (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS class_type (
		id TEXT PRIMARY KEY,
		program_id TEXT NOT NULL,
		name TEXT NOT NULL,
		FOREIGN KEY (program_id) REFERENCES program(id)
	);

	CREATE TABLE IF NOT EXISTS schedule (
		id TEXT PRIMARY KEY,
		class_type_id TEXT NOT NULL,
		day TEXT NOT NULL,
		start_time TEXT NOT NULL,
		end_time TEXT NOT NULL,
		FOREIGN KEY (class_type_id) REFERENCES class_type(id)
	);

	CREATE TABLE IF NOT EXISTS term (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS holiday (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS waiver (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		accepted_terms INTEGER NOT NULL,
		ip_address TEXT,
		signed_at TEXT NOT NULL,
		FOREIGN KEY (member_id) REFERENCES member(id)
	);

	CREATE TABLE IF NOT EXISTS injury (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		body_part TEXT NOT NULL,
		description TEXT,
		reported_at TEXT NOT NULL,
		FOREIGN KEY (member_id) REFERENCES member(id)
	);

	CREATE TABLE IF NOT EXISTS attendance (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		check_in_time TEXT NOT NULL,
		check_out_time TEXT,
		schedule_id TEXT,
		class_date TEXT,
		FOREIGN KEY (member_id) REFERENCES member(id)
	);

	CREATE TABLE IF NOT EXISTS notice (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_by TEXT NOT NULL,
		published_by TEXT,
		target_id TEXT,
		author_name TEXT NOT NULL DEFAULT '',
		show_author INTEGER NOT NULL DEFAULT 0,
		color TEXT NOT NULL DEFAULT 'orange',
		pinned INTEGER NOT NULL DEFAULT 0,
		pinned_at TEXT,
		visible_from TEXT,
		visible_until TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT,
		published_at TEXT
	);

	CREATE TABLE IF NOT EXISTS grading_record (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		belt TEXT NOT NULL,
		stripe INTEGER NOT NULL DEFAULT 0,
		promoted_at TEXT NOT NULL,
		proposed_by TEXT,
		approved_by TEXT,
		method TEXT NOT NULL DEFAULT 'standard',
		FOREIGN KEY (member_id) REFERENCES member(id)
	);

	CREATE TABLE IF NOT EXISTS grading_config (
		id TEXT PRIMARY KEY,
		program TEXT NOT NULL,
		belt TEXT NOT NULL,
		flight_time_hours REAL NOT NULL DEFAULT 0,
		attendance_pct REAL NOT NULL DEFAULT 0,
		stripe_count INTEGER NOT NULL DEFAULT 4
	);

	CREATE TABLE IF NOT EXISTS grading_proposal (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		target_belt TEXT NOT NULL,
		notes TEXT,
		proposed_by TEXT NOT NULL,
		approved_by TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at TEXT NOT NULL,
		decided_at TEXT,
		FOREIGN KEY (member_id) REFERENCES member(id)
	);

	CREATE TABLE IF NOT EXISTS message (
		id TEXT PRIMARY KEY,
		sender_id TEXT NOT NULL,
		receiver_id TEXT NOT NULL,
		subject TEXT,
		content TEXT NOT NULL,
		read_at TEXT,
		created_at TEXT NOT NULL,
		FOREIGN KEY (receiver_id) REFERENCES member(id)
	);

	CREATE TABLE IF NOT EXISTS coach_observation (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		author_id TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT,
		FOREIGN KEY (member_id) REFERENCES member(id)
	);

	CREATE TABLE IF NOT EXISTS milestone (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		metric TEXT NOT NULL,
		threshold REAL NOT NULL,
		badge_icon TEXT
	);

	CREATE TABLE IF NOT EXISTS training_goal (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		target INTEGER NOT NULL,
		period TEXT NOT NULL,
		created_at TEXT NOT NULL,
		active INTEGER NOT NULL DEFAULT 1,
		FOREIGN KEY (member_id) REFERENCES member(id)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}
