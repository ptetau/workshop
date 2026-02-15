package web

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"workshop/internal/adapters/email"
	"workshop/internal/adapters/http/middleware"
	"workshop/internal/adapters/http/perf"
	accountStore "workshop/internal/adapters/storage/account"
	attendanceStore "workshop/internal/adapters/storage/attendance"
	calendarStore "workshop/internal/adapters/storage/calendar"
	classTypeStore "workshop/internal/adapters/storage/classtype"
	clipStore "workshop/internal/adapters/storage/clip"
	emailStore "workshop/internal/adapters/storage/email"
	estimatedHoursStore "workshop/internal/adapters/storage/estimatedhours"
	featureFlagStore "workshop/internal/adapters/storage/featureflag"
	gradingStore "workshop/internal/adapters/storage/grading"
	holidayStore "workshop/internal/adapters/storage/holiday"
	injuryStore "workshop/internal/adapters/storage/injury"
	memberStore "workshop/internal/adapters/storage/member"
	messageStore "workshop/internal/adapters/storage/message"
	milestoneStore "workshop/internal/adapters/storage/milestone"
	noticeStore "workshop/internal/adapters/storage/notice"
	observationStore "workshop/internal/adapters/storage/observation"
	programStore "workshop/internal/adapters/storage/program"
	rotorStore "workshop/internal/adapters/storage/rotor"
	scheduleStore "workshop/internal/adapters/storage/schedule"
	termStore "workshop/internal/adapters/storage/term"
	themeStore "workshop/internal/adapters/storage/theme"
	trainingGoalStore "workshop/internal/adapters/storage/traininggoal"
	waiverStore "workshop/internal/adapters/storage/waiver"
)

// Stores holds all storage dependencies.
type Stores struct {
	AccountStore             accountStore.Store
	FeatureFlagStore         featureFlagStore.Store
	MemberStore              memberStore.Store
	WaiverStore              waiverStore.Store
	InjuryStore              injuryStore.Store
	AttendanceStore          attendanceStore.Store
	ProgramStore             programStore.Store
	ClassTypeStore           classTypeStore.Store
	ScheduleStore            scheduleStore.Store
	TermStore                termStore.Store
	HolidayStore             holidayStore.Store
	NoticeStore              noticeStore.Store
	GradingRecordStore       gradingStore.RecordStore
	GradingConfigStore       gradingStore.ConfigStore
	GradingProposalStore     gradingStore.ProposalStore
	GradingNoteStore         gradingStore.NoteStore
	GradingMemberConfigStore gradingStore.MemberConfigStore
	MessageStore             messageStore.Store
	ObservationStore         observationStore.Store
	MilestoneStore           milestoneStore.Store
	MemberMilestoneStore     milestoneStore.MemberMilestoneStore
	TrainingGoalStore        trainingGoalStore.Store
	ThemeStore               themeStore.Store
	ClipStore                clipStore.Store
	EmailStore               emailStore.Store
	EstimatedHoursStore      estimatedHoursStore.Store
	RotorStore               rotorStore.Store
	CalendarEventStore       calendarStore.Store
}

// loadCSRFKey reads the CSRF secret from WORKSHOP_CSRF_KEY (hex-encoded, 32 bytes).
// In production, the key MUST be set. In development, a random key is generated per startup.
func loadCSRFKey() []byte {
	if keyHex := os.Getenv("WORKSHOP_CSRF_KEY"); keyHex != "" {
		key, err := hex.DecodeString(keyHex)
		if err != nil || len(key) != 32 {
			log.Fatal("WORKSHOP_CSRF_KEY must be 64 hex characters (32 bytes)")
		}
		return key
	}
	if os.Getenv("WORKSHOP_ENV") == "production" {
		log.Fatal("WORKSHOP_CSRF_KEY is required in production")
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatalf("failed to generate CSRF key: %v", err)
	}
	log.Println("WARNING: using random CSRF key (sessions won't survive restart). Set WORKSHOP_CSRF_KEY for production.")
	return key
}

// Global stores instance (set by NewMux)
var stores *Stores

// Global session store instance
var sessions *middleware.SessionStore

// RateLimitPerSecond controls the per-IP rate limit. Tests can increase this.
var RateLimitPerSecond = 10

// Global perf collector (set by NewMux)
var perfCollector *perf.Collector

// Global email sender instance (set by SetEmailSender)
var emailSender email.Sender

// Email configuration
var emailFromAddress string
var emailReplyTo string

// SetEmailSender sets the global email sender for the application.
func SetEmailSender(sender email.Sender, from, replyTo string) {
	emailSender = sender
	emailFromAddress = from
	emailReplyTo = replyTo
}

// NewMux wires HTTP handlers for the app.
func NewMux(staticDir string, s *Stores, collector *perf.Collector) http.Handler {
	stores = s
	perfCollector = collector
	sessions = middleware.NewSessionStore()
	middleware.SecureCookies = os.Getenv("WORKSHOP_ENV") == "production"

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	registerRoutes(mux)

	// CSRF key: 32-byte hex-encoded secret from env var
	csrfKey := loadCSRFKey()

	// Rate limiter: configurable requests per second per IP (OWASP A04)
	limiter := middleware.NewRateLimiter(RateLimitPerSecond, time.Second)

	// Apply middleware: Timing -> Auth -> CSRF -> SecurityHeaders -> RateLimit -> Mux
	return middleware.Chain(mux,
		middleware.SecurityHeaders,
		middleware.CSRF(csrfKey),
		middleware.Auth(sessions),
		middleware.RateLimit(limiter),
		middleware.Timing(collector),
	)
}
