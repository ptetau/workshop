package web

import (
	"encoding/json"
	"net/http"

	"workshop/internal/adapters/http/middleware"
	rotorDomain "workshop/internal/domain/rotor"
)

// handleCalendarRotors handles GET for /api/calendar/rotors
// Returns active rotors with their current and upcoming topics for calendar display.
func handleCalendarRotors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "calendar") {
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get all class types to find active rotors
	classTypes, err := stores.ClassTypeStore.List(ctx)
	if err != nil {
		internalError(w, err)
		return
	}

	var result []rotorCalendarDTO

	for _, ct := range classTypes {
		// Try to get active rotor for this class type
		rotor, err := stores.RotorStore.GetActiveRotor(ctx, ct.ID)
		if err != nil {
			// No active rotor for this class type - skip
			continue
		}

		// Get themes for this rotor
		themes, err := stores.RotorStore.ListThemesByRotor(ctx, rotor.ID)
		if err != nil {
			continue
		}

		var themeDTOs []rotorThemeCalendarDTO
		for _, theme := range themes {
			// Get current active schedule
			schedule, err := stores.RotorStore.GetActiveScheduleForTheme(ctx, theme.ID)
			if err != nil {
				continue
			}

			// Get the current topic
			topic, err := stores.RotorStore.GetTopic(ctx, schedule.TopicID)
			if err != nil {
				continue
			}

			dto := rotorThemeCalendarDTO{
				ThemeID:   theme.ID,
				ThemeName: theme.Name,
				Current: topicCalendarDTO{
					TopicID:     topic.ID,
					TopicName:   topic.Name,
					StartDate:   schedule.StartDate.Format("2006-01-02"),
					EndDate:     schedule.EndDate.Format("2006-01-02"),
					Description: topic.Description,
				},
			}

			// Get next topic if available
			topics, err := stores.RotorStore.ListTopicsByTheme(ctx, theme.ID)
			if err == nil && len(topics) > 1 {
				nextTopic := rotorDomain.NextTopicInQueue(topics, topic.ID)
				if nextTopic != nil {
					// Calculate next start date (after current ends)
					nextStart := schedule.EndDate.AddDate(0, 0, 1)
					nextEnd := nextStart.AddDate(0, 0, nextTopic.DurationWeeks*7-1)

					dto.Next = &topicCalendarDTO{
						TopicID:     nextTopic.ID,
						TopicName:   nextTopic.Name,
						StartDate:   nextStart.Format("2006-01-02"),
						EndDate:     nextEnd.Format("2006-01-02"),
						Description: nextTopic.Description,
					}
				}
			}

			themeDTOs = append(themeDTOs, dto)
		}

		if len(themeDTOs) > 0 {
			result = append(result, rotorCalendarDTO{
				RotorID:       rotor.ID,
				RotorName:     rotor.Name,
				ClassTypeID:   ct.ID,
				ClassTypeName: ct.Name,
				Themes:        themeDTOs,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if result == nil {
		result = []rotorCalendarDTO{}
	}
	json.NewEncoder(w).Encode(result)
}

// DTOs for rotor calendar API
type rotorCalendarDTO struct {
	RotorID       string                  `json:"rotor_id"`
	RotorName     string                  `json:"rotor_name"`
	ClassTypeID   string                  `json:"class_type_id"`
	ClassTypeName string                  `json:"class_type_name"`
	Themes        []rotorThemeCalendarDTO `json:"themes"`
}

type rotorThemeCalendarDTO struct {
	ThemeID   string            `json:"theme_id"`
	ThemeName string            `json:"theme_name"`
	Current   topicCalendarDTO  `json:"current"`
	Next      *topicCalendarDTO `json:"next,omitempty"`
}

type topicCalendarDTO struct {
	TopicID     string `json:"topic_id"`
	TopicName   string `json:"topic_name"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Description string `json:"description,omitempty"`
}
