package orchestrators

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	domain "workshop/internal/domain/calendar"
)

// CalendarEventStoreForSeed defines the store interface needed by SeedCompetitions.
type CalendarEventStoreForSeed interface {
	Save(ctx context.Context, e domain.Event) error
	ListByDateRange(ctx context.Context, from, to string) ([]domain.Event, error)
}

// SeedCompetitionsDeps holds dependencies for SeedCompetitions.
type SeedCompetitionsDeps struct {
	EventStore CalendarEventStoreForSeed
}

// CompetitionSeedData represents a competition to be seeded.
type CompetitionSeedData struct {
	Title           string
	StartDate       string // YYYY-MM-DD
	Location        string
	RegistrationURL string
}

// NZCompetitions2026 is the initial seed data for NZBJJF competitions in 2026.
var NZCompetitions2026 = []CompetitionSeedData{
	{
		Title:           "Kids JJ League Auckland Open II 2026",
		StartDate:       "2026-03-22",
		Location:        "Auckland",
		RegistrationURL: "https://nzbjjf.smoothcomp.com/en/event/27238",
	},
	{
		Title:           "Queenstown Open 2026",
		StartDate:       "2026-04-11",
		Location:        "Queenstown",
		RegistrationURL: "https://nzbjjf.smoothcomp.com/en/event/27243",
	},
	{
		Title:           "Feilding Open 2026",
		StartDate:       "2026-04-12",
		Location:        "Feilding",
		RegistrationURL: "https://nzbjjf.smoothcomp.com/en/event/27241",
	},
	{
		Title:           "Wellington Open 2026",
		StartDate:       "2026-05-24",
		Location:        "Wellington",
		RegistrationURL: "https://nzbjjf.smoothcomp.com/en/event/27235",
	},
	{
		Title:           "Kids & Youth JJ League Whangarei Open II 2026",
		StartDate:       "2026-05-31",
		Location:        "Whangarei",
		RegistrationURL: "https://nzbjjf.smoothcomp.com/en/event/27269",
	},
	{
		Title:           "Pacific Open Gi / No-Gi 2026",
		StartDate:       "2026-06-07",
		Location:        "Mount Maunganui",
		RegistrationURL: "https://nzbjjf.smoothcomp.com/en/event/27236",
	},
}

// ExecuteSeedCompetitions seeds NZ grappling competitions into the calendar.
// It is idempotent - skips any competition that already exists (matched by title + start_date).
func ExecuteSeedCompetitions(ctx context.Context, deps SeedCompetitionsDeps) error {
	// Load existing events to check for duplicates
	// Use a wide date range to catch all potential competitions
	existing, err := deps.EventStore.ListByDateRange(ctx, "2026-01-01", "2026-12-31")
	if err != nil {
		return err
	}

	// Build lookup map for idempotency: key = title + start_date
	existingMap := make(map[string]bool)
	for _, e := range existing {
		key := e.Title + e.StartDate.Format("2006-01-02")
		existingMap[key] = true
	}

	var seeded int
	for _, comp := range NZCompetitions2026 {
		// Check if already exists (idempotent)
		key := comp.Title + comp.StartDate
		if existingMap[key] {
			continue
		}

		startDate, err := time.Parse("2006-01-02", comp.StartDate)
		if err != nil {
			continue // Skip invalid dates
		}

		event := domain.Event{
			ID:              uuid.New().String(),
			Title:           comp.Title,
			Type:            "competition",
			Description:     "NZBJJF sanctioned competition",
			Location:        comp.Location,
			StartDate:       startDate,
			RegistrationURL: comp.RegistrationURL,
			CreatedBy:       "system",
			CreatedAt:       time.Now(),
		}

		if err := event.Validate(); err != nil {
			continue // Skip invalid events
		}

		if err := deps.EventStore.Save(ctx, event); err != nil {
			return err
		}
		seeded++
	}

	if seeded > 0 {
		slog.Info("seed_event", "event", "competitions_seeded", "count", seeded)
	}
	return nil
}
