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

// TestNotice_Validate_Color tests color validation.
func TestNotice_Validate_Color(t *testing.T) {
	base := notice.Notice{
		ID: "1", Type: notice.TypeSchoolWide, Status: notice.StatusDraft,
		Title: "T", Content: "C", CreatedBy: "acct-1",
	}

	t.Run("valid color", func(t *testing.T) {
		n := base
		n.Color = notice.ColorBlue
		if err := n.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("empty color is valid (defaults to orange)", func(t *testing.T) {
		n := base
		n.Color = ""
		if err := n.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("invalid color", func(t *testing.T) {
		n := base
		n.Color = "neon_pink"
		if err := n.Validate(); err != notice.ErrInvalidColor {
			t.Errorf("expected ErrInvalidColor, got %v", err)
		}
	})
}

// TestNotice_Pin_Unpin tests the Pin and Unpin methods.
func TestNotice_Pin_Unpin(t *testing.T) {
	now := time.Now()

	t.Run("pin unpinned notice", func(t *testing.T) {
		n := notice.Notice{Pinned: false}
		if err := n.Pin(now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !n.Pinned {
			t.Error("expected Pinned=true")
		}
		if n.PinnedAt.IsZero() {
			t.Error("expected PinnedAt to be set")
		}
	})
	t.Run("pin already pinned notice", func(t *testing.T) {
		n := notice.Notice{Pinned: true, PinnedAt: now}
		if err := n.Pin(now); err != notice.ErrAlreadyPinned {
			t.Errorf("expected ErrAlreadyPinned, got %v", err)
		}
	})
	t.Run("unpin pinned notice", func(t *testing.T) {
		n := notice.Notice{Pinned: true, PinnedAt: now}
		if err := n.Unpin(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n.Pinned {
			t.Error("expected Pinned=false")
		}
		if !n.PinnedAt.IsZero() {
			t.Error("expected PinnedAt to be zero")
		}
	})
	t.Run("unpin not pinned notice", func(t *testing.T) {
		n := notice.Notice{Pinned: false}
		if err := n.Unpin(); err != notice.ErrNotPinned {
			t.Errorf("expected ErrNotPinned, got %v", err)
		}
	})
}

// TestNotice_IsVisible tests the visibility window logic.
func TestNotice_IsVisible(t *testing.T) {
	now := time.Date(2026, 2, 8, 12, 0, 0, 0, time.UTC)

	t.Run("no window set — always visible", func(t *testing.T) {
		n := notice.Notice{}
		if !n.IsVisible(now) {
			t.Error("expected visible when no window set")
		}
	})
	t.Run("visible_from in past — visible", func(t *testing.T) {
		n := notice.Notice{VisibleFrom: now.Add(-time.Hour)}
		if !n.IsVisible(now) {
			t.Error("expected visible")
		}
	})
	t.Run("visible_from in future — not visible", func(t *testing.T) {
		n := notice.Notice{VisibleFrom: now.Add(time.Hour)}
		if n.IsVisible(now) {
			t.Error("expected not visible")
		}
	})
	t.Run("visible_until in future — visible", func(t *testing.T) {
		n := notice.Notice{VisibleUntil: now.Add(time.Hour)}
		if !n.IsVisible(now) {
			t.Error("expected visible")
		}
	})
	t.Run("visible_until in past — not visible", func(t *testing.T) {
		n := notice.Notice{VisibleUntil: now.Add(-time.Hour)}
		if n.IsVisible(now) {
			t.Error("expected not visible")
		}
	})
	t.Run("within window", func(t *testing.T) {
		n := notice.Notice{VisibleFrom: now.Add(-time.Hour), VisibleUntil: now.Add(time.Hour)}
		if !n.IsVisible(now) {
			t.Error("expected visible within window")
		}
	})
}

// TestNotice_EffectiveColor tests the EffectiveColor method.
func TestNotice_EffectiveColor(t *testing.T) {
	t.Run("empty defaults to orange hex", func(t *testing.T) {
		n := notice.Notice{Color: ""}
		if n.EffectiveColor() != "#F9B232" {
			t.Errorf("expected #F9B232, got %s", n.EffectiveColor())
		}
	})
	t.Run("blue returns blue hex", func(t *testing.T) {
		n := notice.Notice{Color: notice.ColorBlue}
		if n.EffectiveColor() != "#2980b9" {
			t.Errorf("expected #2980b9, got %s", n.EffectiveColor())
		}
	})
	t.Run("invalid color falls back to orange", func(t *testing.T) {
		n := notice.Notice{Color: "neon"}
		if n.EffectiveColor() != "#F9B232" {
			t.Errorf("expected #F9B232, got %s", n.EffectiveColor())
		}
	})
}
