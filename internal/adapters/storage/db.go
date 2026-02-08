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
