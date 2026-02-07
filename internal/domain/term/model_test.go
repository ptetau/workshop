package term_test

import (
	"testing"
	"time"

	"workshop/internal/domain/term"
)

// TestTerm_Validate tests validation of Term.
func TestTerm_Validate(t *testing.T) {
	start := time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		term    term.Term
		wantErr bool
	}{
		{
			name:    "valid term",
			term:    term.Term{ID: "1", Name: "Term 1 2026", StartDate: start, EndDate: end},
			wantErr: false,
		},
		{
			name:    "empty name",
			term:    term.Term{ID: "2", Name: "", StartDate: start, EndDate: end},
			wantErr: true,
		},
		{
			name:    "zero start date",
			term:    term.Term{ID: "3", Name: "Term 1", StartDate: time.Time{}, EndDate: end},
			wantErr: true,
		},
		{
			name:    "zero end date",
			term:    term.Term{ID: "4", Name: "Term 1", StartDate: start, EndDate: time.Time{}},
			wantErr: true,
		},
		{
			name:    "start after end",
			term:    term.Term{ID: "5", Name: "Term 1", StartDate: end, EndDate: start},
			wantErr: true,
		},
		{
			name:    "start equals end",
			term:    term.Term{ID: "6", Name: "Term 1", StartDate: start, EndDate: start},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.term.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Term.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTerm_Contains tests the Contains method on Term.
func TestTerm_Contains(t *testing.T) {
	tm := term.Term{
		ID:        "1",
		Name:      "Term 1 2026",
		StartDate: time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name string
		date time.Time
		want bool
	}{
		{"before term", time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC), false},
		{"first day", time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC), true},
		{"middle of term", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), true},
		{"last day", time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC), true},
		{"after term", time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tm.Contains(tt.date); got != tt.want {
				t.Errorf("Term.Contains(%v) = %v, want %v", tt.date, got, tt.want)
			}
		})
	}
}
