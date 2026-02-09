package projections

import (
	"context"
	"time"

	"workshop/internal/adapters/storage/injury"
	"workshop/internal/adapters/storage/member"
	"workshop/internal/application/listutil"
	domainInjury "workshop/internal/domain/injury"
)

// GetMemberListQuery carries query parameters.
type GetMemberListQuery struct {
	Program string
	Search  string
	Status  string
	Sort    string
	Dir     string
	Page    int
	PerPage int
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
	Members  []MemberWithInjury
	PageInfo listutil.PageInfo
}

// GetMemberListDeps holds dependencies for GetMemberList.
type GetMemberListDeps struct {
	MemberStore MemberStore
	InjuryStore InjuryStore
}

// QueryGetMemberList retrieves members with injury flags.
// PRE: Valid query parameters
// POST: Returns paginated members with active injury indicators
// INVARIANT: Injuries are active if reported within last 7 days
func QueryGetMemberList(ctx context.Context, query GetMemberListQuery, deps GetMemberListDeps) (GetMemberListResult, error) {
	// Build the store filter from query params
	storeFilter := member.ListFilter{
		Program: query.Program,
		Status:  query.Status,
		Search:  query.Search,
		Sort:    query.Sort,
		Dir:     query.Dir,
	}

	// Get total count for pagination
	total, err := deps.MemberStore.Count(ctx, storeFilter)
	if err != nil {
		return GetMemberListResult{}, err
	}

	// Compute pagination
	perPage := query.PerPage
	if perPage <= 0 {
		perPage = listutil.DefaultPerPage
	}
	pageInfo := listutil.NewPageInfo(query.Page, perPage, total)

	// Set limit/offset for the store query
	storeFilter.Limit = pageInfo.PerPage
	storeFilter.Offset = pageInfo.Offset()

	// Get the page of members
	members, err := deps.MemberStore.List(ctx, storeFilter)
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

	return GetMemberListResult{Members: result, PageInfo: pageInfo}, nil
}
