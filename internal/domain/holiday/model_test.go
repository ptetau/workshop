package holiday_test

import (
	"testing"
	"time"

	"workshop/internal/domain/holiday"
)

// TestHoliday_Validate tests validation of Holiday.
func TestHoliday_Validate(t *testing.T) {
	start := time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		hol     holiday.Holiday
		wantErr bool
	}{
		{
			name:    "valid multi-day holiday",
			hol:     holiday.Holiday{ID: "1", Name: "Easter Break", StartDate: start, EndDate: end},
			wantErr: false,
		},
		{
			name:    "valid single-day holiday",
			hol:     holiday.Holiday{ID: "2", Name: "ANZAC Day", StartDate: start, EndDate: start},
			wantErr: false,
		},
		{
			name:    "empty name",
			hol:     holiday.Holiday{ID: "3", Name: "", StartDate: start, EndDate: end},
			wantErr: true,
		},
		{
			name:    "zero start date",
			hol:     holiday.Holiday{ID: "4", Name: "Test", StartDate: time.Time{}, EndDate: end},
			wantErr: true,
		},
		{
			name:    "zero end date",
			hol:     holiday.Holiday{ID: "5", Name: "Test", StartDate: start, EndDate: time.Time{}},
			wantErr: true,
		},
		{
			name:    "start after end",
			hol:     holiday.Holiday{ID: "6", Name: "Test", StartDate: end, EndDate: start},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.hol.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Holiday.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHoliday_Contains tests the Contains method on Holiday.
func TestHoliday_Contains(t *testing.T) {
	hol := holiday.Holiday{
		ID:        "1",
		Name:      "Easter Break",
		StartDate: time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name string
		date time.Time
		want bool
	}{
		{"before holiday", time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC), false},
		{"first day", time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC), true},
		{"middle day", time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC), true},
		{"last day", time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC), true},
		{"after holiday", time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hol.Contains(tt.date); got != tt.want {
				t.Errorf("Holiday.Contains(%v) = %v, want %v", tt.date, got, tt.want)
			}
		})
	}
}
