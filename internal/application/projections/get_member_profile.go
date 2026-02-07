package projections

import (
	"context"
	"time"

	"workshop/internal/adapters/storage/attendance"
	"workshop/internal/adapters/storage/injury"
	"workshop/internal/adapters/storage/waiver"
	domainWaiver "workshop/internal/domain/waiver"
)

// GetMemberProfileQuery carries query parameters.
type GetMemberProfileQuery struct {
	MemberID string
}

// GetMemberProfileResult carries the query result.
type GetMemberProfileResult struct {
	MemberID         string
	Name             string
	Email            string
	Status           string
	Program          string
	HasValidWaiver   bool
	WaiverSignedAt   time.Time
	ActiveInjuries   []string // Body parts
	RecentAttendance int      // Count of check-ins in last 30 days
}

// GetMemberProfileDeps holds dependencies for GetMemberProfile.
type GetMemberProfileDeps struct {
	MemberStore     MemberStore
	WaiverStore     WaiverStore
	InjuryStore     InjuryStore
	AttendanceStore AttendanceStore
}

// QueryGetMemberProfile retrieves complete member profile.
// PRE: Valid member ID
// POST: Returns member details with waiver, injuries, and attendance history
func QueryGetMemberProfile(ctx context.Context, query GetMemberProfileQuery, deps GetMemberProfileDeps) (GetMemberProfileResult, error) {
	// Get member
	m, err := deps.MemberStore.GetByID(ctx, query.MemberID)
	if err != nil {
		return GetMemberProfileResult{}, err
	}

	result := GetMemberProfileResult{
		MemberID: m.ID,
		Name:     m.Name,
		Email:    m.Email,
		Status:   m.Status,
		Program:  m.Program,
	}

	// Get latest waiver
	waivers, err := deps.WaiverStore.List(ctx, waiver.ListFilter{
		Limit:  10,
		Offset: 0,
	})
	if err == nil {
		// Find most recent waiver for this member
		var latestWaiver *domainWaiver.Waiver
		for i := range waivers {
			w := &waivers[i]
			if w.MemberID == query.MemberID {
				if latestWaiver == nil || w.SignedAt.After(latestWaiver.SignedAt) {
					latestWaiver = w
				}
			}
		}
		if latestWaiver != nil {
			result.WaiverSignedAt = latestWaiver.SignedAt
			// Check if still valid (1 year)
			result.HasValidWaiver = time.Since(latestWaiver.SignedAt) < 365*24*time.Hour
		}
	}

	// Get active injuries
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	injuries, err := deps.InjuryStore.List(ctx, injury.ListFilter{
		Limit:  100,
		Offset: 0,
	})
	if err == nil {
		for _, inj := range injuries {
			if inj.MemberID == query.MemberID && inj.ReportedAt.After(sevenDaysAgo) {
				result.ActiveInjuries = append(result.ActiveInjuries, inj.BodyPart)
			}
		}
	}

	// Get recent attendance (last 30 days)
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	attendances, err := deps.AttendanceStore.List(ctx, attendance.ListFilter{
		Limit:  100,
		Offset: 0,
	})
	if err == nil {
		for _, a := range attendances {
			if a.MemberID == query.MemberID && a.CheckInTime.After(thirtyDaysAgo) {
				result.RecentAttendance++
			}
		}
	}

	return result, nil
}
