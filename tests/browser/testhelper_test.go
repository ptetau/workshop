package browser_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"

	_ "modernc.org/sqlite"

	web "workshop/internal/adapters/http"
	"workshop/internal/adapters/http/middleware"
	"workshop/internal/adapters/storage"
	accountStore "workshop/internal/adapters/storage/account"
	attendanceStore "workshop/internal/adapters/storage/attendance"
	classTypeStore "workshop/internal/adapters/storage/classtype"
	clipStore "workshop/internal/adapters/storage/clip"
	gradingStore "workshop/internal/adapters/storage/grading"
	holidayStore "workshop/internal/adapters/storage/holiday"
	injuryStore "workshop/internal/adapters/storage/injury"
	memberStore "workshop/internal/adapters/storage/member"
	messageStore "workshop/internal/adapters/storage/message"
	milestoneStore "workshop/internal/adapters/storage/milestone"
	noticeStore "workshop/internal/adapters/storage/notice"
	observationStore "workshop/internal/adapters/storage/observation"
	programStore "workshop/internal/adapters/storage/program"
	scheduleStore "workshop/internal/adapters/storage/schedule"
	termStore "workshop/internal/adapters/storage/term"
	themeStore "workshop/internal/adapters/storage/theme"
	trainingGoalStore "workshop/internal/adapters/storage/traininggoal"
	waiverStore "workshop/internal/adapters/storage/waiver"
	"workshop/internal/application/orchestrators"
)

// testApp holds the running test server and Playwright handles.
type testApp struct {
	BaseURL string
	DB      *sql.DB
	Server  *http.Server
	PW      *playwright.Playwright
	Browser playwright.Browser
	Stores  *web.Stores
	AdminID string
	tmpDir  string
}

// newTestApp creates a fully wired app with a temp SQLite DB and starts an HTTP server.
func newTestApp(t *testing.T) *testApp {
	t.Helper()

	// Create temp directory for the database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// Run migrations
	if err := storage.MigrateDB(db, dbPath); err != nil {
		t.Fatalf("failed to migrate test DB: %v", err)
	}

	// Create stores
	acctStore := accountStore.NewSQLiteStore(db)
	progStore := programStore.NewSQLiteStore(db)
	ctStore := classTypeStore.NewSQLiteStore(db)
	stores := &web.Stores{
		AccountStore:         acctStore,
		MemberStore:          memberStore.NewSQLiteStore(db),
		WaiverStore:          waiverStore.NewSQLiteStore(db),
		InjuryStore:          injuryStore.NewSQLiteStore(db),
		AttendanceStore:      attendanceStore.NewSQLiteStore(db),
		ProgramStore:         progStore,
		ClassTypeStore:       ctStore,
		ScheduleStore:        scheduleStore.NewSQLiteStore(db),
		TermStore:            termStore.NewSQLiteStore(db),
		HolidayStore:         holidayStore.NewSQLiteStore(db),
		NoticeStore:          noticeStore.NewSQLiteStore(db),
		GradingRecordStore:   gradingStore.NewRecordSQLiteStore(db),
		GradingConfigStore:   gradingStore.NewConfigSQLiteStore(db),
		GradingProposalStore: gradingStore.NewProposalSQLiteStore(db),
		MessageStore:         messageStore.NewSQLiteStore(db),
		ObservationStore:     observationStore.NewSQLiteStore(db),
		MilestoneStore:       milestoneStore.NewSQLiteStore(db),
		TrainingGoalStore:    trainingGoalStore.NewSQLiteStore(db),
		ThemeStore:           themeStore.NewSQLiteStore(db),
		ClipStore:            clipStore.NewSQLiteStore(db),
	}

	// Seed admin (without PasswordChangeRequired so login goes straight to dashboard)
	ctx := context.Background()
	adminID, err := orchestrators.ExecuteCreateAccount(ctx, orchestrators.CreateAccountInput{
		Email:                  "admin@test.com",
		Password:               "TestPass123!",
		Role:                   "admin",
		PasswordChangeRequired: false,
	}, orchestrators.CreateAccountDeps{AccountStore: acctStore})
	if err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	// Seed programs
	seedProgDeps := orchestrators.SeedProgramsDeps{ProgramStore: progStore, ClassTypeStore: ctStore}
	if err := orchestrators.ExecuteSeedPrograms(ctx, seedProgDeps); err != nil {
		t.Fatalf("failed to seed programs: %v", err)
	}

	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Change to project root so relative template/static paths work
	projectRoot := findProjectRoot(t)
	origDir, _ := os.Getwd()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("failed to chdir to project root: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	// Add test port to CSRF trusted origins before creating mux
	middleware.ExtraTrustedOrigins = append(middleware.ExtraTrustedOrigins,
		fmt.Sprintf("127.0.0.1:%d", port),
		fmt.Sprintf("localhost:%d", port),
	)

	// Start HTTP server
	mux := web.NewMux("static", stores)
	srv := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("test server error: %v", err)
		}
	}()

	// Wait for server to be ready
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	for i := 0; i < 50; i++ {
		resp, err := http.Get(baseURL + "/login")
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("failed to start Playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to launch browser: %v", err)
	}

	app := &testApp{
		BaseURL: baseURL,
		DB:      db,
		Server:  srv,
		PW:      pw,
		Browser: browser,
		Stores:  stores,
		AdminID: adminID,
		tmpDir:  tmpDir,
	}

	t.Cleanup(func() {
		browser.Close()
		pw.Stop()
		srv.Close()
		db.Close()
	})

	return app
}

// newPage creates a new browser page (tab).
func (a *testApp) newPage(t *testing.T) playwright.Page {
	t.Helper()
	page, err := a.Browser.NewPage()
	if err != nil {
		t.Fatalf("failed to create page: %v", err)
	}
	t.Cleanup(func() { page.Close() })
	return page
}

// login navigates to the login page and logs in as admin.
func (a *testApp) login(t *testing.T, page playwright.Page) {
	t.Helper()
	_, err := page.Goto(a.BaseURL + "/login")
	if err != nil {
		t.Fatalf("failed to navigate to login: %v", err)
	}
	if err := page.Locator("input[name=Email]").Fill("admin@test.com"); err != nil {
		t.Fatalf("failed to fill email: %v", err)
	}
	if err := page.Locator("input[name=Password]").Fill("TestPass123!"); err != nil {
		t.Fatalf("failed to fill password: %v", err)
	}
	if err := page.Locator("button[type=submit]").Click(); err != nil {
		t.Fatalf("failed to click login: %v", err)
	}
	// Wait for redirect to dashboard
	if err := page.WaitForURL(a.BaseURL+"/dashboard", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("login did not redirect to dashboard: %v", err)
	}
}

// findProjectRoot walks up from the working directory to find the project root (contains go.mod).
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root (go.mod) from working directory")
		}
		dir = parent
	}
}
