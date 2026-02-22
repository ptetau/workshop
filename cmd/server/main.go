package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "modernc.org/sqlite"

	emailPkg "workshop/internal/adapters/email"
	web "workshop/internal/adapters/http"
	"workshop/internal/adapters/http/perf"
	"workshop/internal/adapters/storage"
	accountStore "workshop/internal/adapters/storage/account"
	attendanceStore "workshop/internal/adapters/storage/attendance"
	bugboxStorePkg "workshop/internal/adapters/storage/bugbox"
	calendarStorePkg "workshop/internal/adapters/storage/calendar"
	classTypeStore "workshop/internal/adapters/storage/classtype"
	clipStorePkg "workshop/internal/adapters/storage/clip"
	emailStorePkg "workshop/internal/adapters/storage/email"
	estimatedHoursStorePkg "workshop/internal/adapters/storage/estimatedhours"
	featureFlagStorePkg "workshop/internal/adapters/storage/featureflag"
	gradingStore "workshop/internal/adapters/storage/grading"
	holidayStore "workshop/internal/adapters/storage/holiday"
	injuryStore "workshop/internal/adapters/storage/injury"
	memberStore "workshop/internal/adapters/storage/member"
	messageStore "workshop/internal/adapters/storage/message"
	milestoneStore "workshop/internal/adapters/storage/milestone"
	noticeStore "workshop/internal/adapters/storage/notice"
	observationStore "workshop/internal/adapters/storage/observation"
	outboxStorePkg "workshop/internal/adapters/storage/outbox"
	personalgoalStorePkg "workshop/internal/adapters/storage/personalgoal"
	programStore "workshop/internal/adapters/storage/program"
	rotorStorePkg "workshop/internal/adapters/storage/rotor"
	scheduleStore "workshop/internal/adapters/storage/schedule"
	termStore "workshop/internal/adapters/storage/term"
	themeStorePkg "workshop/internal/adapters/storage/theme"
	trainingGoalStore "workshop/internal/adapters/storage/traininggoal"
	waiverStore "workshop/internal/adapters/storage/waiver"
	"workshop/internal/application/orchestrators"
)

