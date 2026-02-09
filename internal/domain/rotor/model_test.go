package rotor_test

import (
	"testing"
	"time"

	"workshop/internal/domain/rotor"
)

// TestRotor_Validate tests validation of Rotor.
func TestRotor_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rotor   rotor.Rotor
		wantErr error
	}{
		{
			name: "valid draft rotor",
			rotor: rotor.Rotor{
				ID: "1", ClassTypeID: "ct1", Name: "Gi Express v1",
				Version: 1, Status: rotor.StatusDraft, CreatedBy: "admin1",
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			rotor: rotor.Rotor{
				ID: "1", ClassTypeID: "ct1", Name: "",
				Version: 1, Status: rotor.StatusDraft, CreatedBy: "admin1",
			},
			wantErr: rotor.ErrEmptyName,
		},
		{
			name: "empty class type ID",
			rotor: rotor.Rotor{
				ID: "1", ClassTypeID: "", Name: "Gi Express v1",
				Version: 1, Status: rotor.StatusDraft, CreatedBy: "admin1",
			},
			wantErr: rotor.ErrEmptyClassTypeID,
		},
		{
			name: "empty created by",
			rotor: rotor.Rotor{
				ID: "1", ClassTypeID: "ct1", Name: "Gi Express v1",
				Version: 1, Status: rotor.StatusDraft, CreatedBy: "",
			},
			wantErr: rotor.ErrEmptyCreatedBy,
		},
		{
			name: "invalid status",
			rotor: rotor.Rotor{
				ID: "1", ClassTypeID: "ct1", Name: "Gi Express v1",
				Version: 1, Status: "invalid", CreatedBy: "admin1",
			},
			wantErr: rotor.ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rotor.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestRotor_Activate tests the draft-to-active transition.
func TestRotor_Activate(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	t.Run("draft to active", func(t *testing.T) {
		r := &rotor.Rotor{Status: rotor.StatusDraft}
		if err := r.Activate(now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Status != rotor.StatusActive {
			t.Errorf("Status = %q, want %q", r.Status, rotor.StatusActive)
		}
		if r.ActivatedAt != now {
			t.Errorf("ActivatedAt = %v, want %v", r.ActivatedAt, now)
		}
	})

	t.Run("already active", func(t *testing.T) {
		r := &rotor.Rotor{Status: rotor.StatusActive}
		err := r.Activate(now)
		if err != rotor.ErrAlreadyActive {
			t.Errorf("error = %v, want ErrAlreadyActive", err)
		}
	})

	t.Run("archived cannot activate", func(t *testing.T) {
		r := &rotor.Rotor{Status: rotor.StatusArchived}
		err := r.Activate(now)
		if err != rotor.ErrNotDraft {
			t.Errorf("error = %v, want ErrNotDraft", err)
		}
	})
}

// TestRotor_Archive tests the active-to-archived transition.
func TestRotor_Archive(t *testing.T) {
	t.Run("active to archived", func(t *testing.T) {
		r := &rotor.Rotor{Status: rotor.StatusActive}
		if err := r.Archive(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Status != rotor.StatusArchived {
			t.Errorf("Status = %q, want %q", r.Status, rotor.StatusArchived)
		}
	})

	t.Run("draft cannot archive", func(t *testing.T) {
		r := &rotor.Rotor{Status: rotor.StatusDraft}
		err := r.Archive()
		if err != rotor.ErrCannotArchive {
			t.Errorf("error = %v, want ErrCannotArchive", err)
		}
	})
}

// TestRotorTheme_Validate tests validation of RotorTheme.
func TestRotorTheme_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		th := &rotor.RotorTheme{ID: "1", RotorID: "r1", Name: "Standing", Position: 0}
		if err := th.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("empty name", func(t *testing.T) {
		th := &rotor.RotorTheme{ID: "1", RotorID: "r1", Name: ""}
		if err := th.Validate(); err != rotor.ErrEmptyThemeName {
			t.Errorf("error = %v, want ErrEmptyThemeName", err)
		}
	})
	t.Run("empty rotor ID", func(t *testing.T) {
		th := &rotor.RotorTheme{ID: "1", RotorID: "", Name: "Standing"}
		if err := th.Validate(); err != rotor.ErrEmptyRotorID {
			t.Errorf("error = %v, want ErrEmptyRotorID", err)
		}
	})
}

// TestTopic_Validate tests validation of Topic.
func TestTopic_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tp := &rotor.Topic{ID: "1", RotorThemeID: "th1", Name: "Single Leg", DurationWeeks: 1}
		if err := tp.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("empty name", func(t *testing.T) {
		tp := &rotor.Topic{ID: "1", RotorThemeID: "th1", Name: "", DurationWeeks: 1}
		if err := tp.Validate(); err != rotor.ErrEmptyTopicName {
			t.Errorf("error = %v, want ErrEmptyTopicName", err)
		}
	})
	t.Run("empty theme ID", func(t *testing.T) {
		tp := &rotor.Topic{ID: "1", RotorThemeID: "", Name: "Single Leg", DurationWeeks: 1}
		if err := tp.Validate(); err != rotor.ErrEmptyRotorThemeID {
			t.Errorf("error = %v, want ErrEmptyRotorThemeID", err)
		}
	})
	t.Run("zero duration", func(t *testing.T) {
		tp := &rotor.Topic{ID: "1", RotorThemeID: "th1", Name: "Single Leg", DurationWeeks: 0}
		if err := tp.Validate(); err != rotor.ErrInvalidDuration {
			t.Errorf("error = %v, want ErrInvalidDuration", err)
		}
	})
}

// TestTopicSchedule_IsActive tests the schedule active check.
func TestTopicSchedule_IsActive(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	start := time.Date(2025, 6, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 6, 21, 0, 0, 0, 0, time.UTC)

	t.Run("active within range", func(t *testing.T) {
		s := &rotor.TopicSchedule{Status: rotor.ScheduleStatusActive, StartDate: start, EndDate: end}
		if !s.IsActive(now) {
			t.Error("expected IsActive to be true")
		}
	})
	t.Run("not active - wrong status", func(t *testing.T) {
		s := &rotor.TopicSchedule{Status: rotor.ScheduleStatusCompleted, StartDate: start, EndDate: end}
		if s.IsActive(now) {
			t.Error("expected IsActive to be false for completed")
		}
	})
	t.Run("not active - after end", func(t *testing.T) {
		s := &rotor.TopicSchedule{Status: rotor.ScheduleStatusActive, StartDate: start, EndDate: end}
		if s.IsActive(end.Add(24 * time.Hour)) {
			t.Error("expected IsActive to be false after end date")
		}
	})
}
