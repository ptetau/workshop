package email

import (
	"testing"
	"time"
)

var fixedTime = time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

// TestEmail_Validate_Valid tests that a well-formed email passes validation.
func TestEmail_Validate_Valid(t *testing.T) {
	e := Email{
		ID:        "email-1",
		Subject:   "Schedule Change",
		Body:      "The Monday class is moving to Tuesday.",
		SenderID:  "admin-1",
		Status:    StatusDraft,
		CreatedAt: fixedTime,
	}
	if err := e.Validate(); err != nil {
		t.Errorf("expected valid email, got: %v", err)
	}
}

// TestEmail_Validate_MissingSubject tests that empty subject is rejected.
func TestEmail_Validate_MissingSubject(t *testing.T) {
	e := Email{Body: "body", SenderID: "a", CreatedAt: fixedTime}
	if err := e.Validate(); err != ErrEmptySubject {
		t.Errorf("expected ErrEmptySubject, got: %v", err)
	}
}

// TestEmail_Validate_MissingBody tests that empty body is rejected.
func TestEmail_Validate_MissingBody(t *testing.T) {
	e := Email{Subject: "sub", SenderID: "a", CreatedAt: fixedTime}
	if err := e.Validate(); err != ErrEmptyBody {
		t.Errorf("expected ErrEmptyBody, got: %v", err)
	}
}

// TestEmail_Validate_MissingSender tests that empty sender is rejected.
func TestEmail_Validate_MissingSender(t *testing.T) {
	e := Email{Subject: "sub", Body: "body", CreatedAt: fixedTime}
	if err := e.Validate(); err != ErrEmptySenderID {
		t.Errorf("expected ErrEmptySenderID, got: %v", err)
	}
}