// version is set at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	// Initialize database with WAL mode, foreign keys, and busy timeout per DB_GUIDE
	dbPath := "workshop.db"
	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)"
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

	// Run database migrations
	if err := storage.MigrateDB(db, dbPath); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Println("Database initialized successfully!")

	// Performance instrumentation: wrap DB with timing, create collector
	collector := perf.NewCollector(perf.DefaultRingSize)
	timedDB := storage.NewTimedDB(db, collector)

	// Create store instances (using timed DB for query instrumentation)
	acctStore := accountStore.NewSQLiteStore(timedDB)
	progStore := programStore.NewSQLiteStore(timedDB)
	ctStore := classTypeStore.NewSQLiteStore(timedDB)
	stores := &web.Stores{
		AccountStore:             acctStore,
		FeatureFlagStore:         featureFlagStorePkg.NewSQLiteStore(timedDB),
		MemberStore:              memberStore.NewSQLiteStore(timedDB),
		WaiverStore:              waiverStore.NewSQLiteStore(timedDB),
		InjuryStore:              injuryStore.NewSQLiteStore(timedDB),
		AttendanceStore:          attendanceStore.NewSQLiteStore(timedDB),
		ProgramStore:             progStore,
		ClassTypeStore:           ctStore,
		ScheduleStore:            scheduleStore.NewSQLiteStore(timedDB),
		TermStore:                termStore.NewSQLiteStore(timedDB),
		HolidayStore:             holidayStore.NewSQLiteStore(timedDB),
		NoticeStore:              noticeStore.NewSQLiteStore(timedDB),
		GradingRecordStore:       gradingStore.NewRecordSQLiteStore(timedDB),
		GradingConfigStore:       gradingStore.NewConfigSQLiteStore(timedDB),
		GradingProposalStore:     gradingStore.NewProposalSQLiteStore(timedDB),
		GradingNoteStore:         gradingStore.NewNoteSQLiteStore(timedDB),
		GradingMemberConfigStore: gradingStore.NewMemberConfigSQLiteStore(timedDB),
		MessageStore:             messageStore.NewSQLiteStore(timedDB),
		ObservationStore:         observationStore.NewSQLiteStore(timedDB),
		MilestoneStore:           milestoneStore.NewSQLiteStore(timedDB),
		MemberMilestoneStore:     milestoneStore.NewMemberMilestoneSQLiteStore(timedDB),
		TrainingGoalStore:        trainingGoalStore.NewSQLiteStore(timedDB),
		ThemeStore:               themeStorePkg.NewSQLiteStore(timedDB),
		ClipStore:                clipStorePkg.NewSQLiteStore(timedDB),
		EmailStore:               emailStorePkg.NewSQLiteStore(timedDB),
		EstimatedHoursStore:      estimatedHoursStorePkg.NewSQLiteStore(timedDB),
		RotorStore:               rotorStorePkg.NewSQLiteStore(timedDB),
		CalendarEventStore:       calendarStorePkg.NewSQLiteStore(timedDB),
		CompetitionInterestStore: calendarStorePkg.NewSQLiteStore(timedDB),
		BugBoxStore:              bugboxStorePkg.NewSQLiteStore(timedDB),
		OutboxStore:              outboxStorePkg.NewSQLiteStore(timedDB),
		PersonalGoalStore:        personalgoalStorePkg.NewSQLiteStore(timedDB),
	}

	// Seed default admin account if no accounts exist
	adminEmail := envOrDefault("WORKSHOP_ADMIN_EMAIL", "info@workshopjiujitsu.co.nz")
	adminPassword := envOrDefault("WORKSHOP_ADMIN_PASSWORD", "Umami monster")
	seedDeps := orchestrators.CreateAccountDeps{AccountStore: acctStore}
	if err := orchestrators.ExecuteSeedAdmin(context.Background(), seedDeps, adminEmail, adminPassword); err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}

	// Seed default programs and class types
	seedProgDeps := orchestrators.SeedProgramsDeps{ProgramStore: progStore, ClassTypeStore: ctStore}
	if err := orchestrators.ExecuteSeedPrograms(context.Background(), seedProgDeps); err != nil {
		log.Fatalf("failed to seed programs: %v", err)
	}

	// Seed NZ grappling competitions into calendar
	seedCompDeps := orchestrators.SeedCompetitionsDeps{EventStore: stores.CalendarEventStore}
	if err := orchestrators.ExecuteSeedCompetitions(context.Background(), seedCompDeps); err != nil {
		log.Fatalf("failed to seed competitions: %v", err)
	}

	// Seed test accounts for each role (all environments, idempotent)
	testAcctDeps := orchestrators.TestAccountSeedDeps{
		AccountStore: acctStore,
		MemberStore:  stores.MemberStore,
	}
	if err := orchestrators.ExecuteSeedTestAccounts(context.Background(), testAcctDeps); err != nil {
		log.Fatalf("failed to seed test accounts: %v", err)
	}

	// Seed synthetic data for development only
	if os.Getenv("WORKSHOP_ENV") != "production" {
		adminAcct, err := acctStore.GetByEmail(context.Background(), adminEmail)
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
		log.Println("Synthetic seed data loaded (dev mode)")
	}

	// Configure email sender
	resendKey := os.Getenv("WORKSHOP_RESEND_KEY")
	emailFrom := envOrDefault("WORKSHOP_RESEND_FROM", "Workshop Jiu Jitsu <noreply@workshopjiujitsu.co.nz>")
	emailReply := envOrDefault("WORKSHOP_REPLY_TO", "info@workshopjiujitsu.co.nz")
	if resendKey != "" {
		web.SetEmailSender(emailPkg.NewResendSender(resendKey, emailFrom), emailFrom, emailReply)
		log.Println("Email sender configured (Resend)")
	} else {
		web.SetEmailSender(emailPkg.NewNoopSender(), emailFrom, emailReply)
		if os.Getenv("WORKSHOP_ENV") == "production" {
			log.Println("WARNING: WORKSHOP_RESEND_KEY is not set — email delivery is DISABLED in production")
		} else {
			log.Println("Email sender configured (noop — set WORKSHOP_RESEND_KEY for real delivery)")
		}
	}

	// Start outbox background worker for retrying failed external integrations
	outboxStopCh := make(chan struct{})
	outboxProcessor := orchestrators.NewOutboxProcessor(stores.OutboxStore, nil) // Executors wired later
	orchestrators.StartBackgroundWorker(outboxProcessor, 1*time.Minute, outboxStopCh)
	defer close(outboxStopCh)

	// Create HTTP handler with middleware (pass collector for timing + dashboard)
	mux := web.NewMux("static", stores, collector)

	// Start server
	addr := envOrDefault("WORKSHOP_ADDR", ":8080")
	log.Printf("Workshop %s starting on %s (env=%s, schema=%d)", version, addr, envOrDefault("WORKSHOP_ENV", "development"), storage.LatestSchemaVersion())

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
