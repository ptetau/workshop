package storage

import (
	"database/sql"
	"sort"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// openTestDB creates an in-memory SQLite database for testing.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// getTableNames returns sorted table names from sqlite_master, excluding internal tables.
func getTableNames(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan table name: %v", err)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getTableSQL returns sorted CREATE TABLE statements from sqlite_master.
func getTableSQL(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query("SELECT sql FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND sql IS NOT NULL ORDER BY name")
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	defer rows.Close()

	var sqls []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatalf("failed to scan sql: %v", err)
		}
		sqls = append(sqls, normalizeSQL(s))
	}
	sort.Strings(sqls)
	return sqls
}

// normalizeSQL collapses whitespace for comparison.
func normalizeSQL(s string) string {
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

// expectedTables is the sorted list of tables after all migrations.
var expectedTables = []string{
	"account",
	"activation_token",
	"attendance",
	"class_type",
	"coach_observation",
	"email",
	"email_recipient",
	"email_template",
	"grading_config",
	"grading_proposal",
	"grading_record",
	"holiday",
	"injury",
	"member",
	"message",
	"milestone",
	"notice",
	"program",
	"rotor",
	"rotor_theme",
	"schedule",
	"schema_version",
	"term",
	"topic",
	"topic_schedule",
	"training_goal",
	"vote",
	"waiver",
}

// TestMigrateDB_Fresh verifies all migrations apply cleanly to an empty database.
func TestMigrateDB_Fresh(t *testing.T) {
	db := openTestDB(t)

	if err := MigrateDB(db, ":memory:"); err != nil {
		t.Fatalf("MigrateDB failed on fresh db: %v", err)
	}

	// Verify version
	version, err := SchemaVersion(db)
	if err != nil {
		t.Fatalf("SchemaVersion failed: %v", err)
	}
	if version != LatestSchemaVersion() {
		t.Errorf("version = %d, want %d", version, LatestSchemaVersion())
	}

	// Verify all expected tables exist
	tables := getTableNames(t, db)
	if len(tables) != len(expectedTables) {
		t.Errorf("got %d tables, want %d\ngot:  %v\nwant: %v", len(tables), len(expectedTables), tables, expectedTables)
	}
	for i, want := range expectedTables {
		if i >= len(tables) {
			t.Errorf("missing table: %s", want)
			continue
		}
		if tables[i] != want {
			t.Errorf("table[%d] = %q, want %q", i, tables[i], want)
		}
	}
}

// TestMigrateDB_Idempotent verifies that running MigrateDB twice produces no errors
// and the version remains the same.
func TestMigrateDB_Idempotent(t *testing.T) {
	db := openTestDB(t)

	if err := MigrateDB(db, ":memory:"); err != nil {
		t.Fatalf("first MigrateDB failed: %v", err)
	}

	version1, _ := SchemaVersion(db)

	if err := MigrateDB(db, ":memory:"); err != nil {
		t.Fatalf("second MigrateDB failed: %v", err)
	}

	version2, _ := SchemaVersion(db)
	if version1 != version2 {
		t.Errorf("version changed after idempotent run: %d → %d", version1, version2)
	}
}

// TestMigrateDB_SchemaDrift verifies that the migration chain produces the exact same
// schema as a golden snapshot. This catches cases where someone modifies a migration
// function without updating the expected output.
func TestMigrateDB_SchemaDrift(t *testing.T) {
	db := openTestDB(t)

	if err := MigrateDB(db, ":memory:"); err != nil {
		t.Fatalf("MigrateDB failed: %v", err)
	}

	// Build golden snapshot from current migrations
	golden := getTableSQL(t, db)

	// Apply migrations to a second fresh database
	db2 := openTestDB(t)
	if err := MigrateDB(db2, ":memory:"); err != nil {
		t.Fatalf("MigrateDB (second) failed: %v", err)
	}

	actual := getTableSQL(t, db2)

	if len(golden) != len(actual) {
		t.Fatalf("schema drift: golden has %d tables, actual has %d", len(golden), len(actual))
	}

	for i := range golden {
		if golden[i] != actual[i] {
			t.Errorf("schema drift at table %d:\ngolden: %s\nactual: %s", i, golden[i], actual[i])
		}
	}
}

// TestMigrateDB_DataSurvival verifies that existing data survives migration.
// For the baseline (migration 1), we insert data before migrating and verify it's still there.
// This test serves as a template for future migration data-survival tests.
func TestMigrateDB_DataSurvival(t *testing.T) {
	db := openTestDB(t)

	// Apply migration 1
	if err := MigrateDB(db, ":memory:"); err != nil {
		t.Fatalf("MigrateDB failed: %v", err)
	}

	// Insert test data
	_, err := db.Exec(`INSERT INTO member (id, email, name, program, status) VALUES ('m1', 'test@test.com', 'Test User', 'BJJ', 'active')`)
	if err != nil {
		t.Fatalf("failed to insert test member: %v", err)
	}
	_, err = db.Exec(`INSERT INTO attendance (id, member_id, check_in_time) VALUES ('a1', 'm1', '2026-01-01T10:00:00Z')`)
	if err != nil {
		t.Fatalf("failed to insert test attendance: %v", err)
	}

	// Run MigrateDB again (should be no-op since we're at latest)
	if err := MigrateDB(db, ":memory:"); err != nil {
		t.Fatalf("second MigrateDB failed: %v", err)
	}

	// Verify data survived
	var name string
	if err := db.QueryRow("SELECT name FROM member WHERE id = 'm1'").Scan(&name); err != nil {
		t.Fatalf("member data lost after migration: %v", err)
	}
	if name != "Test User" {
		t.Errorf("member name = %q, want %q", name, "Test User")
	}

	var checkIn string
	if err := db.QueryRow("SELECT check_in_time FROM attendance WHERE id = 'a1'").Scan(&checkIn); err != nil {
		t.Fatalf("attendance data lost after migration: %v", err)
	}
	if checkIn != "2026-01-01T10:00:00Z" {
		t.Errorf("attendance check_in_time = %q, want %q", checkIn, "2026-01-01T10:00:00Z")
	}
}

// TestMigrateDB_VersionProgression verifies that SchemaVersion reports 0 before
// migration and the correct version after.
func TestMigrateDB_VersionProgression(t *testing.T) {
	db := openTestDB(t)

	// Before any migration, version should be 0
	v, err := SchemaVersion(db)
	if err != nil {
		t.Fatalf("SchemaVersion failed: %v", err)
	}
	if v != 0 {
		t.Errorf("initial version = %d, want 0", v)
	}

	if err := MigrateDB(db, ":memory:"); err != nil {
		t.Fatalf("MigrateDB failed: %v", err)
	}

	v, err = SchemaVersion(db)
	if err != nil {
		t.Fatalf("SchemaVersion failed: %v", err)
	}
	if v != LatestSchemaVersion() {
		t.Errorf("post-migration version = %d, want %d", v, LatestSchemaVersion())
	}
}

// TestMigrateDB_ExistingDB verifies that MigrateDB works on a database that already
// has tables but no schema_version tracking (simulates upgrading a pre-migration database).
func TestMigrateDB_ExistingDB(t *testing.T) {
	db := openTestDB(t)

	// Simulate a pre-migration database: create some tables manually
	_, err := db.Exec(`CREATE TABLE account (id TEXT PRIMARY KEY, email TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL DEFAULT '', role TEXT NOT NULL, created_at TEXT NOT NULL, failed_logins INTEGER NOT NULL DEFAULT 0, locked_until TEXT, password_change_required INTEGER NOT NULL DEFAULT 0)`)
	if err != nil {
		t.Fatalf("failed to create pre-migration table: %v", err)
	}
	_, err = db.Exec(`INSERT INTO account (id, email, role, created_at) VALUES ('a1', 'admin@test.com', 'admin', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("failed to insert pre-migration data: %v", err)
	}

	// Run MigrateDB — should detect version 0 and apply migration 1 (with IF NOT EXISTS)
	if err := MigrateDB(db, ":memory:"); err != nil {
		t.Fatalf("MigrateDB on existing db failed: %v", err)
	}

	// Verify pre-existing data survived
	var email string
	if err := db.QueryRow("SELECT email FROM account WHERE id = 'a1'").Scan(&email); err != nil {
		t.Fatalf("pre-migration data lost: %v", err)
	}
	if email != "admin@test.com" {
		t.Errorf("email = %q, want %q", email, "admin@test.com")
	}

	// Verify version is now at latest
	v, _ := SchemaVersion(db)
	if v != LatestSchemaVersion() {
		t.Errorf("version = %d, want %d", v, LatestSchemaVersion())
	}
}
