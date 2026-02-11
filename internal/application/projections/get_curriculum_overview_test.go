package projections

import (
	"context"
	"errors"
	"testing"

	"workshop/internal/domain/classtype"
	"workshop/internal/domain/rotor"
)

// mockCurriculumClassTypeStore implements CurriculumOverviewClassTypeStore for testing.
type mockCurriculumClassTypeStore struct {
	classTypes []classtype.ClassType
}

// List implements CurriculumOverviewClassTypeStore.
// PRE: none
// POST: returns stored class types
func (m *mockCurriculumClassTypeStore) List(_ context.Context) ([]classtype.ClassType, error) {
	return m.classTypes, nil
}

// mockCurriculumRotorStore implements CurriculumOverviewRotorStore for testing.
type mockCurriculumRotorStore struct {
	activeRotors map[string]rotor.Rotor         // key: classTypeID
	themes       map[string][]rotor.RotorTheme  // key: rotorID
	topics       map[string][]rotor.Topic       // key: rotorThemeID
	schedules    map[string]rotor.TopicSchedule // key: rotorThemeID
	votes        map[string]int                 // key: topicID
}

// GetActiveRotor implements CurriculumOverviewRotorStore.
// PRE: classTypeID is non-empty
// POST: returns the active rotor or error if none
func (m *mockCurriculumRotorStore) GetActiveRotor(_ context.Context, classTypeID string) (rotor.Rotor, error) {
	r, ok := m.activeRotors[classTypeID]
	if !ok {
		return rotor.Rotor{}, errors.New("no active rotor")
	}
	return r, nil
}

// ListThemesByRotor implements CurriculumOverviewRotorStore.
// PRE: rotorID is non-empty
// POST: returns themes for the rotor
func (m *mockCurriculumRotorStore) ListThemesByRotor(_ context.Context, rotorID string) ([]rotor.RotorTheme, error) {
	return m.themes[rotorID], nil
}

// ListTopicsByTheme implements CurriculumOverviewRotorStore.
// PRE: rotorThemeID is non-empty
// POST: returns topics for the theme
func (m *mockCurriculumRotorStore) ListTopicsByTheme(_ context.Context, rotorThemeID string) ([]rotor.Topic, error) {
	return m.topics[rotorThemeID], nil
}

// GetActiveScheduleForTheme implements CurriculumOverviewRotorStore.
// PRE: rotorThemeID is non-empty
// POST: returns the active schedule or error if none
func (m *mockCurriculumRotorStore) GetActiveScheduleForTheme(_ context.Context, rotorThemeID string) (rotor.TopicSchedule, error) {
	s, ok := m.schedules[rotorThemeID]
	if !ok {
		return rotor.TopicSchedule{}, errors.New("no active schedule")
	}
	return s, nil
}

// CountVotesForTopic implements CurriculumOverviewRotorStore.
// PRE: topicID is non-empty
// POST: returns vote count for the topic
func (m *mockCurriculumRotorStore) CountVotesForTopic(_ context.Context, topicID string) (int, error) {
	return m.votes[topicID], nil
}

// TestQueryCurriculumOverview_ActiveTopics verifies that active topics are returned per theme.
func TestQueryCurriculumOverview_ActiveTopics(t *testing.T) {
	classTypeStore := &mockCurriculumClassTypeStore{
		classTypes: []classtype.ClassType{
			{ID: "ct1", ProgramID: "p1", Name: "Gi Express"},
		},
	}
	rotorStore := &mockCurriculumRotorStore{
		activeRotors: map[string]rotor.Rotor{
			"ct1": {ID: "r1", ClassTypeID: "ct1", Name: "v1", Status: "active", PreviewOn: false},
		},
		themes: map[string][]rotor.RotorTheme{
			"r1": {
				{ID: "th1", RotorID: "r1", Name: "Standing", Position: 0},
				{ID: "th2", RotorID: "r1", Name: "Guard", Position: 1},
			},
		},
		topics: map[string][]rotor.Topic{
			"th1": {
				{ID: "tp1", RotorThemeID: "th1", Name: "Single Leg", DurationWeeks: 1, Position: 0},
				{ID: "tp2", RotorThemeID: "th1", Name: "Double Leg", DurationWeeks: 1, Position: 1},
			},
			"th2": {
				{ID: "tp3", RotorThemeID: "th2", Name: "Closed Guard", DurationWeeks: 1, Position: 0},
			},
		},
		schedules: map[string]rotor.TopicSchedule{
			"th1": {ID: "s1", TopicID: "tp1", RotorThemeID: "th1", Status: "active"},
			"th2": {ID: "s2", TopicID: "tp3", RotorThemeID: "th2", Status: "active"},
		},
		votes: map[string]int{"tp1": 3},
	}

	query := GetCurriculumOverviewQuery{Role: "member"}
	deps := GetCurriculumOverviewDeps{ClassTypeStore: classTypeStore, RotorStore: rotorStore}
	result, err := QueryGetCurriculumOverview(context.Background(), query, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ClassCurriculums) != 1 {
		t.Fatalf("expected 1 class curriculum, got %d", len(result.ClassCurriculums))
	}
	cc := result.ClassCurriculums[0]
	if cc.ClassTypeName != "Gi Express" {
		t.Errorf("class type name = %q, want %q", cc.ClassTypeName, "Gi Express")
	}
	if len(cc.Themes) != 2 {
		t.Fatalf("expected 2 themes, got %d", len(cc.Themes))
	}

	// Theme 1: Standing — active topic is Single Leg with 3 votes
	th1 := cc.Themes[0]
	if th1.ActiveTopic == nil {
		t.Fatal("expected active topic for Standing theme")
	}
	if th1.ActiveTopic.TopicName != "Single Leg" {
		t.Errorf("active topic = %q, want %q", th1.ActiveTopic.TopicName, "Single Leg")
	}
	if th1.ActiveTopic.Votes != 3 {
		t.Errorf("votes = %d, want 3", th1.ActiveTopic.Votes)
	}

	// Theme 2: Guard — active topic is Closed Guard
	th2 := cc.Themes[1]
	if th2.ActiveTopic == nil {
		t.Fatal("expected active topic for Guard theme")
	}
	if th2.ActiveTopic.TopicName != "Closed Guard" {
		t.Errorf("active topic = %q, want %q", th2.ActiveTopic.TopicName, "Closed Guard")
	}
}

