package notice_test

import (
	"testing"
	"time"

	"workshop/internal/domain/notice"
)

// TestNotice_Validate tests validation of Notice.
func TestNotice_Validate(t *testing.T) {
	tests := []struct {
		name    string
		notice  notice.Notice
		wantErr bool
	}{
		{
			name: "valid school_wide notice",
			notice: notice.Notice{
				ID: "1", Type: notice.TypeSchoolWide, Status: notice.StatusDraft,
				Title: "Gym closed Friday", Content: "No classes this Friday.", CreatedBy: "acct-1", CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid class_specific notice",
			notice: notice.Notice{
				ID: "2", Type: notice.TypeClassSpecific, Status: notice.StatusDraft,
				Title: "Grading reminder", Content: "Remind students about Saturday grading.", CreatedBy: "acct-1", TargetID: "ct-1", CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid holiday notice",
			notice: notice.Notice{
				ID: "3", Type: notice.TypeHoliday, Status: notice.StatusPublished,
				Title: "Summer Break", Content: "Gym closed 23 Dec – 7 Jan.", CreatedBy: "system", TargetID: "hol-1", CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:    "empty title",
			notice:  notice.Notice{ID: "4", Type: notice.TypeSchoolWide, Status: notice.StatusDraft, Content: "content"},
			wantErr: true,
		},
		{
			name:    "empty content",
			notice:  notice.Notice{ID: "5", Type: notice.TypeSchoolWide, Status: notice.StatusDraft, Title: "title"},
			wantErr: true,
		},
		{
			name:    "invalid type",
			notice:  notice.Notice{ID: "6", Type: "invalid", Status: notice.StatusDraft, Title: "t", Content: "c"},
			wantErr: true,
		},
		{
			name:    "invalid status",
			notice:  notice.Notice{ID: "7", Type: notice.TypeSchoolWide, Status: "bogus", Title: "t", Content: "c"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.notice.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNotice_Publish tests the Publish method on Notice.
func TestNotice_Publish(t *testing.T) {
	t.Run("publish draft notice", func(t *testing.T) {
		n := notice.Notice{
			ID: "1", Type: notice.TypeSchoolWide, Status: notice.StatusDraft,
			Title: "Test", Content: "Test content", CreatedBy: "acct-1",
		}
		if err := n.Publish("acct-admin"); err != nil {
			t.Errorf("Publish() unexpected error: %v", err)
		}
		if !n.IsPublished() {
			t.Error("expected notice to be published")
		}
		if n.PublishedBy != "acct-admin" {
			t.Errorf("expected PublishedBy=acct-admin, got %s", n.PublishedBy)
		}
		if n.PublishedAt.IsZero() {
			t.Error("expected PublishedAt to be set")
		}
	})

	t.Run("publish already published notice", func(t *testing.T) {
		n := notice.Notice{
			ID: "2", Type: notice.TypeSchoolWide, Status: notice.StatusPublished,
			Title: "Test", Content: "content",
		}
		if err := n.Publish("acct-admin"); err == nil {
			t.Error("expected error when publishing already published notice")
		}
	})

	t.Run("publish with empty publisher", func(t *testing.T) {
		n := notice.Notice{
			ID: "3", Type: notice.TypeSchoolWide, Status: notice.StatusDraft,
			Title: "Test", Content: "content",
		}
		if err := n.Publish(""); err == nil {
			t.Error("expected error when publisher ID is empty")
		}
	})
}

// TestNotice_IsDraft_IsPublished tests the IsDraft and IsPublished methods.
func TestNotice_IsDraft_IsPublished(t *testing.T) {
	draft := notice.Notice{Status: notice.StatusDraft}
	published := notice.Notice{Status: notice.StatusPublished}

	if !draft.IsDraft() {
		t.Error("expected IsDraft=true for draft notice")
	}
	if draft.IsPublished() {
		t.Error("expected IsPublished=false for draft notice")
	}
	if published.IsDraft() {
		t.Error("expected IsDraft=false for published notice")
	}
	if !published.IsPublished() {
		t.Error("expected IsPublished=true for published notice")
	}
}
