package orchestrators

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"workshop/internal/domain/notice"
)

// NoticeStoreForOrchestrator defines the store interface needed by notice orchestrators.
type NoticeStoreForOrchestrator interface {
	GetByID(ctx context.Context, id string) (notice.Notice, error)
	Save(ctx context.Context, n notice.Notice) error
}

// --- Create Notice ---

// CreateNoticeInput carries input for the create notice orchestrator.
type CreateNoticeInput struct {
	Type         string
	Title        string
	Content      string
	TargetID     string
	AuthorName   string
	ShowAuthor   bool
	Color        string
	VisibleFrom  time.Time
	VisibleUntil time.Time
	CreatedBy    string // AccountID of creator
}

// CreateNoticeDeps holds dependencies for CreateNotice.
type CreateNoticeDeps struct {
	NoticeStore NoticeStoreForOrchestrator
	GenerateID  func() string
	Now         func() time.Time
}

// ExecuteCreateNotice creates a new notice in draft status.
// PRE: Title, Content, Type must be non-empty; CreatedBy must be non-empty; Color (if set) must be valid
// POST: Notice created in draft status with generated ID
func ExecuteCreateNotice(ctx context.Context, input CreateNoticeInput, deps CreateNoticeDeps) (notice.Notice, error) {
	if input.CreatedBy == "" {
		return notice.Notice{}, errors.New("creator account ID is required")
	}

	color := input.Color
	if color == "" {
		color = notice.ColorOrange
	}

	n := notice.Notice{
		ID:           deps.GenerateID(),
		Type:         input.Type,
		Status:       notice.StatusDraft,
		Title:        input.Title,
		Content:      input.Content,
		CreatedBy:    input.CreatedBy,
		TargetID:     input.TargetID,
		AuthorName:   input.AuthorName,
		ShowAuthor:   input.ShowAuthor,
		Color:        color,
		VisibleFrom:  input.VisibleFrom,
		VisibleUntil: input.VisibleUntil,
		CreatedAt:    deps.Now(),
	}

	if err := n.Validate(); err != nil {
		return notice.Notice{}, err
	}

	if err := deps.NoticeStore.Save(ctx, n); err != nil {
		return notice.Notice{}, err
	}

	slog.Info("notice_event", "event", "notice_created", "notice_id", n.ID, "type", n.Type, "created_by", input.CreatedBy)
	return n, nil
}

// --- Edit Notice ---

// EditNoticeInput carries input for the edit notice orchestrator.
type EditNoticeInput struct {
	NoticeID     string
	Title        string
	Content      string
	Type         string
	AuthorName   string
	ShowAuthor   bool
	Color        string
	VisibleFrom  time.Time
	VisibleUntil time.Time
	// ClearVisibleFrom and ClearVisibleUntil signal explicit clearing of the window.
	ClearVisibleFrom  bool
	ClearVisibleUntil bool
}

// EditNoticeDeps holds dependencies for EditNotice.
type EditNoticeDeps struct {
	NoticeStore NoticeStoreForOrchestrator
	Now         func() time.Time
}

// ExecuteEditNotice updates fields on an existing notice.
// Partial-update semantics:
//   - Title, Content, Type, Color: only updated when the input value is non-empty (cannot be cleared).
//   - AuthorName, ShowAuthor, VisibleFrom, VisibleUntil: always overwritten (can be cleared by sending zero-values).
//
// PRE: NoticeID must be non-empty; notice must exist
// POST: Notice fields updated, UpdatedAt set
func ExecuteEditNotice(ctx context.Context, input EditNoticeInput, deps EditNoticeDeps) (notice.Notice, error) {
	if input.NoticeID == "" {
		return notice.Notice{}, errors.New("notice ID is required")
	}

	n, err := deps.NoticeStore.GetByID(ctx, input.NoticeID)
	if err != nil {
		return notice.Notice{}, err
	}

	if input.Title != "" {
		n.Title = input.Title
	}
	if input.Content != "" {
		n.Content = input.Content
	}
	if input.Type != "" {
		n.Type = input.Type
	}
	n.AuthorName = input.AuthorName
	n.ShowAuthor = input.ShowAuthor
	if input.Color != "" {
		n.Color = input.Color
	}
	n.VisibleFrom = input.VisibleFrom
	n.VisibleUntil = input.VisibleUntil
	n.UpdatedAt = deps.Now()

	if err := n.Validate(); err != nil {
		return notice.Notice{}, err
	}

	if err := deps.NoticeStore.Save(ctx, n); err != nil {
		return notice.Notice{}, err
	}

	slog.Info("notice_event", "event", "notice_edited", "notice_id", n.ID, "title", n.Title)
	return n, nil
}

// --- Publish Notice ---

// PublishNoticeInput carries input for the publish notice orchestrator.
type PublishNoticeInput struct {
	NoticeID    string
	PublisherID string // AccountID of publisher
}

// PublishNoticeDeps holds dependencies for PublishNotice.
type PublishNoticeDeps struct {
	NoticeStore NoticeStoreForOrchestrator
	Now         func() time.Time
}

// ExecutePublishNotice publishes a draft notice.
// PRE: NoticeID and PublisherID must be non-empty; notice must exist and be in draft status
// POST: Notice status set to published, PublishedBy and PublishedAt set
func ExecutePublishNotice(ctx context.Context, input PublishNoticeInput, deps PublishNoticeDeps) (notice.Notice, error) {
	if input.NoticeID == "" {
		return notice.Notice{}, errors.New("notice ID is required")
	}
	if input.PublisherID == "" {
		return notice.Notice{}, errors.New("publisher ID is required")
	}

	n, err := deps.NoticeStore.GetByID(ctx, input.NoticeID)
	if err != nil {
		return notice.Notice{}, err
	}

	if err := n.Publish(input.PublisherID, deps.Now()); err != nil {
		return notice.Notice{}, err
	}

	if err := deps.NoticeStore.Save(ctx, n); err != nil {
		return notice.Notice{}, err
	}

	slog.Info("notice_event", "event", "notice_published", "notice_id", n.ID, "published_by", input.PublisherID)
	return n, nil
}

// --- Pin/Unpin Notice ---

// PinNoticeInput carries input for the pin/unpin notice orchestrator.
type PinNoticeInput struct {
	NoticeID string
	Pinned   bool // true = pin, false = unpin
}

// PinNoticeDeps holds dependencies for PinNotice.
type PinNoticeDeps struct {
	NoticeStore NoticeStoreForOrchestrator
	Now         func() time.Time
}

// ExecutePinNotice pins or unpins a notice.
// PRE: NoticeID must be non-empty; notice must exist
// POST: Pinned/PinnedAt updated, UpdatedAt set
func ExecutePinNotice(ctx context.Context, input PinNoticeInput, deps PinNoticeDeps) (notice.Notice, error) {
	if input.NoticeID == "" {
		return notice.Notice{}, errors.New("notice ID is required")
	}

	n, err := deps.NoticeStore.GetByID(ctx, input.NoticeID)
	if err != nil {
		return notice.Notice{}, err
	}

	if input.Pinned {
		if err := n.Pin(deps.Now()); err != nil {
			return notice.Notice{}, err
		}
	} else {
		if err := n.Unpin(); err != nil {
			return notice.Notice{}, err
		}
	}
	n.UpdatedAt = deps.Now()

	if err := deps.NoticeStore.Save(ctx, n); err != nil {
		return notice.Notice{}, err
	}

	action := "notice_pinned"
	if !input.Pinned {
		action = "notice_unpinned"
	}
	slog.Info("notice_event", "event", action, "notice_id", n.ID)
	return n, nil
}
