package grading

import (
	"errors"
	"time"
)

// Belt constants for Adults
const (
	BeltWhite  = "white"
	BeltBlue   = "blue"
	BeltPurple = "purple"
	BeltBrown  = "brown"
	BeltBlack  = "black"
)

// Belt constants for Kids
const (
	BeltGrey   = "grey"
	BeltYellow = "yellow"
	BeltOrange = "orange"
	BeltGreen  = "green"
	// Kids can also hold BeltWhite and BeltBlue
)

// Proposal statuses
const (
	ProposalPending  = "pending"
	ProposalApproved = "approved"
	ProposalRejected = "rejected"
)

// Promotion methods
const (
	MethodStandard = "standard"
	MethodOverride = "override"
	MethodInferred = "inferred"
)

// AdultBelts defines the adult belt progression order.
var AdultBelts = []string{BeltWhite, BeltBlue, BeltPurple, BeltBrown, BeltBlack}

// KidsBelts defines the kids belt progression order.
var KidsBelts = []string{BeltWhite, BeltGrey, BeltYellow, BeltOrange, BeltGreen, BeltBlue}

// Domain errors
var (
	ErrEmptyMemberID         = errors.New("member ID is required")
	ErrInvalidBelt           = errors.New("invalid belt value")
	ErrEmptyProposedBy       = errors.New("proposed_by is required")
	ErrInvalidProposalStatus = errors.New("proposal status must be one of: pending, approved, rejected")
	ErrAlreadyDecided        = errors.New("proposal has already been decided")
)

// Record represents an official belt promotion in a member's history.
type Record struct {
	ID         string
	MemberID   string
	Belt       string
	Stripe     int
	PromotedAt time.Time
	ProposedBy string // AccountID of coach who proposed
	ApprovedBy string // AccountID of admin who approved
	Method     string // standard or override
}

// Validate checks if the Record has valid data.
// PRE: Record struct is populated
// POST: Returns nil if valid, error otherwise
func (r *Record) Validate() error {
	if r.MemberID == "" {
		return ErrEmptyMemberID
	}
	if !isValidBelt(r.Belt) {
		return ErrInvalidBelt
	}
	if r.Stripe < 0 || r.Stripe > 4 {
		return errors.New("stripe must be between 0 and 4")
	}
	if r.PromotedAt.IsZero() {
		return errors.New("promoted_at must be set")
	}
	return nil
}

// Config holds per-belt eligibility thresholds configurable by Admin.
type Config struct {
	ID              string
	Program         string  // "adults" or "kids"
	Belt            string  // target belt
	FlightTimeHours float64 // required mat hours for adults (0 = not applicable)
	AttendancePct   float64 // required attendance % for kids (0 = not applicable)
	StripeCount     int     // stripes before next belt (default 4)
}

// Validate checks if the Config has valid data.
// PRE: Config struct is populated
// POST: Returns nil if valid, error otherwise
func (c *Config) Validate() error {
	if c.Program == "" {
		return errors.New("program is required")
	}
	if !isValidBelt(c.Belt) {
		return ErrInvalidBelt
	}
	if c.StripeCount < 0 {
		return errors.New("stripe count cannot be negative")
	}
	return nil
}

// Proposal represents a coach-proposed promotion awaiting admin approval.
type Proposal struct {
	ID         string
	MemberID   string
	TargetBelt string
	Notes      string
	ProposedBy string // Coach AccountID
	ApprovedBy string // Admin AccountID (empty until decided)
	Status     string // pending, approved, rejected
	CreatedAt  time.Time
	DecidedAt  time.Time
}

// Validate checks if the Proposal has valid data.
// PRE: Proposal struct is populated
// POST: Returns nil if valid, error otherwise
func (p *Proposal) Validate() error {
	if p.MemberID == "" {
		return ErrEmptyMemberID
	}
	if !isValidBelt(p.TargetBelt) {
		return ErrInvalidBelt
	}
	if p.ProposedBy == "" {
		return ErrEmptyProposedBy
	}
	if !isValidProposalStatus(p.Status) {
		return ErrInvalidProposalStatus
	}
	return nil
}

// IsPending returns true if the proposal is awaiting decision.
// INVARIANT: Status field is not mutated
func (p *Proposal) IsPending() bool {
	return p.Status == ProposalPending
}

// Approve moves the proposal to approved status.
// PRE: Proposal is pending, adminID is non-empty
// POST: Status is approved, ApprovedBy and DecidedAt are set
func (p *Proposal) Approve(adminID string) error {
	if !p.IsPending() {
		return ErrAlreadyDecided
	}
	if adminID == "" {
		return errors.New("admin ID is required to approve")
	}
	p.Status = ProposalApproved
	p.ApprovedBy = adminID
	p.DecidedAt = time.Now()
	return nil
}

// Reject moves the proposal to rejected status.
// PRE: Proposal is pending, adminID is non-empty
// POST: Status is rejected, ApprovedBy and DecidedAt are set
func (p *Proposal) Reject(adminID string) error {
	if !p.IsPending() {
		return ErrAlreadyDecided
	}
	if adminID == "" {
		return errors.New("admin ID is required to reject")
	}
	p.Status = ProposalRejected
	p.ApprovedBy = adminID
	p.DecidedAt = time.Now()
	return nil
}

// InferStripe calculates the stripe count a member should have on their current belt
// based on accumulated mat hours and the config for the next belt in progression.
// PRE: config.FlightTimeHours > 0 and config.StripeCount > 0
// POST: Returns 0..StripeCount (capped at StripeCount)
func InferStripe(totalMatHours float64, config Config) int {
	if config.FlightTimeHours <= 0 || config.StripeCount <= 0 {
		return 0
	}
	hoursPerStripe := config.FlightTimeHours / float64(config.StripeCount)
	if hoursPerStripe <= 0 {
		return 0
	}
	stripe := int(totalMatHours / hoursPerStripe)
	if stripe > config.StripeCount {
		stripe = config.StripeCount
	}
	return stripe
}

func isValidBelt(belt string) bool {
	all := []string{BeltWhite, BeltBlue, BeltPurple, BeltBrown, BeltBlack, BeltGrey, BeltYellow, BeltOrange, BeltGreen}
	for _, b := range all {
		if b == belt {
			return true
		}
	}
	return false
}

func isValidProposalStatus(s string) bool {
	for _, v := range []string{ProposalPending, ProposalApproved, ProposalRejected} {
		if v == s {
			return true
		}
	}
	return false
}
