package orchestrators

import (
	"context"
	"errors"
	"testing"

	"workshop/internal/domain/observation"
)

// mockObservationStore implements ObservationStoreForOrchestrator for testing.
type mockObservationStore struct {
	observations map[string]observation.Observation
}

// GetByID implements ObservationStoreForOrchestrator.
// PRE: id is non-empty
// POST: returns observation or error
func (m *mockObservationStore) GetByID(_ context.Context, id string) (observation.Observation, error) {
	o, ok := m.observations[id]
	if !ok {
		return observation.Observation{}, errors.New("not found")
	}
	return o, nil
}

// Save implements ObservationStoreForOrchestrator.
// PRE: observation is valid
// POST: observation is persisted
func (m *mockObservationStore) Save(_ context.Context, o observation.Observation) error {
	m.observations[o.ID] = o
	return nil
}

func newMockObservationStore() *mockObservationStore {
	return &mockObservationStore{observations: make(map[string]observation.Observation)}
}

// TestExecuteCreateObservation_Valid tests creating an observation with valid input.
func TestExecuteCreateObservation_Valid(t *testing.T) {
	store := newMockObservationStore()
	obs, err := ExecuteCreateObservation(context.Background(), CreateObservationInput{
		MemberID: "member-001",
		Content:  "Tends to muscle techniques - focus on flow drilling",
		AuthorID: "coach-001",
	}, CreateObservationDeps{
		ObservationStore: store,
		GenerateID:       fixedID,
		Now:              fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ID != "test-id-001" {
		t.Errorf("expected ID=test-id-001, got %s", obs.ID)
	}
	if obs.MemberID != "member-001" {
		t.Errorf("expected MemberID=member-001, got %s", obs.MemberID)
	}
	if obs.AuthorID != "coach-001" {
		t.Errorf("expected AuthorID=coach-001, got %s", obs.AuthorID)
	}
	if _, ok := store.observations["test-id-001"]; !ok {
		t.Error("expected observation to be persisted in store")
	}
}

// TestExecuteCreateObservation_MissingAuthor tests that empty AuthorID is rejected.
func TestExecuteCreateObservation_MissingAuthor(t *testing.T) {
	store := newMockObservationStore()
	_, err := ExecuteCreateObservation(context.Background(), CreateObservationInput{
		MemberID: "member-001",
		Content:  "Some observation",
	}, CreateObservationDeps{
		ObservationStore: store,
		GenerateID:       fixedID,
		Now:              fixedNow,
	})
	if err == nil {
		t.Error("expected error for missing AuthorID")
	}
}

// TestExecuteCreateObservation_MissingContent tests that empty content is rejected.
func TestExecuteCreateObservation_MissingContent(t *testing.T) {
	store := newMockObservationStore()
	_, err := ExecuteCreateObservation(context.Background(), CreateObservationInput{
		MemberID: "member-001",
		AuthorID: "coach-001",
	}, CreateObservationDeps{
		ObservationStore: store,
		GenerateID:       fixedID,
		Now:              fixedNow,
	})
	if err == nil {
		t.Error("expected error for missing content")
	}
}

// TestExecuteCreateObservation_MissingMember tests that empty MemberID is rejected.
func TestExecuteCreateObservation_MissingMember(t *testing.T) {
	store := newMockObservationStore()
	_, err := ExecuteCreateObservation(context.Background(), CreateObservationInput{
		Content:  "Some observation",
		AuthorID: "coach-001",
	}, CreateObservationDeps{
		ObservationStore: store,
		GenerateID:       fixedID,
		Now:              fixedNow,
	})
	if err == nil {
		t.Error("expected error for missing MemberID")
	}
}

// TestExecuteEditObservation_Valid tests editing an observation with valid input.
func TestExecuteEditObservation_Valid(t *testing.T) {
	store := newMockObservationStore()
	store.observations["obs-1"] = observation.Observation{
		ID: "obs-1", MemberID: "member-001", AuthorID: "coach-001",
		Content: "Original note", CreatedAt: fixedTime,
	}

	obs, err := ExecuteEditObservation(context.Background(), EditObservationInput{
		ObservationID: "obs-1",
		Content:       "Updated: focus on hip escapes",
	}, EditObservationDeps{
		ObservationStore: store,
		Now:              fixedNow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.Content != "Updated: focus on hip escapes" {
		t.Errorf("expected updated content, got %s", obs.Content)
	}
	if obs.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

// TestExecuteEditObservation_MissingID tests that empty ObservationID is rejected.
func TestExecuteEditObservation_MissingID(t *testing.T) {
	store := newMockObservationStore()
	_, err := ExecuteEditObservation(context.Background(), EditObservationInput{
		Content: "Updated",
	}, EditObservationDeps{
		ObservationStore: store,
		Now:              fixedNow,
	})
	if err == nil {
		t.Error("expected error for missing ObservationID")
	}
}

// TestExecuteEditObservation_EmptyContent tests that empty content is rejected.
func TestExecuteEditObservation_EmptyContent(t *testing.T) {
	store := newMockObservationStore()
	_, err := ExecuteEditObservation(context.Background(), EditObservationInput{
		ObservationID: "obs-1",
		Content:       "",
	}, EditObservationDeps{
		ObservationStore: store,
		Now:              fixedNow,
	})
	if err == nil {
		t.Error("expected error for empty content")
	}
}

// TestExecuteEditObservation_NotFound tests editing a non-existent observation.
func TestExecuteEditObservation_NotFound(t *testing.T) {
	store := newMockObservationStore()
	_, err := ExecuteEditObservation(context.Background(), EditObservationInput{
		ObservationID: "nonexistent",
		Content:       "Updated",
	}, EditObservationDeps{
		ObservationStore: store,
		Now:              fixedNow,
	})
	if err == nil {
		t.Error("expected error for not found observation")
	}
}
