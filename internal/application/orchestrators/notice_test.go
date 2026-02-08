package orchestrators

import (
	"context"
	"errors"
	"testing"
	"time"

	"workshop/internal/domain/notice"
)

// mockNoticeStoreForOrch implements NoticeStoreForOrchestrator for testing.
type mockNoticeStoreForOrch struct {
	notices map[string]notice.Notice
}

// GetByID implements NoticeStoreForOrchestrator.
// PRE: id is non-empty
// POST: returns notice or error
func (m *mockNoticeStoreForOrch) GetByID(_ context.Context, id string) (notice.Notice, error) {
	n, ok := m.notices[id]
	if !ok {
		return notice.Notice{}, errors.New("not found")
	}
	return n, nil
}

// Save implements NoticeStoreForOrchestrator.
// PRE: notice is valid
// POST: notice is persisted
func (m *mockNoticeStoreForOrch) Save(_ context.Context, n notice.Notice) error {
	m.notices[n.ID] = n
	return nil
}

func newMockNoticeStore() *mockNoticeStoreForOrch {
	return &mockNoticeStoreForOrch{notices: make(map[string]notice.Notice)}
}

var fixedTime = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

func fixedNow() time.Time { return fixedTime }

func fixedID() string { return "test-id-001" }

// --- ExecuteCreateNotice tests ---

