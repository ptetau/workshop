package clip

import (
	"testing"
	"time"
)

// TestClip_Validate_Valid tests that a valid clip passes validation.
func TestClip_Validate_Valid(t *testing.T) {
	c := Clip{
		ThemeID:      "theme-1",
		Title:        "Leg Lasso Entry",
		YouTubeURL:   "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		StartSeconds: 30,
		EndSeconds:   45,
		CreatedAt:    time.Now(),
	}
	if err := c.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestClip_Validate_EmptyTitle tests that an empty title fails validation.
func TestClip_Validate_EmptyTitle(t *testing.T) {
	c := Clip{ThemeID: "t1", Title: "", YouTubeURL: "https://youtu.be/abc", StartSeconds: 0, EndSeconds: 10}
	if err := c.Validate(); err == nil {
		t.Error("expected error for empty title")
	}
}

// TestClip_Validate_StartAfterEnd tests that start >= end fails validation.
func TestClip_Validate_StartAfterEnd(t *testing.T) {
	c := Clip{ThemeID: "t1", Title: "Test", YouTubeURL: "https://youtu.be/abc", StartSeconds: 30, EndSeconds: 10}
	if err := c.Validate(); err == nil {
		t.Error("expected error for start >= end")
	}
}

// TestClip_ExtractYouTubeID_Standard tests extraction from standard YouTube URL.
func TestClip_ExtractYouTubeID_Standard(t *testing.T) {
	c := Clip{YouTubeURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}
	if err := c.ExtractYouTubeID(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.YouTubeID != "dQw4w9WgXcQ" {
		t.Errorf("got %q, want dQw4w9WgXcQ", c.YouTubeID)
	}
}

// TestClip_ExtractYouTubeID_Short tests extraction from short YouTube URL.
func TestClip_ExtractYouTubeID_Short(t *testing.T) {
	c := Clip{YouTubeURL: "https://youtu.be/dQw4w9WgXcQ"}
	if err := c.ExtractYouTubeID(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.YouTubeID != "dQw4w9WgXcQ" {
		t.Errorf("got %q, want dQw4w9WgXcQ", c.YouTubeID)
	}
}

// TestClip_ExtractYouTubeID_Invalid tests that an invalid URL returns error.
func TestClip_ExtractYouTubeID_Invalid(t *testing.T) {
	c := Clip{YouTubeURL: "https://example.com/video"}
	if err := c.ExtractYouTubeID(); err == nil {
		t.Error("expected error for invalid YouTube URL")
	}
}

// TestClip_Promote tests promoting a clip.
func TestClip_Promote(t *testing.T) {
	c := Clip{Promoted: false}
	if err := c.Promote("coach-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.Promoted {
		t.Error("expected Promoted to be true")
	}
	if c.PromotedBy != "coach-1" {
		t.Errorf("got PromotedBy %q, want coach-1", c.PromotedBy)
	}
}

// TestClip_Promote_AlreadyPromoted tests that promoting twice fails.
func TestClip_Promote_AlreadyPromoted(t *testing.T) {
	c := Clip{Promoted: true, PromotedBy: "coach-1"}
	if err := c.Promote("admin-1"); err == nil {
		t.Error("expected error for already promoted clip")
	}
}

// TestClip_DurationSeconds tests duration calculation.
func TestClip_DurationSeconds(t *testing.T) {
	c := Clip{StartSeconds: 30, EndSeconds: 45}
	if d := c.DurationSeconds(); d != 15 {
		t.Errorf("got %d, want 15", d)
	}
}
