package projections

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"workshop/internal/domain/milestone"
)

// CheckMilestonesMilestoneStore defines the milestone store interface for milestone checking.
type CheckMilestonesMilestoneStore interface {
	List(ctx context.Context) ([]milestone.Milestone, error)
}

// CheckMilestonesMemberMilestoneStore defines the member milestone store interface.
type CheckMilestonesMemberMilestoneStore interface {
	Save(ctx context.Context, value milestone.MemberMilestone) error
	ListByMemberID(ctx context.Context, memberID string) ([]milestone.MemberMilestone, error)
}

// CheckMilestonesDeps holds dependencies for the milestone check projection.
type CheckMilestonesDeps struct {
	MilestoneStore       CheckMilestonesMilestoneStore
	MemberMilestoneStore CheckMilestonesMemberMilestoneStore
}

// EarnedMilestone pairs a milestone definition with when it was earned.
type EarnedMilestone struct {
	Milestone milestone.Milestone
	EarnedAt  time.Time
	Notified  bool
}

// CheckMilestonesInput holds the training stats to check against milestones.
type CheckMilestonesInput struct {
	MemberID      string
	TotalClasses  int
	TotalMatHours float64
	CurrentStreak int
}

// QueryCheckMilestones evaluates all configured milestones against a member's stats
// and awards any newly earned milestones.
// PRE: input contains valid training stats
// POST: Returns all earned milestones (both new and existing), newly earned ones are saved
func QueryCheckMilestones(ctx context.Context, input CheckMilestonesInput, deps CheckMilestonesDeps) ([]EarnedMilestone, error) {
	allMilestones, err := deps.MilestoneStore.List(ctx)
	if err != nil {
		return nil, err
	}

	earned, err := deps.MemberMilestoneStore.ListByMemberID(ctx, input.MemberID)
	if err != nil {
		return nil, err
	}

	// Build a set of already-earned milestone IDs
	earnedSet := make(map[string]milestone.MemberMilestone)
	for _, memberMilestone := range earned {
		earnedSet[memberMilestone.MilestoneID] = memberMilestone
	}

	var results []EarnedMilestone
	now := time.Now()

	for _, ms := range allMilestones {
		var value float64
		switch ms.Metric {
		case milestone.MetricClasses:
			value = float64(input.TotalClasses)
		case milestone.MetricMatHours:
			value = input.TotalMatHours
		case milestone.MetricStreakWeeks:
			value = float64(input.CurrentStreak)
		default:
			continue
		}

		if value >= ms.Threshold {
			if existing, alreadyEarned := earnedSet[ms.ID]; alreadyEarned {
				results = append(results, EarnedMilestone{
					Milestone: ms,
					EarnedAt:  existing.EarnedAt,
					Notified:  existing.Notified,
				})
			} else {
				// Newly earned — save it
				newMM := milestone.MemberMilestone{
					ID:          generateMilestoneID(),
					MemberID:    input.MemberID,
					MilestoneID: ms.ID,
					EarnedAt:    now,
					Notified:    false,
				}
				if saveErr := deps.MemberMilestoneStore.Save(ctx, newMM); saveErr != nil {
					return nil, saveErr
				}
				results = append(results, EarnedMilestone{
					Milestone: ms,
					EarnedAt:  now,
					Notified:  false,
				})
			}
		}
	}

	return results, nil
}

// generateMilestoneID produces a unique ID for a member milestone.
func generateMilestoneID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return "mm-" + hex.EncodeToString(bytes)
}
