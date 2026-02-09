package grading_test

import (
	"testing"
	"time"

	"workshop/internal/domain/grading"
)

// TestRecord_Validate tests validation of grading Record.
func TestRecord_Validate(t *testing.T) {
	tests := []struct {
		name    string
		record  grading.Record
		wantErr bool
	}{
		{
			name:    "valid record",
			record:  grading.Record{ID: "1", MemberID: "m1", Belt: grading.BeltBlue, Stripe: 0, PromotedAt: time.Now(), Method: grading.MethodStandard},
			wantErr: false,
		},
		{
			name:    "empty member ID",
			record:  grading.Record{ID: "2", Belt: grading.BeltBlue, PromotedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "invalid belt",
			record:  grading.Record{ID: "3", MemberID: "m1", Belt: "rainbow", PromotedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "stripe too high",
			record:  grading.Record{ID: "4", MemberID: "m1", Belt: grading.BeltWhite, Stripe: 5, PromotedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "zero promoted_at",
			record:  grading.Record{ID: "5", MemberID: "m1", Belt: grading.BeltWhite, Stripe: 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Record.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfig_Validate tests validation of grading Config.
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  grading.Config
		wantErr bool
	}{
		{
			name:    "valid adults config",
			config:  grading.Config{ID: "1", Program: "adults", Belt: grading.BeltPurple, FlightTimeHours: 150, StripeCount: 4},
			wantErr: false,
		},
		{
			name:    "valid kids config",
			config:  grading.Config{ID: "2", Program: "kids", Belt: grading.BeltYellow, AttendancePct: 80, StripeCount: 4},
			wantErr: false,
		},
		{
			name:    "empty program",
			config:  grading.Config{ID: "3", Belt: grading.BeltBlue},
			wantErr: true,
		},
		{
			name:    "invalid belt",
			config:  grading.Config{ID: "4", Program: "adults", Belt: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfig_Validate_AllBelts verifies every valid belt is accepted.
func TestConfig_Validate_AllBelts(t *testing.T) {
	allBelts := []string{
		grading.BeltWhite, grading.BeltBlue, grading.BeltPurple, grading.BeltBrown, grading.BeltBlack,
		grading.BeltGrey, grading.BeltYellow, grading.BeltOrange, grading.BeltGreen,
	}
	for _, belt := range allBelts {
		config := grading.Config{ID: "t", Program: "adults", Belt: belt, StripeCount: 4}
		if err := config.Validate(); err != nil {
			t.Errorf("Config.Validate() rejected valid belt %q: %v", belt, err)
		}
	}
}

// TestConfig_Validate_CaseSensitive verifies uppercase belts are rejected (handler normalizes).
func TestConfig_Validate_CaseSensitive(t *testing.T) {
	config := grading.Config{ID: "t", Program: "adults", Belt: "Blue", StripeCount: 4}
	if err := config.Validate(); err == nil {
		t.Error("Config.Validate() should reject uppercase belt 'Blue'")
	}
}

// TestProposal_Validate tests validation of grading Proposal.
func TestProposal_Validate(t *testing.T) {
	tests := []struct {
		name     string
		proposal grading.Proposal
		wantErr  bool
	}{
		{
			name:     "valid proposal",
			proposal: grading.Proposal{ID: "1", MemberID: "m1", TargetBelt: grading.BeltPurple, ProposedBy: "coach1", Status: grading.ProposalPending},
			wantErr:  false,
		},
		{
			name:     "empty member ID",
			proposal: grading.Proposal{ID: "2", TargetBelt: grading.BeltBlue, ProposedBy: "coach1", Status: grading.ProposalPending},
			wantErr:  true,
		},
		{
			name:     "invalid belt",
			proposal: grading.Proposal{ID: "3", MemberID: "m1", TargetBelt: "invalid", ProposedBy: "coach1", Status: grading.ProposalPending},
			wantErr:  true,
		},
		{
			name:     "empty proposed_by",
			proposal: grading.Proposal{ID: "4", MemberID: "m1", TargetBelt: grading.BeltBlue, Status: grading.ProposalPending},
			wantErr:  true,
		},
		{
			name:     "invalid status",
			proposal: grading.Proposal{ID: "5", MemberID: "m1", TargetBelt: grading.BeltBlue, ProposedBy: "coach1", Status: "bogus"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.proposal.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Proposal.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestProposal_Approve tests the Approve method on grading Proposal.
func TestProposal_Approve(t *testing.T) {
	t.Run("approve pending", func(t *testing.T) {
		p := grading.Proposal{ID: "1", MemberID: "m1", TargetBelt: grading.BeltBlue, ProposedBy: "coach1", Status: grading.ProposalPending}
		if err := p.Approve("admin1"); err != nil {
			t.Errorf("Approve() unexpected error: %v", err)
		}
		if p.Status != grading.ProposalApproved {
			t.Errorf("expected status approved, got %s", p.Status)
		}
		if p.ApprovedBy != "admin1" {
			t.Errorf("expected ApprovedBy=admin1, got %s", p.ApprovedBy)
		}
	})

	t.Run("approve already decided", func(t *testing.T) {
		p := grading.Proposal{Status: grading.ProposalApproved}
		if err := p.Approve("admin1"); err == nil {
			t.Error("expected error when approving already decided proposal")
		}
	})
}

// TestProposal_Reject tests the Reject method on grading Proposal.
func TestProposal_Reject(t *testing.T) {
	t.Run("reject pending", func(t *testing.T) {
		p := grading.Proposal{ID: "1", MemberID: "m1", TargetBelt: grading.BeltBlue, ProposedBy: "coach1", Status: grading.ProposalPending}
		if err := p.Reject("admin1"); err != nil {
			t.Errorf("Reject() unexpected error: %v", err)
		}
		if p.Status != grading.ProposalRejected {
			t.Errorf("expected status rejected, got %s", p.Status)
		}
	})

	t.Run("reject already decided", func(t *testing.T) {
		p := grading.Proposal{Status: grading.ProposalRejected}
		if err := p.Reject("admin1"); err == nil {
			t.Error("expected error when rejecting already decided proposal")
		}
	})
}
