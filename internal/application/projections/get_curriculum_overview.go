package projections

import (
	"context"

	"workshop/internal/domain/classtype"
	"workshop/internal/domain/rotor"
)

// CurriculumOverviewClassTypeStore defines the class type store interface needed by the curriculum overview projection.
type CurriculumOverviewClassTypeStore interface {
	List(ctx context.Context) ([]classtype.ClassType, error)
}

// CurriculumOverviewRotorStore defines the rotor store interface needed by the curriculum overview projection.
type CurriculumOverviewRotorStore interface {
	GetActiveRotor(ctx context.Context, classTypeID string) (rotor.Rotor, error)
	ListThemesByRotor(ctx context.Context, rotorID string) ([]rotor.RotorTheme, error)
	ListTopicsByTheme(ctx context.Context, rotorThemeID string) ([]rotor.Topic, error)
	GetActiveScheduleForTheme(ctx context.Context, rotorThemeID string) (rotor.TopicSchedule, error)
	CountVotesForTopic(ctx context.Context, topicID string) (int, error)
}

// GetCurriculumOverviewQuery carries input for the curriculum overview projection.
type GetCurriculumOverviewQuery struct {
	Role string // viewer's role: admin, coach, member, trial
}

// GetCurriculumOverviewDeps holds dependencies for the curriculum overview projection.
type GetCurriculumOverviewDeps struct {
	ClassTypeStore CurriculumOverviewClassTypeStore
	RotorStore     CurriculumOverviewRotorStore
}

// CurriculumOverviewResult carries the output of the curriculum overview projection.
type CurriculumOverviewResult struct {
	ClassCurriculums []ClassCurriculum `json:"class_curriculums"`
}

// ClassCurriculum represents the active curriculum for a single class type.
type ClassCurriculum struct {
	ClassTypeID   string                `json:"class_type_id"`
	ClassTypeName string                `json:"class_type_name"`
	RotorID       string                `json:"rotor_id"`
	RotorName     string                `json:"rotor_name"`
	PreviewOn     bool                  `json:"preview_on"`
	Themes        []CurriculumThemeView `json:"themes"`
}

// CurriculumThemeView represents a theme with its active and upcoming topics.
type CurriculumThemeView struct {
	ThemeID     string                `json:"theme_id"`
	ThemeName   string                `json:"theme_name"`
	Position    int                   `json:"position"`
	ActiveTopic *CurriculumTopicView  `json:"active_topic"`
	Upcoming    []CurriculumTopicView `json:"upcoming"`
}

// CurriculumTopicView represents a topic in the curriculum overview.
type CurriculumTopicView struct {
	TopicID       string `json:"topic_id"`
	TopicName     string `json:"topic_name"`
	Description   string `json:"description"`
	DurationWeeks int    `json:"duration_weeks"`
	Votes         int    `json:"votes"`
	Position      int    `json:"position"`
}

// QueryGetCurriculumOverview aggregates the active curriculum across all class types.
// PRE: deps are valid and non-nil
// POST: returns the curriculum overview or error
func QueryGetCurriculumOverview(ctx context.Context, query GetCurriculumOverviewQuery, deps GetCurriculumOverviewDeps) (CurriculumOverviewResult, error) {
	result := CurriculumOverviewResult{}

	classTypes, err := deps.ClassTypeStore.List(ctx)
	if err != nil {
		return result, err
	}

	for _, ct := range classTypes {
		activeRotor, err := deps.RotorStore.GetActiveRotor(ctx, ct.ID)
		if err != nil {
			// No active rotor for this class type — skip
			continue
		}

		cc := ClassCurriculum{
			ClassTypeID:   ct.ID,
			ClassTypeName: ct.Name,
			RotorID:       activeRotor.ID,
			RotorName:     activeRotor.Name,
			PreviewOn:     activeRotor.PreviewOn,
		}

		themes, _ := deps.RotorStore.ListThemesByRotor(ctx, activeRotor.ID)
		for _, th := range themes {
			tv := CurriculumThemeView{
				ThemeID:   th.ID,
				ThemeName: th.Name,
				Position:  th.Position,
			}

			topics, _ := deps.RotorStore.ListTopicsByTheme(ctx, th.ID)
			activeSched, schedErr := deps.RotorStore.GetActiveScheduleForTheme(ctx, th.ID)

			for _, tp := range topics {
				votes, _ := deps.RotorStore.CountVotesForTopic(ctx, tp.ID)
				topicView := CurriculumTopicView{
					TopicID:       tp.ID,
					TopicName:     tp.Name,
					Description:   tp.Description,
					DurationWeeks: tp.DurationWeeks,
					Votes:         votes,
					Position:      tp.Position,
				}

				if schedErr == nil && activeSched.TopicID == tp.ID {
					tv.ActiveTopic = &topicView
				} else if activeRotor.PreviewOn || query.Role == "admin" || query.Role == "coach" {
					tv.Upcoming = append(tv.Upcoming, topicView)
				}
			}

			if tv.Upcoming == nil {
				tv.Upcoming = []CurriculumTopicView{}
			}
			cc.Themes = append(cc.Themes, tv)
		}
		if cc.Themes == nil {
			cc.Themes = []CurriculumThemeView{}
		}

		result.ClassCurriculums = append(result.ClassCurriculums, cc)
	}

	if result.ClassCurriculums == nil {
		result.ClassCurriculums = []ClassCurriculum{}
	}

	return result, nil
}
