package export

import (
	"encoding/json"
	"errors"
	"time"
)

// Status constants for export request lifecycle.
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusReady      = "ready"
	StatusDownloaded = "downloaded"
	StatusExpired    = "expired"
)

// Format constants for export file format.
const (
	FormatJSON = "json"
	FormatCSV  = "csv"
)

// Domain errors.
var (
	ErrEmptyMemberID  = errors.New("member_id is required")
	ErrEmptyRequestID = errors.New("request_id is required")
	ErrInvalidStatus  = errors.New("invalid status transition")
	ErrNotReady       = errors.New("export not ready for download")
)

// Request represents a member's data export request (GDPR Article 20).
type Request struct {
	ID           string
	MemberID     string
	Status       string
	Format       string
	RequestedAt  time.Time
	CompletedAt  *time.Time
	DownloadedAt *time.Time
	ExpiredAt    *time.Time
	FilePath     string // Temporary file path (secure, random filename)
	FileSize     int64
	IPAddress    string // Audit trail
	UserAgent    string // Audit trail
}

// Data represents the complete member data export payload.
// This includes all personal data across all domains.
type Data struct {
	Member              MemberData           `json:"member"`
	Account             AccountData          `json:"account,omitempty"`
	Attendance          []AttendanceRecord   `json:"attendance,omitempty"`
	Injuries            []InjuryRecord       `json:"injuries,omitempty"`
	Waivers             []WaiverRecord       `json:"waivers,omitempty"`
	GradingHistory      []GradingRecord      `json:"grading_history,omitempty"`
	Messages            []MessageRecord      `json:"messages,omitempty"`
	Observations        []ObservationRecord  `json:"observations,omitempty"`
	Milestones          []MilestoneRecord    `json:"milestones,omitempty"`
	TrainingGoals       []TrainingGoalRecord `json:"training_goals,omitempty"`
	PersonalGoals       []PersonalGoalRecord `json:"personal_goals,omitempty"`
	CompetitionInterest []CompetitionRecord  `json:"competition_interest,omitempty"`
	BugReports          []BugReportRecord    `json:"bug_reports,omitempty"`
	ExportMetadata      Metadata             `json:"export_metadata"`
}

// MemberData represents core member information.
type MemberData struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Program       string    `json:"program"`
	Fee           int       `json:"fee"`
	Frequency     string    `json:"frequency"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	Belt          string    `json:"belt"`
	Stripe        int       `json:"stripe"`
	GradingMetric string    `json:"grading_metric"`
}

// AccountData represents account-level information.
type AccountData struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}

// AttendanceRecord represents a class attendance.
type AttendanceRecord struct {
	ID           string     `json:"id"`
	ClassDate    string     `json:"class_date"`
	CheckInTime  time.Time  `json:"check_in_time"`
	CheckOutTime *time.Time `json:"check_out_time,omitempty"`
	MatHours     float64    `json:"mat_hours"`
	ClassType    string     `json:"class_type"`
}

// InjuryRecord represents an injury report.
type InjuryRecord struct {
	ID          string    `json:"id"`
	BodyPart    string    `json:"body_part"`
	Description string    `json:"description"`
	ReportedAt  time.Time `json:"reported_at"`
	Status      string    `json:"status,omitempty"`
}

// WaiverRecord represents a signed waiver.
type WaiverRecord struct {
	ID            string    `json:"id"`
	AcceptedTerms bool      `json:"accepted_terms"`
	SignedAt      time.Time `json:"signed_at"`
	IPaddress     string    `json:"ip_address,omitempty"`
}

// GradingRecord represents a belt promotion.
type GradingRecord struct {
	ID         string    `json:"id"`
	Belt       string    `json:"belt"`
	Stripe     int       `json:"stripe"`
	PromotedAt time.Time `json:"promoted_at"`
	Method     string    `json:"method"`
}

// MessageRecord represents a message/conversation.
type MessageRecord struct {
	ID        string     `json:"id"`
	Subject   string     `json:"subject"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

// ObservationRecord represents a coach observation.
type ObservationRecord struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// MilestoneRecord represents earned milestones.
type MilestoneRecord struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	EarnedAt time.Time `json:"earned_at"`
}

// TrainingGoalRecord represents training goals.
type TrainingGoalRecord struct {
	ID        string    `json:"id"`
	Target    int       `json:"target"`
	Period    string    `json:"period"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
}

// PersonalGoalRecord represents personal goals.
type PersonalGoalRecord struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Target      int       `json:"target"`
	Unit        string    `json:"unit"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	Progress    int       `json:"progress"`
	CreatedAt   time.Time `json:"created_at"`
}

