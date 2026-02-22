package storage

import (
	"database/sql"
	"fmt"
	"log/slog"
)

// migration is a numbered schema change function.
// PRE: tx is a valid transaction
// POST: schema changes applied within the transaction
type migration struct {
	version     int
	description string
	apply       func(tx *sql.Tx) error
}

// migrations is the ordered list of all schema migrations.
// New migrations are appended here — never modify existing ones.
var migrations = []migration{
	{version: 1, description: "baseline schema", apply: migrate1},
	{version: 2, description: "email system tables", apply: migrate2},
	{version: 3, description: "account activation", apply: migrate3},
	{version: 4, description: "curriculum rotor system", apply: migrate4},
	{version: 5, description: "attendance mat hours", apply: migrate5},
	{version: 6, description: "member milestone achievements", apply: migrate6},
	{version: 7, description: "estimated hours", apply: migrate7},
	{version: 8, description: "member grading metric", apply: migrate8},
	{version: 9, description: "grading notes", apply: migrate9},
	{version: 10, description: "grading member config overrides", apply: migrate10},
	{version: 11, description: "self-estimate review fields", apply: migrate11},
	{version: 12, description: "hidden surprise themes", apply: migrate12},
	{version: 13, description: "calendar events", apply: migrate13},
	{version: 14, description: "calendar event indexes", apply: migrate14},
	{version: 15, description: "feature flags and beta cohort", apply: migrate15},
	{version: 16, description: "class type metadata", apply: migrate16},
	{version: 17, description: "bugbox submissions", apply: migrate17},
	{version: 18, description: "competition interest tracking", apply: migrate18},
	{version: 19, description: "personal goals", apply: migrate19},
	{version: 20, description: "personal goal type", apply: migrate20},
	{version: 21, description: "outbox for external integrations", apply: migrate21},
	{version: 22, description: "attendance and member performance indexes", apply: migrate22},
}

// SchemaVersion returns the current schema version of the database.
// PRE: db is a valid database connection
// POST: returns the version number (0 if no schema_version table exists)
func SchemaVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT version FROM schema_version LIMIT 1").Scan(&version)
	if err != nil {
		return 0, nil
	}
	return version, nil
}

// LatestSchemaVersion returns the highest migration version available.
func LatestSchemaVersion() int {
	if len(migrations) == 0 {
		return 0
	}
	return migrations[len(migrations)-1].version
}

