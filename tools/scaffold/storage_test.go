package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestScaffoldedStorageLayer(t *testing.T) {
	// 1. Setup temporary workspace
	// Use MkdirTemp + defer RemoveAll to avoid failing test on Windows cleanup locks
	workDir, err := os.MkdirTemp("", "scaffold_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	// 2. Scaffold a basic concept "Widget" with storage
	runInit([]string{
		"--root", workDir,
		"--module", "testapp",
		"--concept", "Widget",
		"--field", "Widget:Name:string",
		"--field", "Widget:Price:int",
	})

	// 3. Verify files exist
	expectedFiles := []string{
		"internal/adapters/storage/db.go",
		"internal/adapters/storage/widget/store.go",
		"internal/adapters/storage/widget/sqlite_store.go",
		"internal/adapters/storage/migrations",
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(workDir, f)); err != nil {
			t.Errorf("missing scaffolded file: %s", f)
		}
	}

	// 4. Verify DB Init Compilation & Runtime (via a generated test file)
	// Since we can't easily compile the full app in this test environment without
	// downloading deps, we'll write a main_test.go in the scaffolded dir
	// that imports the local packages and runs a test.

	// However, running `go test` inside the temp dir requires `go mod tidy`
	// and internet access, which might be flaky or slow.
	// Instead, we can verify the *generated logic* structurally or
	// by parsing the generated files.

	// BETTER approach for "Comprehensive":
	// We manually load the generated schema into an in-memory DB
	// and verify it matches expectations.

	// Use delete journal mode to avoid SHM/WAL files locking on Windows during cleanup
	dbPath := filepath.Join(workDir, "test.db?_journal_mode=DELETE")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// We must close DB to release lock for TempDir cleanup
	defer func() {
		_ = db.Close()
	}()

	// Find the migration file
	matches, _ := filepath.Glob(filepath.Join(workDir, "internal/adapters/storage/migrations/*.sql"))
	if len(matches) == 0 {
		t.Fatal("no migrations found")
	}

	migrationSQL, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("failed to apply migration: %v", err)
	}

	// Verify table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='widget'").Scan(&tableName)
	if err != nil {
		t.Fatalf("table 'widget' not created: %v", err)
	}

	// Verify columns
	rows, err := db.Query("PRAGMA table_info(widget)")
	if err != nil {
		t.Fatalf("failed to query table info: %v", err)
	}
	defer rows.Close()

	cols := map[string]string{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt interface{}
		rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk)
		cols[name] = ctype
	}

	if cols["name"] != "TEXT" {
		t.Errorf("expected Name column to be TEXT, got %s", cols["name"])
	}
	if cols["price"] != "INTEGER" {
		t.Errorf("expected Price column to be INTEGER, got %s", cols["price"])
	}

	// 5. Verify compliance (invoke lintguidelines)
	// We need to build lintguidelines first or run it via "go run"
	// To be robust, let's just create a dummy "lintguidelines" command or trust the integration test logic.
	// Actually, we can just shell out to "go run ../lintguidelines" assuming we are in tools/scaffold.
	// But paths are tricky. Let's assume the user has built the tools or we use the source.

	// We'll skip invoking the full linter binary to keep the test standalone/fast,
	// relying on the fact that TestScaffoldedAppPassesLint (in lintguidelines/integration_test.go)
	// already covers the linting aspect. This test focuses on the DB.
}

// TestConceptStorageParity verifies that all fields in a concept are correctly
// propagated to the storage layer (Store struct, SQL columns, INSERT/SCAN statements).
func TestConceptStorageParity(t *testing.T) {
	// Use MkdirTemp + defer RemoveAll to avoid failing test on Windows cleanup locks
	workDir, err := os.MkdirTemp("", "parity_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	// 1. Scaffold a complex concept
	runInit([]string{
		"--root", workDir,
		"--module", "parityapp",
		"--concept", "ComplexItem",
		"--field", "ComplexItem:Name:string",
		"--field", "ComplexItem:Count:int",
		"--field", "ComplexItem:IsActive:bool",
		"--field", "ComplexItem:Rating:float64",
	})

	// 2. Parse the generated sqlite_store.go to verify logic
	storeContent := readFile(t, filepath.Join(workDir, "internal/adapters/storage/complex_item/sqlite_store.go"))

	// Check 1: Scan fields
	// We expect Scan to look like:
	// &entity.ID,
	// &entity.Name,
	// &entity.Count,
	// &entity.IsActive,
	// &entity.Rating,
	expectedScans := []string{
		"&entity.ID",
		"&entity.Name",
		"&entity.Count",
		"&entity.IsActive",
		"&entity.Rating",
	}
	for _, s := range expectedScans {
		if !strings.Contains(storeContent, s) {
			t.Errorf("sqlite_store.go missing Scan for %s", s)
		}
	}

	// Check 2: Upsert fields list
	// We'll normalize whitespace to avoid fragility
	normalizedStore := strings.ReplaceAll(storeContent, " ", "")
	normalizedStore = strings.ReplaceAll(normalizedStore, "\t", "")
	// Fields are sorted by name in normalizeConcept.
	// Sort order: Count, ID, IsActive, Name, Rating
	// Snake case: count, id, is_active, name, rating
	expectedFieldsDef := `fields:=[]string{"id","count","is_active","name","rating",}`
	if !strings.Contains(normalizedStore, expectedFieldsDef) {
		t.Errorf("sqlite_store.go missing or incorrect fields definition.\nExpected (normalized): %s\nActual:\n%s", expectedFieldsDef, storeContent)
	}

	// Check 3: Placeholders count
	// placeholders := []string{ "?","?","?","?","?", }
	expectedPlaceholders := `placeholders:=[]string{"?","?","?","?","?",}`
	if !strings.Contains(normalizedStore, expectedPlaceholders) {
		t.Errorf("sqlite_store.go incorrect placeholders.\nExpected (normalized): %s", expectedPlaceholders)
	}

	// 3. Verify Migration Content (Parity with DB schema)
	migrations, _ := filepath.Glob(filepath.Join(workDir, "internal/adapters/storage/migrations/*_create_complex_item.sql"))
	if len(migrations) != 1 {
		t.Fatal("failed to find create migration for complex item")
	}
	migrationSQL := readFile(t, migrations[0])

	// Check column definitions
	checks := []struct {
		col  string
		kind string
	}{
		{"id", "TEXT"},
		{"name", "TEXT"},
		{"count", "INTEGER"},           // int maps to INTEGER
		{"is_active", "BOOLEAN"},       // bool maps to BOOLEAN
		{"rating", "DOUBLE PRECISION"}, // float64 maps to DOUBLE PRECISION
	}

	for _, c := range checks {
		// regex or simple string match.
		// "name TEXT"
		target := fmt.Sprintf("%s %s", c.col, c.kind)
		if !strings.Contains(migrationSQL, target) {
			t.Errorf("migration SQL missing column definition for %s. Expected '%s'", c.col, target)
		}
	}
}
