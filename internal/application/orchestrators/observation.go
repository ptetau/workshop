package orchestrators

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"workshop/internal/domain/observation"
)

// ObservationStoreForOrchestrator defines the store interface needed by observation orchestrators.
type ObservationStoreForOrchestrator interface {
	GetByID(ctx context.Context, id string) (observation.Observation, error)
	Save(ctx context.Context, o observation.Observation) error
}

// --- Create Observation ---

// CreateObservationInput carries input for the create observation orchestrator.
type CreateObservationInput struct {
	MemberID string
	Content  string
	AuthorID string // AccountID of coach/admin creating the observation
}

// CreateObservationDeps holds dependencies for CreateObservation.
type CreateObservationDeps struct {
	ObservationStore ObservationStoreForOrchestrator
	GenerateID       func() string
	Now              func() time.Time
}

// ExecuteCreateObservation creates a new private observation on a member's profile.
// PRE: MemberID, Content, and AuthorID must be non-empty
// POST: Observation created with generated ID and timestamps
func ExecuteCreateObservation(ctx context.Context, input CreateObservationInput, deps CreateObservationDeps) (observation.Observation, error) {
	if input.AuthorID == "" {
		return observation.Observation{}, errors.New("author ID is required")
	}

	obs := observation.Observation{
		ID:        deps.GenerateID(),
		MemberID:  input.MemberID,
		AuthorID:  input.AuthorID,
		Content:   input.Content,
		CreatedAt: deps.Now(),
	}

	if err := obs.Validate(); err != nil {
		return observation.Observation{}, err
	}

	if err := deps.ObservationStore.Save(ctx, obs); err != nil {
		return observation.Observation{}, err
	}

	slog.Info("observation_event", "event", "observation_created", "observation_id", obs.ID, "member_id", obs.MemberID, "author_id", obs.AuthorID)
	return obs, nil
}

// --- Edit Observation ---

// EditObservationInput carries input for the edit observation orchestrator.
type EditObservationInput struct {
	ObservationID string
	Content       string
}

// EditObservationDeps holds dependencies for EditObservation.
type EditObservationDeps struct {
	ObservationStore ObservationStoreForOrchestrator
	Now              func() time.Time
}

// ExecuteEditObservation updates the content of an existing observation.
// PRE: ObservationID must be non-empty; observation must exist; Content must be non-empty
// POST: Observation content and UpdatedAt updated
func ExecuteEditObservation(ctx context.Context, input EditObservationInput, deps EditObservationDeps) (observation.Observation, error) {
	if input.ObservationID == "" {
		return observation.Observation{}, errors.New("observation ID is required")
	}
	if input.Content == "" {
		return observation.Observation{}, observation.ErrEmptyContent
	}

	obs, err := deps.ObservationStore.GetByID(ctx, input.ObservationID)
	if err != nil {
		return observation.Observation{}, err
	}

	obs.Content = input.Content
	obs.UpdatedAt = deps.Now()

	if err := obs.Validate(); err != nil {
		return observation.Observation{}, err
	}

	if err := deps.ObservationStore.Save(ctx, obs); err != nil {
		return observation.Observation{}, err
	}

	slog.Info("observation_event", "event", "observation_edited", "observation_id", obs.ID)
	return obs, nil
}