// MigrateDB applies all pending migrations to bring the database to the latest schema version.
// PRE: db is a valid database connection
// POST: All pending migrations applied, schema_version updated
// INVARIANT: Each migration runs in its own transaction — failure rolls back cleanly
func MigrateDB(db *sql.DB, dbPath string) error {
	// Ensure schema_version table exists
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`); err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}

	// Seed version 0 if table is empty
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&count); err != nil {
		return fmt.Errorf("failed to count schema_version rows: %w", err)
	}
	if count == 0 {
		if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (0)"); err != nil {
			return fmt.Errorf("failed to seed schema_version: %w", err)
		}
	}

	// Read current version
	current, err := SchemaVersion(db)
	if err != nil {
		return fmt.Errorf("failed to read schema version: %w", err)
	}

	// Find pending migrations
	var pending []migration
	for _, m := range migrations {
		if m.version > current {
			pending = append(pending, m)
		}
	}

	if len(pending) == 0 {
		slog.Info("schema_up_to_date", "version", current)
		return nil
	}

	// Backup before migrating (only for file-based databases, not :memory:)
	if dbPath != "" && dbPath != ":memory:" {
		backupPath := dbPath + fmt.Sprintf(".bak-v%d", current)
		slog.Info("schema_backup", "from_version", current, "backup", backupPath)
		if _, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath)); err != nil {
			return fmt.Errorf("failed to backup database before migration: %w", err)
		}
	}

	// Apply each pending migration in its own transaction
	for _, m := range pending {
		slog.Info("schema_migrate", "version", m.version, "description", m.description)

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("migration %d: failed to begin transaction: %w", m.version, err)
		}

		if err := m.apply(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d (%s): %w", m.version, m.description, err)
		}

		if _, err := tx.Exec("UPDATE schema_version SET version = ?", m.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d: failed to update version: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("migration %d: failed to commit: %w", m.version, err)
		}

		slog.Info("schema_migrated", "version", m.version)
	}

	slog.Info("schema_migration_complete", "from", current, "to", migrations[len(migrations)-1].version)
	return nil
}

// --- Migration 1: Baseline schema ---
// This is the initial schema. All tables use CREATE TABLE IF NOT EXISTS
// so it is safe to run on both new and existing databases.
func migrate1(tx *sql.Tx) error {
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

	_, err := tx.Exec(schema)
	return err
}

// --- Migration 2: Email system tables ---
// Adds email, email_recipient, and email_template tables for §8.2 Email System.
func migrate2(tx *sql.Tx) error {
	schema := `
	CREATE TABLE IF NOT EXISTS email (
		id TEXT PRIMARY KEY,
		subject TEXT NOT NULL,
		body TEXT NOT NULL,
		sender_id TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'draft',
		scheduled_at TEXT,
		sent_at TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT,
		resend_message_id TEXT,
		template_version_id TEXT,
		FOREIGN KEY (sender_id) REFERENCES account(id)
	);

	CREATE TABLE IF NOT EXISTS email_recipient (
		email_id TEXT NOT NULL,
		member_id TEXT NOT NULL,
		member_name TEXT NOT NULL DEFAULT '',
		member_email TEXT NOT NULL DEFAULT '',
		delivery_status TEXT NOT NULL DEFAULT '',
		PRIMARY KEY (email_id, member_id),
		FOREIGN KEY (email_id) REFERENCES email(id) ON DELETE CASCADE,
		FOREIGN KEY (member_id) REFERENCES member(id)
	);

	CREATE TABLE IF NOT EXISTS email_template (
		id TEXT PRIMARY KEY,
		header TEXT NOT NULL DEFAULT '',
		footer TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		active INTEGER NOT NULL DEFAULT 1
	);

	CREATE INDEX IF NOT EXISTS idx_email_status ON email(status);
	CREATE INDEX IF NOT EXISTS idx_email_sender ON email(sender_id);
	CREATE INDEX IF NOT EXISTS idx_email_recipient_member ON email_recipient(member_id);
	`
	_, err := tx.Exec(schema)
	return err
}

// --- Migration 3: Account activation ---
// Adds activation_token table and status column to account for §8.2.6 Account Activation.
func migrate3(tx *sql.Tx) error {
	schema := `
	ALTER TABLE account ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

	CREATE TABLE IF NOT EXISTS activation_token (
		id TEXT PRIMARY KEY,
		account_id TEXT NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at TEXT NOT NULL,
		used INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_activation_token_account ON activation_token(account_id);
	CREATE INDEX IF NOT EXISTS idx_activation_token_token ON activation_token(token);
	`
	_, err := tx.Exec(schema)
	return err
}

// --- Migration 4: Curriculum rotor system ---
// Adds rotor, rotor_theme, topic, topic_schedule, and vote tables for §5/§6 Curriculum & Voting.
func migrate4(tx *sql.Tx) error {
	schema := `
	CREATE TABLE IF NOT EXISTS rotor (
		id TEXT PRIMARY KEY,
		class_type_id TEXT NOT NULL,
		name TEXT NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		status TEXT NOT NULL DEFAULT 'draft',
		preview_on INTEGER NOT NULL DEFAULT 0,
		created_by TEXT NOT NULL,
		created_at TEXT NOT NULL,
		activated_at TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (class_type_id) REFERENCES class_type(id) ON DELETE CASCADE,
		FOREIGN KEY (created_by) REFERENCES account(id)
	);

	CREATE INDEX IF NOT EXISTS idx_rotor_class_type ON rotor(class_type_id);
	CREATE INDEX IF NOT EXISTS idx_rotor_status ON rotor(status);

	CREATE TABLE IF NOT EXISTS rotor_theme (
		id TEXT PRIMARY KEY,
		rotor_id TEXT NOT NULL,
		name TEXT NOT NULL,
		position INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (rotor_id) REFERENCES rotor(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_rotor_theme_rotor ON rotor_theme(rotor_id);

	CREATE TABLE IF NOT EXISTS topic (
		id TEXT PRIMARY KEY,
		rotor_theme_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		duration_weeks INTEGER NOT NULL DEFAULT 1,
		position INTEGER NOT NULL DEFAULT 0,
		last_covered TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (rotor_theme_id) REFERENCES rotor_theme(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_topic_theme ON topic(rotor_theme_id);

	CREATE TABLE IF NOT EXISTS topic_schedule (
		id TEXT PRIMARY KEY,
		topic_id TEXT NOT NULL,
		rotor_theme_id TEXT NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'scheduled',
		FOREIGN KEY (topic_id) REFERENCES topic(id) ON DELETE CASCADE,
		FOREIGN KEY (rotor_theme_id) REFERENCES rotor_theme(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_topic_schedule_theme ON topic_schedule(rotor_theme_id);
	CREATE INDEX IF NOT EXISTS idx_topic_schedule_status ON topic_schedule(status);

	CREATE TABLE IF NOT EXISTS vote (
		id TEXT PRIMARY KEY,
		topic_id TEXT NOT NULL,
		account_id TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY (topic_id) REFERENCES topic(id) ON DELETE CASCADE,
		FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
		UNIQUE(topic_id, account_id)
	);

	CREATE INDEX IF NOT EXISTS idx_vote_topic ON vote(topic_id);
	`
	_, err := tx.Exec(schema)
	return err
}

// migrate5 adds mat_hours column to attendance table.
func migrate5(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE attendance ADD COLUMN mat_hours REAL NOT NULL DEFAULT 0`)
	return err
}

// --- Migration 6: Member milestone achievements ---
// Tracks which milestones each member has earned.
func migrate6(tx *sql.Tx) error {
	schema := `
	CREATE TABLE IF NOT EXISTS member_milestone (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		milestone_id TEXT NOT NULL,
		earned_at TEXT NOT NULL,
		notified INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (member_id) REFERENCES member(id),
		FOREIGN KEY (milestone_id) REFERENCES milestone(id),
		UNIQUE(member_id, milestone_id)
	);

	CREATE INDEX IF NOT EXISTS idx_member_milestone_member ON member_milestone(member_id);
	`
	_, err := tx.Exec(schema)
	return err
}

// --- Migration 7: Estimated hours ---
// Bulk-estimated mat hours for members with incomplete check-in records.
func migrate7(tx *sql.Tx) error {
	schema := `
	CREATE TABLE IF NOT EXISTS estimated_hours (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		weekly_hours REAL NOT NULL,
		total_hours REAL NOT NULL,
		source TEXT NOT NULL DEFAULT 'estimate',
		status TEXT NOT NULL DEFAULT 'approved',
		note TEXT NOT NULL DEFAULT '',
		created_by TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY (member_id) REFERENCES member(id)
	);

	CREATE INDEX IF NOT EXISTS idx_estimated_hours_member ON estimated_hours(member_id);
	`
	_, err := tx.Exec(schema)
	return err
}

// --- Migration 9: Grading notes ---
// Adds grading_note table for coach/admin notes on member readiness.
func migrate9(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS grading_note (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		content TEXT NOT NULL,
		created_by TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY (member_id) REFERENCES member(id)
	);
	CREATE INDEX IF NOT EXISTS idx_grading_note_member ON grading_note(member_id);
	`)
	return err
}

// --- Migration 16: Class type metadata ---
// Adds optional fields to class_type for richer timetable rendering and filtering.
func migrate16(tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE class_type ADD COLUMN description TEXT NOT NULL DEFAULT '';
	ALTER TABLE class_type ADD COLUMN attire TEXT NOT NULL DEFAULT '';
	ALTER TABLE class_type ADD COLUMN level TEXT NOT NULL DEFAULT '';
	`)
	return err
}

// --- Migration 15: Feature flags and beta cohort ---
// Adds:
// - account.beta_tester: marks an account as part of a test cohort
// - feature_flag: server-enforced feature availability by role, with beta override
func migrate15(tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE account ADD COLUMN beta_tester INTEGER NOT NULL DEFAULT 0;

	CREATE TABLE IF NOT EXISTS feature_flag (
		key TEXT PRIMARY KEY,
		description TEXT NOT NULL,
		enabled_admin INTEGER NOT NULL DEFAULT 1,
		enabled_coach INTEGER NOT NULL DEFAULT 1,
		enabled_member INTEGER NOT NULL DEFAULT 0,
		enabled_trial INTEGER NOT NULL DEFAULT 0,
		beta_override INTEGER NOT NULL DEFAULT 0
	);
	`)
	return err
}

// --- Migration 10: Grading member config overrides ---
// Adds per-member threshold overrides for grading eligibility.
func migrate10(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS grading_member_config (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		belt TEXT NOT NULL,
		flight_time_hours REAL NOT NULL DEFAULT 0,
		attendance_pct REAL NOT NULL DEFAULT 0,
		FOREIGN KEY (member_id) REFERENCES member(id),
		UNIQUE(member_id, belt)
	);
	CREATE INDEX IF NOT EXISTS idx_grading_member_config_member ON grading_member_config(member_id);
	`)
	return err
}

// --- Migration 11: Self-estimate review fields ---
// Adds review columns for the self-estimate approval workflow (§3.5).
func migrate11(tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE estimated_hours ADD COLUMN reviewed_by TEXT NOT NULL DEFAULT '';
	ALTER TABLE estimated_hours ADD COLUMN reviewed_at TEXT NOT NULL DEFAULT '';
	ALTER TABLE estimated_hours ADD COLUMN review_note TEXT NOT NULL DEFAULT '';
	CREATE INDEX IF NOT EXISTS idx_estimated_hours_status ON estimated_hours(status);
	`)
	return err
}

// --- Migration 12: Hidden surprise themes ---
// Adds hidden column to rotor_theme so coaches can mark themes as surprises.
func migrate12(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE rotor_theme ADD COLUMN hidden INTEGER NOT NULL DEFAULT 0`)
	return err
}

// --- Migration 13: Calendar events ---
// Creates calendar_events table for club events and competitions.
func migrate13(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS calendar_event (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'event',
		description TEXT NOT NULL DEFAULT '',
		location TEXT NOT NULL DEFAULT '',
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL DEFAULT '',
		registration_url TEXT NOT NULL DEFAULT '',
		created_by TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

// --- Migration 14: Calendar event indexes ---
// Backfills end_date for existing single-day events and adds an index to speed up
// overlap queries (start_date <= to AND end_date >= from).
func migrate14(tx *sql.Tx) error {
	_, err := tx.Exec(`
	UPDATE calendar_event
	SET end_date = start_date
	WHERE end_date = '';

	CREATE INDEX IF NOT EXISTS idx_calendar_event_date_range
	ON calendar_event(start_date, end_date);
	`)
	return err
}

// --- Migration 17: BugBox submissions ---
// Creates bugbox_submission table for in-app bug reports (§15.3).
func migrate17(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS bugbox_submission (
		id TEXT PRIMARY KEY,
		summary TEXT NOT NULL,
		description TEXT NOT NULL,
		steps TEXT NOT NULL DEFAULT '',
		expected TEXT NOT NULL DEFAULT '',
		actual TEXT NOT NULL DEFAULT '',
		route TEXT NOT NULL DEFAULT '',
		user_agent TEXT NOT NULL DEFAULT '',
		viewport TEXT NOT NULL DEFAULT '',
		role TEXT NOT NULL DEFAULT '',
		impersonated_role TEXT NOT NULL DEFAULT '',
		submitted_at TEXT NOT NULL,
		screenshot_path TEXT NOT NULL DEFAULT '',
		github_issue_number INTEGER NOT NULL DEFAULT 0,
		github_issue_url TEXT NOT NULL DEFAULT ''
	)`)
	return err
}

// --- Migration 18: Competition interest tracking ---
// Tracks which members are interested in attending competitions.
func migrate18(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS competition_interest (
		id TEXT PRIMARY KEY,
		event_id TEXT NOT NULL,
		member_id TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (event_id) REFERENCES calendar_event(id) ON DELETE CASCADE,
		FOREIGN KEY (member_id) REFERENCES member(id) ON DELETE CASCADE,
		UNIQUE(event_id, member_id)
	);

	CREATE INDEX IF NOT EXISTS idx_competition_interest_event ON competition_interest(event_id);
	CREATE INDEX IF NOT EXISTS idx_competition_interest_member ON competition_interest(member_id);
	`)
	return err
}

// --- Migration 8: Member grading metric ---
// Adds grading_metric column so kids can be toggled between sessions and hours mode.
func migrate8(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE member ADD COLUMN grading_metric TEXT NOT NULL DEFAULT 'sessions'`)
	return err
}

// --- Migration 19: Personal goals ---
// Creates personal_goal table for member-defined training targets with calendar display.
func migrate19(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS personal_goal (
		id TEXT PRIMARY KEY,
		member_id TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		target INTEGER NOT NULL,
		unit TEXT NOT NULL DEFAULT 'sessions',
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		color TEXT NOT NULL DEFAULT '#F9B232',
		progress INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY (member_id) REFERENCES member(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_personal_goal_member ON personal_goal(member_id);
	CREATE INDEX IF NOT EXISTS idx_personal_goal_dates ON personal_goal(start_date, end_date);
	`)
	return err
}

// --- Migration 20: Personal goal type ---
// Adds type column to personal_goal for manual vs auto-tracked hours goals.
func migrate20(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE personal_goal ADD COLUMN type TEXT NOT NULL DEFAULT 'manual'`)
	return err
}

// --- Migration 21: Outbox for external integrations ---
// Creates outbox table for reliable external integration retries.
func migrate21(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS outbox (
		id TEXT PRIMARY KEY,
		action_type TEXT NOT NULL,
		payload TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		attempts INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 5,
		last_attempted_at TEXT,
		created_at TEXT NOT NULL,
		external_id TEXT NOT NULL DEFAULT '',
		error_message TEXT NOT NULL DEFAULT ''
	);

	CREATE INDEX IF NOT EXISTS idx_outbox_status ON outbox(status);
	CREATE INDEX IF NOT EXISTS idx_outbox_action_type ON outbox(action_type);
	CREATE INDEX IF NOT EXISTS idx_outbox_created_at ON outbox(created_at);
	`)
	return err
}

// --- Migration 22: Attendance and member performance indexes ---
// Adds indexes to speed up attendance lookups by member and date,
// and member queries by status (used by inactive radar).
func migrate22(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE INDEX IF NOT EXISTS idx_attendance_member ON attendance(member_id);
	CREATE INDEX IF NOT EXISTS idx_attendance_class_date ON attendance(class_date);
	CREATE INDEX IF NOT EXISTS idx_member_status ON member(status);
	`)
	return err
}