// TestEmail_MarkQueued_FromDraft tests draft→queued transition.
func TestEmail_MarkQueued_FromDraft(t *testing.T) {
	e := Email{Status: StatusDraft}
	if err := e.MarkQueued(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if e.Status != StatusQueued {
		t.Errorf("expected queued, got %s", e.Status)
	}
}

// TestEmail_MarkQueued_FromFailed tests failed→queued retry transition.
func TestEmail_MarkQueued_FromFailed(t *testing.T) {
	e := Email{Status: StatusFailed}
	if err := e.MarkQueued(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if e.Status != StatusQueued {
		t.Errorf("expected queued, got %s", e.Status)
	}
}

// TestEmail_MarkQueued_FromSent tests that sent emails cannot be queued.
func TestEmail_MarkQueued_FromSent(t *testing.T) {
	e := Email{Status: StatusSent}
	if err := e.MarkQueued(); err != ErrNotDraft {
		t.Errorf("expected ErrNotDraft, got: %v", err)
	}
}

// TestEmail_MarkSent tests the sent transition.
func TestEmail_MarkSent(t *testing.T) {
	e := Email{Status: StatusQueued}
	now := time.Now()
	e.MarkSent(now, "resend-123")
	if e.Status != StatusSent {
		t.Errorf("expected sent, got %s", e.Status)
	}
	if e.ResendMessageID != "resend-123" {
		t.Errorf("expected resend-123, got %s", e.ResendMessageID)
	}
	if e.SentAt.IsZero() {
		t.Error("expected SentAt to be set")
	}
}

// TestEmail_Schedule_FromDraft tests scheduling a draft.
func TestEmail_Schedule_FromDraft(t *testing.T) {
	e := Email{Status: StatusDraft}
	future := time.Now().Add(24 * time.Hour)
	if err := e.Schedule(future); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if e.Status != StatusScheduled {
		t.Errorf("expected scheduled, got %s", e.Status)
	}
}

// TestEmail_Schedule_FromSent tests that sent emails cannot be scheduled.
func TestEmail_Schedule_FromSent(t *testing.T) {
	e := Email{Status: StatusSent}
	if err := e.Schedule(time.Now()); err != ErrNotDraft {
		t.Errorf("expected ErrNotDraft, got: %v", err)
	}
}

// TestEmail_Reschedule tests rescheduling a scheduled email.
func TestEmail_Reschedule(t *testing.T) {
	e := Email{Status: StatusScheduled, ScheduledAt: time.Now()}
	newTime := time.Now().Add(48 * time.Hour)
	if err := e.Reschedule(newTime); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !e.ScheduledAt.Equal(newTime) {
		t.Errorf("expected updated schedule time")
	}
}

// TestEmail_Reschedule_NotScheduled tests that non-scheduled emails can't be rescheduled.
func TestEmail_Reschedule_NotScheduled(t *testing.T) {
	e := Email{Status: StatusDraft}
	if err := e.Reschedule(time.Now()); err != ErrNotScheduled {
		t.Errorf("expected ErrNotScheduled, got: %v", err)
	}
}

// TestEmail_Cancel_Scheduled tests cancelling a scheduled email.
func TestEmail_Cancel_Scheduled(t *testing.T) {
	e := Email{Status: StatusScheduled}
	if err := e.Cancel(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if e.Status != StatusCancelled {
		t.Errorf("expected cancelled, got %s", e.Status)
	}
}

// TestEmail_Cancel_Draft tests cancelling a draft email.
func TestEmail_Cancel_Draft(t *testing.T) {
	e := Email{Status: StatusDraft}
	if err := e.Cancel(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if e.Status != StatusCancelled {
		t.Errorf("expected cancelled, got %s", e.Status)
	}
}

// TestEmail_Cancel_Sent tests that sent emails cannot be cancelled.
func TestEmail_Cancel_Sent(t *testing.T) {
	e := Email{Status: StatusSent}
	if err := e.Cancel(); err != ErrNotCancellable {
		t.Errorf("expected ErrNotCancellable, got: %v", err)
	}
}

// TestEmail_StatusChecks tests the boolean status helpers.
func TestEmail_StatusChecks(t *testing.T) {
	cases := []struct {
		status      string
		isDraft     bool
		isSent      bool
		isScheduled bool
	}{
		{StatusDraft, true, false, false},
		{StatusSent, false, true, false},
		{StatusScheduled, false, false, true},
		{StatusQueued, false, false, false},
		{StatusCancelled, false, false, false},
	}
	for _, tc := range cases {
		e := Email{Status: tc.status}
		if e.IsDraft() != tc.isDraft {
			t.Errorf("status=%s: IsDraft()=%v, want %v", tc.status, e.IsDraft(), tc.isDraft)
		}
		if e.IsSent() != tc.isSent {
			t.Errorf("status=%s: IsSent()=%v, want %v", tc.status, e.IsSent(), tc.isSent)
		}
		if e.IsScheduled() != tc.isScheduled {
			t.Errorf("status=%s: IsScheduled()=%v, want %v", tc.status, e.IsScheduled(), tc.isScheduled)
		}
	}
}

// TestEmailTemplate_WrapBody tests that WrapBody correctly wraps content with header and footer.
func TestEmailTemplate_WrapBody(t *testing.T) {
	tpl := EmailTemplate{
		Header: "<div>Header</div>",
		Footer: "<div>Footer</div>",
	}
	result := tpl.WrapBody("<p>Body</p>")
	expected := "<div>Header</div><p>Body</p><div>Footer</div>"
	if result != expected {
		t.Errorf("WrapBody = %q, want %q", result, expected)
	}
}

// TestEmailTemplate_WrapBody_Empty tests WrapBody with empty header/footer.
func TestEmailTemplate_WrapBody_Empty(t *testing.T) {
	tpl := EmailTemplate{}
	result := tpl.WrapBody("<p>Body</p>")
	if result != "<p>Body</p>" {
		t.Errorf("WrapBody with empty template = %q, want %q", result, "<p>Body</p>")
	}
}