// TestExecuteCreateNotice_Valid tests creating a notice with valid input.
func TestExecuteCreateNotice_Valid(t *testing.T) {
	store := newMockNoticeStore()
	n, err := ExecuteCreateNotice(context.Background(), CreateNoticeInput{
		Type:       notice.TypeSchoolWide,
		Title:      "Open Mat Friday",
		Content:    "**All levels** welcome",
		AuthorName: "Coach Pat",
		ShowAuthor: true,
		Color:      "blue",
		CreatedBy:  "admin-001",
	}, CreateNoticeDeps{
		NoticeStore: store,
		GenerateID:  fixedID,
		Now:         fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.ID != "test-id-001" {
		t.Errorf("expected ID=test-id-001, got %s", n.ID)
	}
	if n.Status != notice.StatusDraft {
		t.Errorf("expected status=draft, got %s", n.Status)
	}
	if n.Color != "blue" {
		t.Errorf("expected color=blue, got %s", n.Color)
	}
	if n.AuthorName != "Coach Pat" {
		t.Errorf("expected AuthorName=Coach Pat, got %s", n.AuthorName)
	}
	if _, ok := store.notices["test-id-001"]; !ok {
		t.Error("expected notice to be persisted in store")
	}
}

// TestExecuteCreateNotice_DefaultColor tests that empty color defaults to orange.
func TestExecuteCreateNotice_DefaultColor(t *testing.T) {
	store := newMockNoticeStore()
	n, err := ExecuteCreateNotice(context.Background(), CreateNoticeInput{
		Type:      notice.TypeSchoolWide,
		Title:     "Test",
		Content:   "content",
		CreatedBy: "admin-001",
	}, CreateNoticeDeps{
		NoticeStore: store,
		GenerateID:  fixedID,
		Now:         fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Color != notice.ColorOrange {
		t.Errorf("expected color=orange, got %s", n.Color)
	}
}

// TestExecuteCreateNotice_InvalidColor tests that invalid color is rejected.
func TestExecuteCreateNotice_InvalidColor(t *testing.T) {
	store := newMockNoticeStore()
	_, err := ExecuteCreateNotice(context.Background(), CreateNoticeInput{
		Type:      notice.TypeSchoolWide,
		Title:     "Test",
		Content:   "content",
		Color:     "neon_pink",
		CreatedBy: "admin-001",
	}, CreateNoticeDeps{
		NoticeStore: store,
		GenerateID:  fixedID,
		Now:         fixedNow,
	})
	if err == nil {
		t.Error("expected error for invalid color")
	}
}

// TestExecuteCreateNotice_MissingCreatedBy tests that empty CreatedBy is rejected.
func TestExecuteCreateNotice_MissingCreatedBy(t *testing.T) {
	store := newMockNoticeStore()
	_, err := ExecuteCreateNotice(context.Background(), CreateNoticeInput{
		Type:    notice.TypeSchoolWide,
		Title:   "Test",
		Content: "content",
	}, CreateNoticeDeps{
		NoticeStore: store,
		GenerateID:  fixedID,
		Now:         fixedNow,
	})
	if err == nil {
		t.Error("expected error for missing CreatedBy")
	}
}

// --- ExecuteEditNotice tests ---

// TestExecuteEditNotice_Valid tests editing a notice with valid input.
func TestExecuteEditNotice_Valid(t *testing.T) {
	store := newMockNoticeStore()
	store.notices["n1"] = notice.Notice{
		ID: "n1", Type: notice.TypeSchoolWide, Status: notice.StatusDraft,
		Title: "Original", Content: "Original content", Color: notice.ColorOrange,
	}

	n, err := ExecuteEditNotice(context.Background(), EditNoticeInput{
		NoticeID:   "n1",
		Title:      "Updated Title",
		Content:    "Updated content",
		Color:      "red",
		AuthorName: "Coach Marcus",
		ShowAuthor: true,
	}, EditNoticeDeps{
		NoticeStore: store,
		Now:         fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Title != "Updated Title" {
		t.Errorf("expected title=Updated Title, got %s", n.Title)
	}
	if n.Color != "red" {
		t.Errorf("expected color=red, got %s", n.Color)
	}
	if n.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

// TestExecuteEditNotice_MissingID tests that empty NoticeID is rejected.
func TestExecuteEditNotice_MissingID(t *testing.T) {
	store := newMockNoticeStore()
	_, err := ExecuteEditNotice(context.Background(), EditNoticeInput{
		Title: "Updated",
	}, EditNoticeDeps{
		NoticeStore: store,
		Now:         fixedNow,
	})
	if err == nil {
		t.Error("expected error for missing NoticeID")
	}
}

// TestExecuteEditNotice_NotFound tests editing a non-existent notice.
func TestExecuteEditNotice_NotFound(t *testing.T) {
	store := newMockNoticeStore()
	_, err := ExecuteEditNotice(context.Background(), EditNoticeInput{
		NoticeID: "nonexistent",
		Title:    "Updated",
	}, EditNoticeDeps{
		NoticeStore: store,
		Now:         fixedNow,
	})
	if err == nil {
		t.Error("expected error for not found notice")
	}
}

// --- ExecutePublishNotice tests ---

// TestExecutePublishNotice_Valid tests publishing a draft notice.
func TestExecutePublishNotice_Valid(t *testing.T) {
	store := newMockNoticeStore()
	store.notices["n1"] = notice.Notice{
		ID: "n1", Type: notice.TypeSchoolWide, Status: notice.StatusDraft,
		Title: "Test", Content: "content", Color: notice.ColorOrange,
	}

	n, err := ExecutePublishNotice(context.Background(), PublishNoticeInput{
		NoticeID:    "n1",
		PublisherID: "admin-001",
	}, PublishNoticeDeps{
		NoticeStore: store,
		Now:         fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Status != notice.StatusPublished {
		t.Errorf("expected status=published, got %s", n.Status)
	}
	if n.PublishedBy != "admin-001" {
		t.Errorf("expected PublishedBy=admin-001, got %s", n.PublishedBy)
	}
	if n.PublishedAt.IsZero() {
		t.Error("expected PublishedAt to be set")
	}
}

// TestExecutePublishNotice_AlreadyPublished tests publishing an already-published notice.
func TestExecutePublishNotice_AlreadyPublished(t *testing.T) {
	store := newMockNoticeStore()
	store.notices["n1"] = notice.Notice{
		ID: "n1", Type: notice.TypeSchoolWide, Status: notice.StatusPublished,
		Title: "Test", Content: "content", Color: notice.ColorOrange,
	}

	_, err := ExecutePublishNotice(context.Background(), PublishNoticeInput{
		NoticeID:    "n1",
		PublisherID: "admin-001",
	}, PublishNoticeDeps{
		NoticeStore: store,
		Now:         fixedNow,
	})
	if err == nil {
		t.Error("expected error for already published notice")
	}
}

// TestExecutePublishNotice_MissingPublisherID tests that empty publisher ID is rejected.
func TestExecutePublishNotice_MissingPublisherID(t *testing.T) {
	store := newMockNoticeStore()
	_, err := ExecutePublishNotice(context.Background(), PublishNoticeInput{
		NoticeID: "n1",
	}, PublishNoticeDeps{
		NoticeStore: store,
		Now:         fixedNow,
	})
	if err == nil {
		t.Error("expected error for missing PublisherID")
	}
}

// --- ExecutePinNotice tests ---

// TestExecutePinNotice_Pin tests pinning a notice.
func TestExecutePinNotice_Pin(t *testing.T) {
	store := newMockNoticeStore()
	store.notices["n1"] = notice.Notice{
		ID: "n1", Type: notice.TypeSchoolWide, Status: notice.StatusPublished,
		Title: "Test", Content: "content", Color: notice.ColorOrange,
	}

	n, err := ExecutePinNotice(context.Background(), PinNoticeInput{
		NoticeID: "n1",
		Pinned:   true,
	}, PinNoticeDeps{
		NoticeStore: store,
		Now:         fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !n.Pinned {
		t.Error("expected Pinned=true")
	}
	if n.PinnedAt.IsZero() {
		t.Error("expected PinnedAt to be set")
	}
}

// TestExecutePinNotice_Unpin tests unpinning a notice.
func TestExecutePinNotice_Unpin(t *testing.T) {
	store := newMockNoticeStore()
	store.notices["n1"] = notice.Notice{
		ID: "n1", Type: notice.TypeSchoolWide, Status: notice.StatusPublished,
		Title: "Test", Content: "content", Color: notice.ColorOrange,
		Pinned: true, PinnedAt: fixedTime,
	}

	n, err := ExecutePinNotice(context.Background(), PinNoticeInput{
		NoticeID: "n1",
		Pinned:   false,
	}, PinNoticeDeps{
		NoticeStore: store,
		Now:         fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Pinned {
		t.Error("expected Pinned=false")
	}
}

// TestExecutePinNotice_MissingID tests that empty NoticeID is rejected.
func TestExecutePinNotice_MissingID(t *testing.T) {
	store := newMockNoticeStore()
	_, err := ExecutePinNotice(context.Background(), PinNoticeInput{
		Pinned: true,
	}, PinNoticeDeps{
		NoticeStore: store,
		Now:         fixedNow,
	})
	if err == nil {
		t.Error("expected error for missing NoticeID")
	}
}

// TestExecutePinNotice_AlreadyPinned tests pinning an already-pinned notice.
func TestExecutePinNotice_AlreadyPinned(t *testing.T) {
	store := newMockNoticeStore()
	store.notices["n1"] = notice.Notice{
		ID: "n1", Type: notice.TypeSchoolWide, Status: notice.StatusPublished,
		Title: "Test", Content: "content", Color: notice.ColorOrange,
		Pinned: true, PinnedAt: fixedTime,
	}

	_, err := ExecutePinNotice(context.Background(), PinNoticeInput{
		NoticeID: "n1",
		Pinned:   true,
	}, PinNoticeDeps{
		NoticeStore: store,
		Now:         fixedNow,
	})
	if err == nil {
		t.Error("expected error for already pinned notice")
	}
}
