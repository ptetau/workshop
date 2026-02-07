package kiosk_test

import (
	"testing"
	"time"

	"workshop/internal/domain/kiosk"
)

// TestSession_Validate tests validation of kiosk Session.
func TestSession_Validate(t *testing.T) {
	tests := []struct {
		name    string
		session kiosk.Session
		wantErr bool
	}{
		{
			name:    "valid session",
			session: kiosk.Session{ID: "1", AccountID: "acct-1", StartedAt: time.Now()},
			wantErr: false,
		},
		{
			name:    "empty account ID",
			session: kiosk.Session{ID: "2", AccountID: "", StartedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "zero started_at",
			session: kiosk.Session{ID: "3", AccountID: "acct-1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSession_IsActive tests the IsActive method on kiosk Session.
func TestSession_IsActive(t *testing.T) {
	t.Run("active session", func(t *testing.T) {
		s := kiosk.Session{ID: "1", AccountID: "acct-1", StartedAt: time.Now()}
		if !s.IsActive() {
			t.Error("expected active session")
		}
	})

	t.Run("ended session", func(t *testing.T) {
		s := kiosk.Session{ID: "2", AccountID: "acct-1", StartedAt: time.Now(), EndedAt: time.Now()}
		if s.IsActive() {
			t.Error("expected inactive session")
		}
	})
}

// TestSession_End tests the End method on kiosk Session.
func TestSession_End(t *testing.T) {
	t.Run("end active session", func(t *testing.T) {
		s := kiosk.Session{ID: "1", AccountID: "acct-1", StartedAt: time.Now()}
		if err := s.End(); err != nil {
			t.Errorf("End() unexpected error: %v", err)
		}
		if s.IsActive() {
			t.Error("session should be ended")
		}
	})

	t.Run("end already ended session", func(t *testing.T) {
		s := kiosk.Session{ID: "2", AccountID: "acct-1", StartedAt: time.Now(), EndedAt: time.Now()}
		err := s.End()
		if err == nil {
			t.Error("End() should fail on already ended session")
		}
	})
}
