package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	_ "modernc.org/sqlite"

	web "workshop/internal/adapters/http"
	"workshop/internal/adapters/storage"
	accountStore "workshop/internal/adapters/storage/account"
	attendanceStore "workshop/internal/adapters/storage/attendance"
	classTypeStore "workshop/internal/adapters/storage/classtype"
	clipStorePkg "workshop/internal/adapters/storage/clip"
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
	themeStorePkg "workshop/internal/adapters/storage/theme"
	trainingGoalStore "workshop/internal/adapters/storage/traininggoal"
	waiverStore "workshop/internal/adapters/storage/waiver"
	"workshop/internal/application/orchestrators"
)

func main() {
	// Initialize database with WAL mode, foreign keys, and busy timeout per DB_GUIDE
	dsn := "workshop.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Connection pool settings for WAL mode
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// Health check
	if err := db.Ping(); err != nil {
		log.Fatalf("database unreachable: %v", err)
	}

	// Initialize database schema
	if err := storage.InitDB(db); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	log.Println("Database initialized successfully!")

	// Create store instances
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
		ThemeStore:           themeStorePkg.NewSQLiteStore(db),
		ClipStore:            clipStorePkg.NewSQLiteStore(db),
	}

	// Seed default admin account if no accounts exist
	seedDeps := orchestrators.CreateAccountDeps{AccountStore: acctStore}
	if err := orchestrators.ExecuteSeedAdmin(context.Background(), seedDeps, "admin@workshop.co.nz", "workshop12345!"); err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}

	// Seed default programs and class types
	seedProgDeps := orchestrators.SeedProgramsDeps{ProgramStore: progStore, ClassTypeStore: ctStore}
	if err := orchestrators.ExecuteSeedPrograms(context.Background(), seedProgDeps); err != nil {
		log.Fatalf("failed to seed programs: %v", err)
	}

	// Seed synthetic data for development
	adminAcct, err := acctStore.GetByEmail(context.Background(), "admin@workshop.co.nz")
	if err != nil {
		log.Fatalf("failed to get admin account for seeding: %v", err)
	}
	synDeps := orchestrators.SyntheticSeedDeps{
		AccountStore:         acctStore,
		MemberStore:          stores.MemberStore,
		WaiverStore:          stores.WaiverStore,
		InjuryStore:          stores.InjuryStore,
		AttendanceStore:      stores.AttendanceStore,
		ScheduleStore:        stores.ScheduleStore,
		TermStore:            stores.TermStore,
		HolidayStore:         stores.HolidayStore,
		NoticeStore:          stores.NoticeStore,
		GradingRecordStore:   stores.GradingRecordStore,
		GradingConfigStore:   stores.GradingConfigStore,
		GradingProposalStore: stores.GradingProposalStore,
		MessageStore:         stores.MessageStore,
		ObservationStore:     stores.ObservationStore,
		MilestoneStore:       stores.MilestoneStore,
		TrainingGoalStore:    stores.TrainingGoalStore,
		ClassTypeStore:       stores.ClassTypeStore,
		ThemeStore:           stores.ThemeStore,
		ClipStore:            stores.ClipStore,
	}
	if err := orchestrators.ExecuteSeedSynthetic(context.Background(), synDeps, adminAcct.ID); err != nil {
		log.Fatalf("failed to seed synthetic data: %v", err)
	}

	// Create HTTP handler with middleware
	mux := web.NewMux("static", stores)

	// Start server
	log.Println("Layer 1a+1b+2 Core Operations + Engagement + Spine is ready!")
	log.Println("Server starting on http://localhost:8080")
	log.Println("\nAvailable endpoints:")
	log.Println("  GET  /login                 - Login page")
	log.Println("  POST /login                 - Authenticate")
	log.Println("  POST /logout                - Log out")
	log.Println("  POST /members               - Register member")
	log.Println("  GET  /members               - List members")
	log.Println("  GET  /members/profile       - Get member profile")
	log.Println("  POST /checkin               - Check in member")
	log.Println("  GET  /attendance            - Today's attendance")
	log.Println("  POST /injuries              - Report injury")
	log.Println("  POST /waivers               - Sign waiver")
	log.Println("  GET  /api/members/search    - Search members by name")
	log.Println("  POST /api/members/archive   - Archive a member")
	log.Println("  POST /api/members/restore   - Restore archived member")
	log.Println("  POST /api/guest/checkin     - Guest check-in + waiver")
	log.Println("  GET  /api/classes/today     - Today's classes")
	log.Println("  POST /api/kiosk/launch      - Launch kiosk mode")
	log.Println("  POST /api/kiosk/exit        - Exit kiosk mode")
	log.Println("  --- Layer 1b: Engagement ---")
	log.Println("  GET  /api/training-log      - Member training log")
	log.Println("  GET  /api/members/inactive  - Inactive member radar")
	log.Println("  *    /api/notices           - Notices (GET/POST)")
	log.Println("  *    /api/grading/proposals - Grading proposals (GET/POST)")
	log.Println("  *    /api/messages          - Direct messages (GET/POST)")
	log.Println("  *    /api/observations      - Coach observations (GET/POST)")
	log.Println("\nDefault admin: admin@workshop.co.nz / workshop12345!")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