// TestQueryCurriculumOverview_PreviewOff verifies that upcoming topics are hidden for members when preview is off.
func TestQueryCurriculumOverview_PreviewOff(t *testing.T) {
	classTypeStore := &mockCurriculumClassTypeStore{
		classTypes: []classtype.ClassType{
			{ID: "ct1", ProgramID: "p1", Name: "Gi Express"},
		},
	}
	rotorStore := &mockCurriculumRotorStore{
		activeRotors: map[string]rotor.Rotor{
			"ct1": {ID: "r1", ClassTypeID: "ct1", Name: "v1", Status: "active", PreviewOn: false},
		},
		themes: map[string][]rotor.RotorTheme{
			"r1": {{ID: "th1", RotorID: "r1", Name: "Standing", Position: 0}},
		},
		topics: map[string][]rotor.Topic{
			"th1": {
				{ID: "tp1", RotorThemeID: "th1", Name: "Single Leg", DurationWeeks: 1, Position: 0},
				{ID: "tp2", RotorThemeID: "th1", Name: "Double Leg", DurationWeeks: 1, Position: 1},
				{ID: "tp3", RotorThemeID: "th1", Name: "Arm Drag", DurationWeeks: 1, Position: 2},
			},
		},
		schedules: map[string]rotor.TopicSchedule{
			"th1": {ID: "s1", TopicID: "tp1", RotorThemeID: "th1", Status: "active"},
		},
		votes: map[string]int{},
	}

	query := GetCurriculumOverviewQuery{Role: "member"}
	deps := GetCurriculumOverviewDeps{ClassTypeStore: classTypeStore, RotorStore: rotorStore}
	result, err := QueryGetCurriculumOverview(context.Background(), query, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	th := result.ClassCurriculums[0].Themes[0]
	if th.ActiveTopic == nil {
		t.Fatal("expected active topic")
	}
	if len(th.Upcoming) != 0 {
		t.Errorf("expected 0 upcoming topics for member with preview off, got %d", len(th.Upcoming))
	}
}

// TestQueryCurriculumOverview_PreviewOn verifies that upcoming topics are shown when preview is on.
func TestQueryCurriculumOverview_PreviewOn(t *testing.T) {
	classTypeStore := &mockCurriculumClassTypeStore{
		classTypes: []classtype.ClassType{
			{ID: "ct1", ProgramID: "p1", Name: "Gi Express"},
		},
	}
	rotorStore := &mockCurriculumRotorStore{
		activeRotors: map[string]rotor.Rotor{
			"ct1": {ID: "r1", ClassTypeID: "ct1", Name: "v1", Status: "active", PreviewOn: true},
		},
		themes: map[string][]rotor.RotorTheme{
			"r1": {{ID: "th1", RotorID: "r1", Name: "Standing", Position: 0}},
		},
		topics: map[string][]rotor.Topic{
			"th1": {
				{ID: "tp1", RotorThemeID: "th1", Name: "Single Leg", DurationWeeks: 1, Position: 0},
				{ID: "tp2", RotorThemeID: "th1", Name: "Double Leg", DurationWeeks: 1, Position: 1},
				{ID: "tp3", RotorThemeID: "th1", Name: "Arm Drag", DurationWeeks: 1, Position: 2},
			},
		},
		schedules: map[string]rotor.TopicSchedule{
			"th1": {ID: "s1", TopicID: "tp1", RotorThemeID: "th1", Status: "active"},
		},
		votes: map[string]int{},
	}

	query := GetCurriculumOverviewQuery{Role: "member"}
	deps := GetCurriculumOverviewDeps{ClassTypeStore: classTypeStore, RotorStore: rotorStore}
	result, err := QueryGetCurriculumOverview(context.Background(), query, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	th := result.ClassCurriculums[0].Themes[0]
	if th.ActiveTopic == nil {
		t.Fatal("expected active topic")
	}
	if len(th.Upcoming) != 2 {
		t.Errorf("expected 2 upcoming topics with preview on, got %d", len(th.Upcoming))
	}
	if th.Upcoming[0].TopicName != "Double Leg" {
		t.Errorf("first upcoming = %q, want %q", th.Upcoming[0].TopicName, "Double Leg")
	}
}

// TestQueryCurriculumOverview_AdminSeesUpcomingEvenWithPreviewOff verifies admin override.
func TestQueryCurriculumOverview_AdminSeesUpcomingEvenWithPreviewOff(t *testing.T) {
	classTypeStore := &mockCurriculumClassTypeStore{
		classTypes: []classtype.ClassType{
			{ID: "ct1", ProgramID: "p1", Name: "Gi Express"},
		},
	}
	rotorStore := &mockCurriculumRotorStore{
		activeRotors: map[string]rotor.Rotor{
			"ct1": {ID: "r1", ClassTypeID: "ct1", Name: "v1", Status: "active", PreviewOn: false},
		},
		themes: map[string][]rotor.RotorTheme{
			"r1": {{ID: "th1", RotorID: "r1", Name: "Standing", Position: 0}},
		},
		topics: map[string][]rotor.Topic{
			"th1": {
				{ID: "tp1", RotorThemeID: "th1", Name: "Single Leg", DurationWeeks: 1, Position: 0},
				{ID: "tp2", RotorThemeID: "th1", Name: "Double Leg", DurationWeeks: 1, Position: 1},
			},
		},
		schedules: map[string]rotor.TopicSchedule{
			"th1": {ID: "s1", TopicID: "tp1", RotorThemeID: "th1", Status: "active"},
		},
		votes: map[string]int{},
	}

	query := GetCurriculumOverviewQuery{Role: "admin"}
	deps := GetCurriculumOverviewDeps{ClassTypeStore: classTypeStore, RotorStore: rotorStore}
	result, err := QueryGetCurriculumOverview(context.Background(), query, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	th := result.ClassCurriculums[0].Themes[0]
	if len(th.Upcoming) != 1 {
		t.Errorf("expected 1 upcoming topic for admin even with preview off, got %d", len(th.Upcoming))
	}
}

// TestQueryCurriculumOverview_NoActiveRotor verifies class types without active rotors are skipped.
func TestQueryCurriculumOverview_NoActiveRotor(t *testing.T) {
	classTypeStore := &mockCurriculumClassTypeStore{
		classTypes: []classtype.ClassType{
			{ID: "ct1", ProgramID: "p1", Name: "Gi Express"},
			{ID: "ct2", ProgramID: "p1", Name: "No-Gi Advanced"},
		},
	}
	rotorStore := &mockCurriculumRotorStore{
		activeRotors: map[string]rotor.Rotor{
			// Only ct1 has an active rotor; ct2 does not
			"ct1": {ID: "r1", ClassTypeID: "ct1", Name: "v1", Status: "active"},
		},
		themes:    map[string][]rotor.RotorTheme{"r1": {}},
		topics:    map[string][]rotor.Topic{},
		schedules: map[string]rotor.TopicSchedule{},
		votes:     map[string]int{},
	}

	query := GetCurriculumOverviewQuery{Role: "member"}
	deps := GetCurriculumOverviewDeps{ClassTypeStore: classTypeStore, RotorStore: rotorStore}
	result, err := QueryGetCurriculumOverview(context.Background(), query, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ClassCurriculums) != 1 {
		t.Errorf("expected 1 class curriculum (only ct1 has active rotor), got %d", len(result.ClassCurriculums))
	}
	if result.ClassCurriculums[0].ClassTypeName != "Gi Express" {
		t.Errorf("class type = %q, want %q", result.ClassCurriculums[0].ClassTypeName, "Gi Express")
	}
}

// TestQueryCurriculumOverview_EmptyClassTypes verifies zero state.
func TestQueryCurriculumOverview_EmptyClassTypes(t *testing.T) {
	classTypeStore := &mockCurriculumClassTypeStore{
		classTypes: []classtype.ClassType{},
	}
	rotorStore := &mockCurriculumRotorStore{
		activeRotors: map[string]rotor.Rotor{},
		themes:       map[string][]rotor.RotorTheme{},
		topics:       map[string][]rotor.Topic{},
		schedules:    map[string]rotor.TopicSchedule{},
		votes:        map[string]int{},
	}

	query := GetCurriculumOverviewQuery{Role: "member"}
	deps := GetCurriculumOverviewDeps{ClassTypeStore: classTypeStore, RotorStore: rotorStore}
	result, err := QueryGetCurriculumOverview(context.Background(), query, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ClassCurriculums) != 0 {
		t.Errorf("expected 0 class curriculums, got %d", len(result.ClassCurriculums))
	}
}
