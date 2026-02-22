package consent

import (
	"time"
)

// Type represents the type of consent given.
type Type string

const (
	TypeMarketing      Type = "marketing"
	TypeDataProcessing Type = "data_processing"
	TypePhotos         Type = "photos"
	TypeThirdParty     Type = "third_party"
)

// Consent represents a member's consent record.
type Consent struct {
	ID        string     `json:"id"`
	MemberID  string     `json:"member_id"`
	Type      Type       `json:"type"`
	Granted   bool       `json:"granted"`
	GrantedAt time.Time  `json:"granted_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	Source    string     `json:"source"`
	IPAddress string     `json:"ip_address"`
	UserAgent string     `json:"user_agent"`
	Version   string     `json:"version"`
}

// NewConsent creates a new consent record.
// PRE: memberID and consentType are non-empty
// POST: Returns a Consent with current timestamp and granted=true
func NewConsent(memberID string, consentType Type, source, ipAddress, userAgent, version string) Consent {
	return Consent{
		ID:        generateID(),
		MemberID:  memberID,
		Type:      consentType,
		Granted:   true,
		GrantedAt: time.Now(),
		Source:    source,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Version:   version,
	}
}

// Revoke marks the consent as revoked.
// PRE: Consent was previously granted and not already revoked
// POST: RevokedAt is set to current time, Granted is false
func (c *Consent) Revoke() error {
	if !c.Granted {
		return ErrConsentNotGranted
	}
	if c.RevokedAt != nil {
		return ErrConsentAlreadyRevoked
	}
	now := time.Now()
	c.RevokedAt = &now
	c.Granted = false
	return nil
}

// IsValid returns true if consent is currently granted and not revoked.
// INVARIANT: Granted=true and RevokedAt=nil
func (c Consent) IsValid() bool {
	return c.Granted && c.RevokedAt == nil
}

// ErrConsentNotGranted is returned when trying to revoke consent that was never granted.
var ErrConsentNotGranted = &ConsentError{Message: "consent was never granted"}

// ErrConsentAlreadyRevoked is returned when trying to revoke already revoked consent.
var ErrConsentAlreadyRevoked = &ConsentError{Message: "consent already revoked"}

// ConsentError represents a consent-related error.
type ConsentError struct {
	Message string
}

// Error implements the error interface.
func (e *ConsentError) Error() string {
	return e.Message
}

// generateID generates a unique identifier.
func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

// randomString generates a random alphanumeric string.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
