package web

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"workshop/internal/adapters/http/middleware"
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
)

// Stores holds all storage dependencies.
type Stores struct {
	AccountStore         accountStore.Store
	MemberStore          memberStore.Store
	WaiverStore          waiverStore.Store
	InjuryStore          injuryStore.Store
	AttendanceStore      attendanceStore.Store
	ProgramStore         programStore.Store
	ClassTypeStore       classTypeStore.Store
	ScheduleStore        scheduleStore.Store
	TermStore            termStore.Store
	HolidayStore         holidayStore.Store
	NoticeStore          noticeStore.Store
	GradingRecordStore   gradingStore.RecordStore
	GradingConfigStore   gradingStore.ConfigStore
	GradingProposalStore gradingStore.ProposalStore
	MessageStore         messageStore.Store
	ObservationStore     observationStore.Store
	MilestoneStore       milestoneStore.Store
	TrainingGoalStore    trainingGoalStore.Store
	ThemeStore           themeStore.Store
	ClipStore            clipStore.Store
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

// NewMux wires HTTP handlers for the app.
func NewMux(staticDir string, s *Stores) http.Handler {
	stores = s
	sessions = middleware.NewSessionStore()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	registerRoutes(mux)

	// CSRF key: 32-byte hex-encoded secret from env var
	csrfKey := loadCSRFKey()

	// Rate limiter: 10 requests per second per IP (OWASP A04)
	limiter := middleware.NewRateLimiter(10, time.Second)

	// Apply middleware: Auth -> CSRF -> SecurityHeaders -> RateLimit -> Mux
	return middleware.Chain(mux,
		middleware.SecurityHeaders,
		middleware.CSRF(csrfKey),
		middleware.Auth(sessions),
		middleware.RateLimit(limiter),
	)
}
