package calendar

import (
	"testing"
	"time"
)

// TestEvent_Validate tests Event validation rules.
func TestEvent_Validate(t *testing.T) {
	valid := Event{
		ID:        "e1",
		Title:     "End of Year BBQ",
		Type:      TypeEvent,
		StartDate: time.Now(),
		CreatedBy: "admin1",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid event, got: %v", err)
	}

	tests := []struct {
		name    string
		modify  func(e *Event)
		wantErr string
	}{
		{"empty title", func(e *Event) { e.Title = "" }, "title cannot be empty"},
		{"title too long", func(e *Event) { e.Title = string(make([]byte, MaxTitleLength+1)) }, "title cannot exceed"},
		{"invalid type", func(e *Event) { e.Type = "party" }, "type must be"},
		{"missing start date", func(e *Event) { e.StartDate = time.Time{} }, "start date is required"},
		{"end before start", func(e *Event) {
			e.StartDate = time.Now()
			e.EndDate = e.StartDate.Add(-time.Hour)
		}, "end date cannot be before"},
		{"description too long", func(e *Event) { e.Description = string(make([]byte, MaxDescriptionLength+1)) }, "description cannot exceed"},
		{"location too long", func(e *Event) { e.Location = string(make([]byte, MaxLocationLength+1)) }, "location cannot exceed"},
		{"url too long", func(e *Event) { e.RegistrationURL = string(make([]byte, MaxRegistrationURLLength+1)) }, "URL cannot exceed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := valid
			tc.modify(&e)
			err := e.Validate()
			if err == nil {
				t.Fatal("expected error")
			}
			if !contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got: %v", tc.wantErr, err)
			}
		})
	}
}

// TestEvent_Validate_Competition tests valid competition event.
func TestEvent_Validate_Competition(t *testing.T) {
	e := Event{
		ID:              "c1",
		Title:           "Grappling Industries",
		Type:            TypeCompetition,
		StartDate:       time.Now(),
		RegistrationURL: "https://grapplingindustries.com/register",
		CreatedBy:       "admin1",
	}
	if err := e.Validate(); err != nil {
		t.Fatalf("expected valid competition, got: %v", err)
	}
}

// TestEvent_IsMultiDay tests single-day vs multi-day detection.
func TestEvent_IsMultiDay(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)

	single := Event{StartDate: now}
	if single.IsMultiDay() {
		t.Fatal("single day event should not be multi-day")
	}

	sameDay := Event{StartDate: now, EndDate: now.Add(6 * time.Hour)}
	if sameDay.IsMultiDay() {
		t.Fatal("same-day event should not be multi-day")
	}

	multi := Event{StartDate: now, EndDate: now.Add(48 * time.Hour)}
	if !multi.IsMultiDay() {
		t.Fatal("multi-day event should be multi-day")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchSubstring(s, sub)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
