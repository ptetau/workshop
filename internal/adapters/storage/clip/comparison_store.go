package clip

import (
	"context"

	domain "workshop/internal/domain/clip"
)

// ComparisonStore persists ComparisonSession and ResearchNote state.
type ComparisonStore interface {
	// Comparison session management
	SaveSession(ctx context.Context, session domain.ComparisonSession) error
	GetSessionByID(ctx context.Context, id string) (domain.ComparisonSession, error)
	ListSessionsByUser(ctx context.Context, userID string) ([]domain.ComparisonSession, error)
	DeleteSession(ctx context.Context, id string) error

	// Research notes
	SaveResearchNote(ctx context.Context, note domain.ResearchNote) error
	GetResearchNoteBySession(ctx context.Context, sessionID string) (domain.ResearchNote, error)
	DeleteResearchNote(ctx context.Context, id string) error
}