// CompetitionRecord represents competition interest.
type CompetitionRecord struct {
	ID         string    `json:"id"`
	EventID    string    `json:"event_id"`
	EventTitle string    `json:"event_title"`
	CreatedAt  time.Time `json:"created_at"`
}

// BugReportRecord represents submitted bug reports.
type BugReportRecord struct {
	ID          string    `json:"id"`
	Summary     string    `json:"summary"`
	Route       string    `json:"route"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// Metadata contains information about the export itself.
type Metadata struct {
	ExportDate  time.Time `json:"export_date"`
	Format      string    `json:"format"`
	Version     string    `json:"version"`
	RecordCount int       `json:"record_count"`
}

// Validate checks that the Request has valid data.
// PRE: Request fields may be empty
// POST: Returns nil if valid, error otherwise
// INVARIANT: ID, MemberID, RequestedAt must be non-empty, Format must be valid
func (r *Request) Validate() error {
	if r.ID == "" {
		return ErrEmptyRequestID
	}
	if r.MemberID == "" {
		return ErrEmptyMemberID
	}
	if r.RequestedAt.IsZero() {
		return errors.New("requested_at must be set")
	}
	if r.Format == "" {
		r.Format = FormatJSON // Default
	}
	if r.Format != FormatJSON && r.Format != FormatCSV {
		return errors.New("invalid format: must be 'json' or 'csv'")
	}
	return nil
}

// MarkProcessing transitions the request to processing state.
// PRE: Status is pending
// POST: Status set to processing
// INVARIANT: Request must be in pending status
func (r *Request) MarkProcessing() error {
	if r.Status != StatusPending {
		return ErrInvalidStatus
	}
	r.Status = StatusProcessing
	return nil
}

// MarkReady transitions the request to ready state with file info.
// PRE: Status is processing, filePath and fileSize are valid
// POST: Status set to ready, FilePath/Size set, CompletedAt set, ExpiredAt set (7 days)
// INVARIANT: Request must be in processing status
func (r *Request) MarkReady(filePath string, fileSize int64) error {
	if r.Status != StatusProcessing {
		return ErrInvalidStatus
	}
	now := time.Now()
	r.Status = StatusReady
	r.FilePath = filePath
	r.FileSize = fileSize
	r.CompletedAt = &now
	// Set expiration (7 days from now)
	expiredAt := now.Add(7 * 24 * time.Hour)
	r.ExpiredAt = &expiredAt
	return nil
}

// MarkDownloaded records that the export was downloaded.
// PRE: Status is ready
// POST: Status set to downloaded, DownloadedAt set to now
// INVARIANT: Request must be in ready status
func (r *Request) MarkDownloaded() error {
	if r.Status != StatusReady {
		return ErrNotReady
	}
	now := time.Now()
	r.Status = StatusDownloaded
	r.DownloadedAt = &now
	return nil
}

// IsExpired returns true if the download link has expired.
// PRE: ExpiredAt may be nil
// POST: Returns true if ExpiredAt is set and current time is after it
// INVARIANT: None
func (r *Request) IsExpired() bool {
	if r.ExpiredAt == nil {
		return false
	}
	return time.Now().After(*r.ExpiredAt)
}

// CanDownload returns true if the export is ready and not expired.
// PRE: Status and ExpiredAt are known
// POST: Returns true if status is ready and not expired
// INVARIANT: Status must be ready, not past expiration
func (r *Request) CanDownload() bool {
	return r.Status == StatusReady && !r.IsExpired()
}

// ToJSON serializes the Data to JSON format.
// PRE: Data fields are populated
// POST: Returns JSON-encoded bytes of Data
// INVARIANT: None
func (d *Data) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

// ToCSV serializes specific Data sections to CSV format.
// PRE: Data fields are populated
// POST: Returns map of section names to CSV bytes
// INVARIANT: Data must be valid (currently not implemented)
func (d *Data) ToCSV() (map[string][]byte, error) {
	// CSV export is complex - would need proper CSV encoding
	// For now, return placeholder
	return nil, errors.New("CSV export not yet implemented")
}

// NewRequest creates a new export request.
func NewRequest(id, memberID, format, ipAddress, userAgent string) *Request {
	if format == "" {
		format = FormatJSON
	}
	return &Request{
		ID:          id,
		MemberID:    memberID,
		Status:      StatusPending,
		Format:      format,
		RequestedAt: time.Now(),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}
}
