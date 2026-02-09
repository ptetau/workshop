package projections

import (
	"context"

	"workshop/internal/adapters/storage/attendance"
	"workshop/internal/adapters/storage/injury"
	"workshop/internal/adapters/storage/member"
	"workshop/internal/adapters/storage/waiver"
	domainAttendance "workshop/internal/domain/attendance"
	domainGrading "workshop/internal/domain/grading"
	domainInjury "workshop/internal/domain/injury"
	domainMember "workshop/internal/domain/member"
	domainWaiver "workshop/internal/domain/waiver"
)

// MemberStore interface for member queries.
type MemberStore interface {
	GetByID(ctx context.Context, id string) (domainMember.Member, error)
	List(ctx context.Context, filter member.ListFilter) ([]domainMember.Member, error)
	Count(ctx context.Context, filter member.ListFilter) (int, error)
}

// WaiverStore interface for waiver queries.
type WaiverStore interface {
	List(ctx context.Context, filter waiver.ListFilter) ([]domainWaiver.Waiver, error)
}

// InjuryStore interface for injury queries.
type InjuryStore interface {
	List(ctx context.Context, filter injury.ListFilter) ([]domainInjury.Injury, error)
}

// AttendanceStore interface for attendance queries.
type AttendanceStore interface {
	List(ctx context.Context, filter attendance.ListFilter) ([]domainAttendance.Attendance, error)
}

// GradingRecordStore interface for grading record queries.
type GradingRecordStore interface {
	ListByMemberID(ctx context.Context, memberID string) ([]domainGrading.Record, error)
}
