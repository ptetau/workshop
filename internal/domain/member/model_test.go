package member_test

import (
	"testing"
	"workshop/internal/domain/member"
)

// TestMemberValidation tests validation of Member.
func TestMemberValidation(t *testing.T) {
	tests := []struct {
		name    string
		member  member.Member
		wantErr bool
	}{
		{
			name: "valid member",
			member: member.Member{
				ID:      "123",
				Name:    "John Doe",
				Email:   "john@example.com",
				Program: member.ProgramAdults,
				Status:  member.StatusActive,
				Fee:     100,
			},
			wantErr: false,
		},
		{
			name: "valid archived member",
			member: member.Member{
				ID:      "123",
				Name:    "John Doe",
				Email:   "john@example.com",
				Program: member.ProgramAdults,
				Status:  member.StatusArchived,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			member: member.Member{
				ID:      "123",
				Name:    "",
				Email:   "john@example.com",
				Program: member.ProgramAdults,
				Status:  member.StatusActive,
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			member: member.Member{
				ID:      "123",
				Name:    "John Doe",
				Email:   "invalid-email",
				Program: member.ProgramAdults,
				Status:  member.StatusActive,
			},
			wantErr: true,
		},
		{
			name: "invalid program",
			member: member.Member{
				ID:      "123",
				Name:    "John Doe",
				Email:   "john@example.com",
				Program: "invalid",
				Status:  member.StatusActive,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			member: member.Member{
				ID:      "123",
				Name:    "John Doe",
				Email:   "john@example.com",
				Program: member.ProgramAdults,
				Status:  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.member.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Member.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMemberIsActive tests the IsActive method on Member.
func TestMemberIsActive(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"active member", member.StatusActive, true},
		{"inactive member", member.StatusInactive, false},
		{"archived member", member.StatusArchived, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := member.Member{Status: tt.status}
			if got := m.IsActive(); got != tt.want {
				t.Errorf("Member.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMemberArchive tests the Archive method on Member.
func TestMemberArchive(t *testing.T) {
	t.Run("archive active member", func(t *testing.T) {
		m := member.Member{Status: member.StatusActive}
		if err := m.Archive(); err != nil {
			t.Errorf("Archive() unexpected error: %v", err)
		}
		if m.Status != member.StatusArchived {
			t.Errorf("Status = %v, want %v", m.Status, member.StatusArchived)
		}
	})

	t.Run("archive inactive member", func(t *testing.T) {
		m := member.Member{Status: member.StatusInactive}
		if err := m.Archive(); err != nil {
			t.Errorf("Archive() unexpected error: %v", err)
		}
		if m.Status != member.StatusArchived {
			t.Errorf("Status = %v, want %v", m.Status, member.StatusArchived)
		}
	})

	t.Run("archive already archived member", func(t *testing.T) {
		m := member.Member{Status: member.StatusArchived}
		err := m.Archive()
		if err == nil {
			t.Error("Archive() should fail on already archived member")
		}
	})
}

// TestMemberRestore tests the Restore method on Member.
func TestMemberRestore(t *testing.T) {
	t.Run("restore archived member", func(t *testing.T) {
		m := member.Member{Status: member.StatusArchived}
		if err := m.Restore(); err != nil {
			t.Errorf("Restore() unexpected error: %v", err)
		}
		if m.Status != member.StatusActive {
			t.Errorf("Status = %v, want %v", m.Status, member.StatusActive)
		}
	})

	t.Run("restore active member fails", func(t *testing.T) {
		m := member.Member{Status: member.StatusActive}
		err := m.Restore()
		if err == nil {
			t.Error("Restore() should fail on non-archived member")
		}
	})

	t.Run("restore inactive member fails", func(t *testing.T) {
		m := member.Member{Status: member.StatusInactive}
		err := m.Restore()
		if err == nil {
			t.Error("Restore() should fail on non-archived member")
		}
	})
}

// TestMemberIsArchived tests the IsArchived method on Member.
func TestMemberIsArchived(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"active", member.StatusActive, false},
		{"inactive", member.StatusInactive, false},
		{"archived", member.StatusArchived, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := member.Member{Status: tt.status}
			if got := m.IsArchived(); got != tt.want {
				t.Errorf("IsArchived() = %v, want %v", got, tt.want)
			}
		})
	}
}
