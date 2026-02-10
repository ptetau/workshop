package projections

import (
	"context"
	"time"

	"workshop/internal/domain/grading"
	"workshop/internal/domain/member"
	"workshop/internal/domain/notice"
	"workshop/internal/domain/traininggoal"
	"workshop/internal/domain/waiver"
)

// DashboardNoticeStore defines the notice store interface needed by the dashboard projection.
type DashboardNoticeStore interface {
	ListPublished(ctx context.Context, noticeType string, now time.Time) ([]notice.Notice, error)
}

// DashboardProposalStore defines the grading proposal store interface needed by the dashboard projection.
type DashboardProposalStore interface {
	ListPending(ctx context.Context) ([]grading.Proposal, error)
}

// DashboardMessageStore defines the message store interface needed by the dashboard projection.
type DashboardMessageStore interface {
	CountUnread(ctx context.Context, receiverID string) (int, error)
}

// DashboardTrainingGoalStore defines the training goal store interface needed by the dashboard projection.
type DashboardTrainingGoalStore interface {
	GetActiveByMemberID(ctx context.Context, memberID string) (traininggoal.TrainingGoal, error)
}

// DashboardMemberStore defines the member store interface needed by the dashboard projection.
type DashboardMemberStore interface {
	GetByEmail(ctx context.Context, email string) (member.Member, error)
}

// DashboardWaiverStore defines the waiver store interface needed by the dashboard projection.
type DashboardWaiverStore interface {
	GetByMemberID(ctx context.Context, memberID string) (waiver.Waiver, error)
}

// GetDashboardQuery carries input for the dashboard projection.
type GetDashboardQuery struct {
	Role         string // admin, coach, member, trial
	AccountEmail string // used to resolve MemberID for member/trial role
}

// GetDashboardDeps holds dependencies for the dashboard projection.
type GetDashboardDeps struct {
	TodaysClassesDeps  GetTodaysClassesDeps
	AttendanceDeps     GetAttendanceTodayDeps
	InactiveDeps       GetInactiveMembersDeps
	TrainingLogDeps    GetTrainingLogDeps
	NoticeStore        DashboardNoticeStore
	ProposalStore      DashboardProposalStore
	MessageStore       DashboardMessageStore
	TrainingGoalStore  DashboardTrainingGoalStore
	MemberStore        DashboardMemberStore
	GradingRecordStore GradingRecordStore   // optional: nil skips belt lookup
	WaiverStore        DashboardWaiverStore // optional: nil skips waiver check
}

// DashboardResult carries the output of the dashboard projection.
type DashboardResult struct {
	Role         string
	WaiverSigned bool // added field

	// Shared
	TodaysClasses []TodaysClassResult
	Notices       []notice.Notice

	// Admin
	PendingProposals int
	InactiveCount    int

	// Coach
	Attendees []AttendanceWithMember

	// Member
	TrainingLog  *TrainingLogResult
	UnreadCount  int
	TrainingGoal *traininggoal.TrainingGoal
	Belt         string
	Stripe       int
	IsTrial      bool
}

// QueryGetDashboard aggregates dashboard data based on the user's role.
func QueryGetDashboard(ctx context.Context, query GetDashboardQuery, deps GetDashboardDeps, now time.Time) (DashboardResult, error) {
	result := DashboardResult{Role: query.Role}

	// All roles: today's classes
	classes, err := QueryGetTodaysClasses(ctx, now, deps.TodaysClassesDeps)
	if err == nil {
		result.TodaysClasses = classes
	}

	// All roles: published school-wide notices
	notices, err := deps.NoticeStore.ListPublished(ctx, "school_wide", now)
	if err == nil {
		result.Notices = notices
	}

	switch query.Role {
	case "admin":
		// Pending grading proposals
		proposals, err := deps.ProposalStore.ListPending(ctx)
		if err == nil {
			result.PendingProposals = len(proposals)
		}
		// Inactive members
		inactiveQuery := GetInactiveMembersQuery{DaysSinceLastCheckIn: 30}
		inactive, err := QueryGetInactiveMembers(ctx, inactiveQuery, deps.InactiveDeps)
		if err == nil {
			result.InactiveCount = len(inactive)
		}

	case "coach":
		// Today's attendance with injury flags
		attendanceQuery := GetAttendanceTodayQuery{}
		attendanceResult, err := QueryGetAttendanceToday(ctx, attendanceQuery, deps.AttendanceDeps)
		if err == nil {
			result.Attendees = attendanceResult.Attendees
		}

	case "member", "trial":
		if query.AccountEmail != "" {
			// Resolve member ID from account email
			memberRecord, err := deps.MemberStore.GetByEmail(ctx, query.AccountEmail)
			if err == nil && memberRecord.ID != "" {
				memberID := memberRecord.ID
				// Training log summary
				logQuery := GetTrainingLogQuery{MemberID: memberID}
				logResult, err := QueryGetTrainingLog(ctx, logQuery, deps.TrainingLogDeps)
				if err == nil {
					result.TrainingLog = &logResult
				}
				// Unread messages
				count, err := deps.MessageStore.CountUnread(ctx, memberID)
				if err == nil {
					result.UnreadCount = count
				}
				// Active training goal
				goal, err := deps.TrainingGoalStore.GetActiveByMemberID(ctx, memberID)
				if err == nil && goal.ID != "" {
					result.TrainingGoal = &goal
				}
				// Latest belt
				if deps.GradingRecordStore != nil {
					if records, err := deps.GradingRecordStore.ListByMemberID(ctx, memberID); err == nil && len(records) > 0 {
						latest := records[0]
						for _, r := range records[1:] {
							if r.PromotedAt.After(latest.PromotedAt) {
								latest = r
							}
						}
						result.Belt = latest.Belt
						result.Stripe = latest.Stripe
					}
				}
				// Waiver status for trial users
				if query.Role == "trial" {
					result.IsTrial = true
					if deps.WaiverStore != nil {
						w, err := deps.WaiverStore.GetByMemberID(ctx, memberID)
						if err == nil && w.ID != "" {
							result.WaiverSigned = true
						}
					}
				}
			}
		}
	}

	return result, nil
}
