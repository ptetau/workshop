package projections

import (
	"context"
	"time"

	"workshop/internal/adapters/storage/injury"
	"workshop/internal/adapters/storage/member"
	domainInjury "workshop/internal/domain/injury"
)

// GetMemberListQuery carries query parameters.
type GetMemberListQuery struct {
	Program string
}

// MemberWithInjury represents a member with injury status.
type MemberWithInjury struct {
	ID             string
	Name           string
	Email          string
	Program        string
	Status         string
	HasInjury      bool
	InjuryBodyPart string
}

// GetMemberListResult carries the query result.
type GetMemberListResult struct {
	Members []MemberWithInjury
}

// GetMemberListDeps holds dependencies for GetMemberList.
type GetMemberListDeps struct {
	MemberStore MemberStore
	InjuryStore InjuryStore
}

// QueryGetMemberList retrieves members with injury flags.
// PRE: Valid query parameters
// POST: Returns members filtered by program with active injury indicators
// INVARIANT: Injuries are active if reported within last 7 days
func QueryGetMemberList(ctx context.Context, query GetMemberListQuery, deps GetMemberListDeps) (GetMemberListResult, error) {
	// Get all members (or filter by program if specified)
	members, err := deps.MemberStore.List(ctx, member.ListFilter{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		return GetMemberListResult{}, err
	}

	// Get all active injuries (last 7 days)
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	injuries, err := deps.InjuryStore.List(ctx, injury.ListFilter{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return GetMemberListResult{}, err
	}

	// Build injury map for quick lookup
	injuryMap := make(map[string]domainInjury.Injury)
	for _, inj := range injuries {
		// Only include active injuries (within 7 days)
		if inj.ReportedAt.After(sevenDaysAgo) {
			injuryMap[inj.MemberID] = inj
		}
	}

	// Build result with injury flags
	var result []MemberWithInjury
	for _, m := range members {
		// Filter by program if specified
		if query.Program != "" && m.Program != query.Program {
			continue
		}

		mwi := MemberWithInjury{
			ID:      m.ID,
			Name:    m.Name,
			Email:   m.Email,
			Program: m.Program,
			Status:  m.Status,
		}

		// Check for active injury
		if inj, hasInjury := injuryMap[m.ID]; hasInjury {
			mwi.HasInjury = true
			mwi.InjuryBodyPart = inj.BodyPart
		}

		result = append(result, mwi)
	}

	return GetMemberListResult{Members: result}, nil
}
