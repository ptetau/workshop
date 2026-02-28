package audit

import (
	"time"
)

// Category represents the type of audit event.
type Category string

const (
	CategoryAccount    Category = "account"
	CategoryMember     Category = "member"
	CategoryAttendance Category = "attendance"
	CategoryPrivacy    Category = "privacy"
	CategorySecurity   Category = "security"
	CategoryBilling    Category = "billing"
	CategorySystem     Category = "system"
)

// Action represents the action that occurred.
type Action string

const (
	ActionCreate   Action = "create"
	ActionUpdate   Action = "update"
	ActionDelete   Action = "delete"
	ActionLogin    Action = "login"
	ActionLogout   Action = "logout"
	ActionExport   Action = "export"
	ActionDownload Action = "download"
	ActionView     Action = "view"
)

// Severity represents the severity level of an audit event.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Event represents a single audit log entry.
type Event struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Category     Category  `json:"category"`
	Action       Action    `json:"action"`
	Severity     Severity  `json:"severity"`
	ActorID      string    `json:"actor_id"`
	ActorEmail   string    `json:"actor_email"`
	ActorRole    string    `json:"actor_role"`
	ResourceID   string    `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	Description  string    `json:"description"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	Metadata     string    `json:"metadata"`
}

// NewEvent creates a new audit event with the current timestamp.
// PRE: actorID and action are non-empty
// POST: Returns an Event with the current timestamp and provided fields
func NewEvent(actorID, actorEmail, actorRole string, category Category, action Action) Event {
	return Event{
		ID:         generateID(),
		Timestamp:  time.Now(),
		Category:   category,
		Action:     action,
		Severity:   SeverityInfo,
		ActorID:    actorID,
		ActorEmail: actorEmail,
		ActorRole:  actorRole,
	}
}

// WithSeverity sets the severity level.
// PRE: s is valid severity
// POST: Event severity is updated
func (e Event) WithSeverity(s Severity) Event {
	e.Severity = s
	return e
}

// WithResource sets resource information.
// PRE: resourceType and resourceID are non-empty
// POST: Event resource fields are populated
func (e Event) WithResource(resourceType, resourceID string) Event {
	e.ResourceType = resourceType
	e.ResourceID = resourceID
	return e
}

// WithDescription sets the event description.
// PRE: description is non-empty
// POST: Event description is set
func (e Event) WithDescription(desc string) Event {
	e.Description = desc
	return e
}

// WithRequest sets IP address and user agent from HTTP request.
// PRE: ipAddress and userAgent are non-empty
// POST: Event network fields are populated
func (e Event) WithRequest(ipAddress, userAgent string) Event {
	e.IPAddress = ipAddress
	e.UserAgent = userAgent
	return e
}

// WithMetadata sets optional JSON metadata.
// PRE: metadata is valid JSON or empty
// POST: Event metadata is set
func (e Event) WithMetadata(metadata string) Event {
	e.Metadata = metadata
	return e
}

// generateID generates a unique identifier for the event.
func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

// randomString generates a random alphanumeric string of length n.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
