package orchestrators

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"workshop/internal/domain/account"
	"workshop/internal/domain/attendance"
	"workshop/internal/domain/classtype"
	"workshop/internal/domain/clip"
	"workshop/internal/domain/grading"
	"workshop/internal/domain/holiday"
	"workshop/internal/domain/injury"
	"workshop/internal/domain/member"
	"workshop/internal/domain/message"
	"workshop/internal/domain/milestone"
	"workshop/internal/domain/notice"
	"workshop/internal/domain/observation"
	"workshop/internal/domain/schedule"
	"workshop/internal/domain/term"
	"workshop/internal/domain/theme"
	"workshop/internal/domain/traininggoal"
	"workshop/internal/domain/waiver"

	memberFilter "workshop/internal/adapters/storage/member"

	"github.com/google/uuid"
)

// SyntheticSeedDeps holds all stores needed for synthetic data seeding.
type SyntheticSeedDeps struct {
	AccountStore         synAccountStore
	MemberStore          synMemberStore
	WaiverStore          synWaiverStore
	InjuryStore          synInjuryStore
	AttendanceStore      synAttendanceStore
	ScheduleStore        synScheduleStore
	TermStore            synTermStore
	HolidayStore         synHolidayStore
	NoticeStore          synNoticeStore
	GradingRecordStore   synGradingRecordStore
	GradingConfigStore   synGradingConfigStore
	GradingProposalStore synGradingProposalStore
	MessageStore         synMessageStore
	ObservationStore     synObservationStore
	MilestoneStore       synMilestoneStore
	TrainingGoalStore    synTrainingGoalStore
	ClassTypeStore       synClassTypeStore
	ThemeStore           synThemeStore
	ClipStore            synClipStore
}

type synMemberStore interface {
	Save(ctx context.Context, m member.Member) error
	List(ctx context.Context, filter memberFilter.ListFilter) ([]member.Member, error)
}
type synWaiverStore interface {
	Save(ctx context.Context, w waiver.Waiver) error
}
type synInjuryStore interface {
	Save(ctx context.Context, i injury.Injury) error
}
type synAttendanceStore interface {
	Save(ctx context.Context, a attendance.Attendance) error
}
type synScheduleStore interface {
	Save(ctx context.Context, s schedule.Schedule) error
	List(ctx context.Context) ([]schedule.Schedule, error)
}
type synTermStore interface {
	Save(ctx context.Context, t term.Term) error
	List(ctx context.Context) ([]term.Term, error)
}
type synHolidayStore interface {
	Save(ctx context.Context, h holiday.Holiday) error
	List(ctx context.Context) ([]holiday.Holiday, error)
}
type synNoticeStore interface {
	Save(ctx context.Context, n notice.Notice) error
}
type synGradingRecordStore interface {
	Save(ctx context.Context, r grading.Record) error
}
type synGradingConfigStore interface {
	Save(ctx context.Context, c grading.Config) error
	List(ctx context.Context) ([]grading.Config, error)
}
type synGradingProposalStore interface {
	Save(ctx context.Context, p grading.Proposal) error
}
type synMessageStore interface {
	Save(ctx context.Context, m message.Message) error
}
type synObservationStore interface {
	Save(ctx context.Context, o observation.Observation) error
}
type synMilestoneStore interface {
	Save(ctx context.Context, m milestone.Milestone) error
	List(ctx context.Context) ([]milestone.Milestone, error)
}
type synTrainingGoalStore interface {
	Save(ctx context.Context, g traininggoal.TrainingGoal) error
}
type synAccountStore interface {
	Save(ctx context.Context, a account.Account) error
	GetByEmail(ctx context.Context, email string) (account.Account, error)
}
type synClassTypeStore interface {
	List(ctx context.Context) ([]classtype.ClassType, error)
}
type synThemeStore interface {
	Save(ctx context.Context, t theme.Theme) error
	List(ctx context.Context) ([]theme.Theme, error)
}
type synClipStore interface {
	Save(ctx context.Context, c clip.Clip) error
}

// ExecuteSeedSynthetic populates the database with realistic BJJ school data.
// It is idempotent — skips if members already exist beyond the initial seed.
func ExecuteSeedSynthetic(ctx context.Context, deps SyntheticSeedDeps, adminAccountID string) error {
	existing, err := deps.MemberStore.List(ctx, memberFilter.ListFilter{Limit: 100, Offset: 0})
	if err != nil {
		return fmt.Errorf("seed_synthetic: list members: %w", err)
	}
	now := time.Now()

	// --- Layer 2: Themes & Clips (runs independently of member seeding) ---
	if deps.ThemeStore != nil && deps.ClipStore != nil {
		existingThemes, _ := deps.ThemeStore.List(ctx)
		if len(existingThemes) == 0 {
			// Need a coach ID for CreatedBy — try to find existing coach
			coachForThemes := adminAccountID
			existingCoach, coachErr := deps.AccountStore.GetByEmail(ctx, "coach@workshop.co.nz")
			if coachErr == nil {
				coachForThemes = existingCoach.ID
			}
			if err := seedThemesAndClips(ctx, deps, now, coachForThemes); err != nil {
				return err
			}
		}
	}

	if len(existing) > 5 {
		slog.Info("seed_event", "event", "synthetic_skip", "reason", "already_seeded")
		return nil
	}

	// --- Coach account ---
	coachAccountID := ""
	existingCoach, coachErr := deps.AccountStore.GetByEmail(ctx, "coach@workshop.co.nz")
	if coachErr != nil {
		coachAcct := account.Account{
			ID:        uuid.New().String(),
			Email:     "coach@workshop.co.nz",
			Role:      account.RoleCoach,
			CreatedAt: now,
		}
		if err := coachAcct.SetPassword("workshop12345!"); err != nil {
			return fmt.Errorf("seed coach password: %w", err)
		}
		if err := deps.AccountStore.Save(ctx, coachAcct); err != nil {
			return fmt.Errorf("seed coach account: %w", err)
		}
		coachAccountID = coachAcct.ID
		slog.Info("seed_event", "event", "coach_account_created", "email", "coach@workshop.co.nz")
	} else {
		coachAccountID = existingCoach.ID
	}

	// --- Member account (for testing member login) ---
	_, memberAcctErr := deps.AccountStore.GetByEmail(ctx, "marcus@email.com")
	if memberAcctErr != nil {
		memberAcct := account.Account{
			ID:        uuid.New().String(),
			Email:     "marcus@email.com",
			Role:      account.RoleMember,
			CreatedAt: now,
		}
		if err := memberAcct.SetPassword("workshop12345!"); err != nil {
			return fmt.Errorf("seed member password: %w", err)
		}
		if err := deps.AccountStore.Save(ctx, memberAcct); err != nil {
			return fmt.Errorf("seed member account: %w", err)
		}
		slog.Info("seed_event", "event", "member_account_created", "email", "marcus@email.com")
	}

	// --- Class types (already seeded, fetch IDs) ---
	classTypes, err := deps.ClassTypeStore.List(ctx)
	if err != nil {
		return fmt.Errorf("seed_synthetic: list class types: %w", err)
	}
	ctMap := map[string]string{}
	for _, ct := range classTypes {
		ctMap[ct.Name] = ct.ID
	}

	// --- Members: realistic BJJ school roster ---
	type memberSeed struct {
		Name    string
		Email   string
		Program string
		Belt    string
	}
	roster := []memberSeed{
		{"Marcus Oliveira", "marcus@email.com", "adults", "purple"},
		{"Sarah Chen", "sarah.chen@email.com", "adults", "blue"},
		{"Tane Patel", "tane.p@email.com", "adults", "blue"},
		{"Emily Rodriguez", "emily.r@email.com", "adults", "white"},
		{"James Mitchell", "james.m@email.com", "adults", "white"},
		{"Aroha Williams", "aroha.w@email.com", "adults", "white"},
		{"Dave Thompson", "dave.t@email.com", "adults", "blue"},
		{"Mika Tanaka", "mika.t@email.com", "adults", "white"},
		{"Liam O'Brien", "liam.ob@email.com", "adults", "purple"},
		{"Ngaire Henare", "ngaire.h@email.com", "adults", "white"},
		{"Ruby Mackenzie", "ruby.m@email.com", "kids", "yellow"},
		{"Finn Mackenzie", "finn.m@email.com", "kids", "grey"},
		{"Aiden Shaw", "aiden.s@email.com", "kids", "orange"},
	}

	memberIDs := make([]string, len(roster))
	for i, ms := range roster {
		id := uuid.New().String()
		memberIDs[i] = id
		m := member.Member{
			ID:      id,
			Name:    ms.Name,
			Email:   ms.Email,
			Program: ms.Program,
			Status:  member.StatusActive,
		}
		if err := deps.MemberStore.Save(ctx, m); err != nil {
			return fmt.Errorf("seed member %s: %w", ms.Name, err)
		}
	}

	// --- Waivers for all members (signed within last year) ---
	for i, id := range memberIDs {
		daysAgo := 30 + (i * 25) // stagger signing dates
		w := waiver.Waiver{
			ID:            uuid.New().String(),
			MemberID:      id,
			AcceptedTerms: true,
			SignedAt:      now.AddDate(0, 0, -daysAgo),
			IPAddress:     "192.168.1.1",
		}
		if err := deps.WaiverStore.Save(ctx, w); err != nil {
			return fmt.Errorf("seed waiver: %w", err)
		}
	}

	// --- Weekly schedule (realistic timetable) ---
	type schedSeed struct {
		Day       string
		Start     string
		End       string
		ClassName string
	}
	timetable := []schedSeed{
		{"monday", "06:00", "07:30", "Fundamentals"},
		{"monday", "12:00", "13:00", "No-Gi"},
		{"monday", "17:00", "18:00", "Kids Fundamentals"},
		{"tuesday", "06:00", "07:30", "No-Gi"},
		{"tuesday", "18:00", "19:30", "Competition"},
		{"wednesday", "06:00", "07:30", "Fundamentals"},
		{"wednesday", "17:00", "18:00", "Kids Advanced"},
		{"wednesday", "18:00", "19:30", "Fundamentals"},
		{"thursday", "06:00", "07:30", "No-Gi"},
		{"thursday", "18:00", "19:30", "Competition"},
		{"friday", "06:00", "07:30", "Fundamentals"},
		{"friday", "16:00", "17:30", "Open Mat"},
		{"saturday", "09:00", "10:00", "Kids Fundamentals"},
		{"saturday", "10:00", "11:30", "Fundamentals"},
	}

	existingScheds, _ := deps.ScheduleStore.List(ctx)
	scheduleIDs := make(map[string]string) // "Day-Start" -> ID
	for _, s := range existingScheds {
		scheduleIDs[s.Day+"-"+s.StartTime] = s.ID
	}

	for _, ts := range timetable {
		key := ts.Day + "-" + ts.Start
		if _, exists := scheduleIDs[key]; exists {
			continue
		}
		ctID := ctMap[ts.ClassName]
		if ctID == "" {
			continue
		}
		id := uuid.New().String()
		s := schedule.Schedule{
			ID:          id,
			ClassTypeID: ctID,
			Day:         ts.Day,
			StartTime:   ts.Start,
			EndTime:     ts.End,
		}
		if err := deps.ScheduleStore.Save(ctx, s); err != nil {
			return fmt.Errorf("seed schedule: %w", err)
		}
		scheduleIDs[key] = id
	}

	// --- NZ School Terms 2026 ---
	existingTerms, _ := deps.TermStore.List(ctx)
	if len(existingTerms) == 0 {
		terms := []term.Term{
			{ID: uuid.New().String(), Name: "Term 1 2026", StartDate: time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 4, 17, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New().String(), Name: "Term 2 2026", StartDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New().String(), Name: "Term 3 2026", StartDate: time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New().String(), Name: "Term 4 2026", StartDate: time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC)},
		}
		for _, t := range terms {
			if err := deps.TermStore.Save(ctx, t); err != nil {
				return fmt.Errorf("seed term: %w", err)
			}
		}
	}

	// --- NZ Public Holidays 2026 ---
	existingHols, _ := deps.HolidayStore.List(ctx)
	if len(existingHols) == 0 {
		holidays := []holiday.Holiday{
			{ID: uuid.New().String(), Name: "Waitangi Day", StartDate: time.Date(2026, 2, 6, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 2, 6, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New().String(), Name: "Good Friday", StartDate: time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New().String(), Name: "Easter Monday", StartDate: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New().String(), Name: "ANZAC Day", StartDate: time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New().String(), Name: "Queen's Birthday", StartDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New().String(), Name: "Matariki", StartDate: time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New().String(), Name: "Inter-term Break 1", StartDate: time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)},
		}
		for _, h := range holidays {
			if err := deps.HolidayStore.Save(ctx, h); err != nil {
				return fmt.Errorf("seed holiday: %w", err)
			}
		}
	}

	// --- Attendance: spread over last 60 days, realistic frequency ---
	// Marcus (purple belt) trains 5x/week, Sarah (blue) 4x, Emily (white) 2x, etc.
	trainFreq := []int{5, 4, 4, 2, 3, 2, 3, 1, 4, 2, 3, 3, 2} // sessions per week per member
	schedKeys := []string{}
	for k := range scheduleIDs {
		schedKeys = append(schedKeys, k)
	}

	for i, memberID := range memberIDs {
		freq := trainFreq[i]
		program := roster[i].Program

		for daysBack := 60; daysBack >= 0; daysBack-- {
			date := now.AddDate(0, 0, -daysBack)
			weekday := date.Weekday()

			// Simple frequency: train on certain days based on frequency
			shouldTrain := false
			switch {
			case freq >= 5:
				shouldTrain = weekday != time.Sunday
			case freq >= 4:
				shouldTrain = weekday == time.Monday || weekday == time.Tuesday || weekday == time.Wednesday || weekday == time.Thursday
			case freq >= 3:
				shouldTrain = weekday == time.Monday || weekday == time.Wednesday || weekday == time.Friday
			case freq >= 2:
				shouldTrain = weekday == time.Monday || weekday == time.Thursday
			default:
				shouldTrain = weekday == time.Wednesday
			}

			if !shouldTrain {
				continue
			}

			// Pick an appropriate schedule based on day and program
			var schedID string
			dayStr := dayOfWeek(weekday)
			for _, sk := range schedKeys {
				if len(sk) > len(dayStr) && sk[:len(dayStr)] == dayStr {
					sid := scheduleIDs[sk]
					// Kids go to kids classes only
					if program == "kids" {
						if sk == dayStr+"-17:00" || sk == dayStr+"-09:00" {
							schedID = sid
							break
						}
					} else {
						if sk != dayStr+"-17:00" && sk != dayStr+"-09:00" {
							schedID = sid
							break
						}
					}
				}
			}
			if schedID == "" {
				continue
			}

			classDate := date.Format("2006-01-02")
			checkinTime := date.Add(6 * time.Hour)
			checkoutTime := checkinTime.Add(90 * time.Minute)

			a := attendance.Attendance{
				ID:           uuid.New().String(),
				MemberID:     memberID,
				ScheduleID:   schedID,
				ClassDate:    classDate,
				CheckInTime:  checkinTime,
				CheckOutTime: checkoutTime,
			}
			if err := deps.AttendanceStore.Save(ctx, a); err != nil {
				return fmt.Errorf("seed attendance: %w", err)
			}
		}
	}

	// --- Active injuries ---
	injuries := []struct {
		memberIdx int
		bodyPart  string
		desc      string
		daysAgo   int
	}{
		{2, injury.BodyPartKnee, "Tweaked left knee during takedown drill", 3},
		{6, injury.BodyPartShoulder, "Sore right shoulder from kimura defense", 1},
		{0, injury.BodyPartRib, "Bruised rib from pressure passing", 5},
	}
	for _, inj := range injuries {
		i := injury.Injury{
			ID:          uuid.New().String(),
			MemberID:    memberIDs[inj.memberIdx],
			BodyPart:    inj.bodyPart,
			Description: inj.desc,
			ReportedAt:  now.AddDate(0, 0, -inj.daysAgo),
		}
		if err := deps.InjuryStore.Save(ctx, i); err != nil {
			return fmt.Errorf("seed injury: %w", err)
		}
	}

	// --- Observations (coach notes) ---
	obs := []struct {
		memberIdx int
		content   string
		daysAgo   int
	}{
		{0, "Excellent pressure from top half guard. Ready to start teaching fundamentals classes.", 2},
		{1, "Guard retention improving dramatically. Needs more work on submission chains from closed guard.", 5},
		{2, "Strong takedowns but needs to slow down rolling — getting caught in rushed submissions.", 3},
		{3, "Great attitude, always first to drill. Hip escapes getting much cleaner.", 7},
		{4, "Competed at Auckland Open — lost on points in semis. Good experience.", 10},
		{6, "Consistent training partner, technical game. Should consider competing.", 4},
		{8, "Teaching ability is excellent — consider for assistant instructor role.", 1},
		{10, "Really engaged in class, listens well. Ready for next stripe.", 6},
		{11, "Shy but improving. Pairs well with Ruby for drills.", 8},
		{12, "Natural athlete, picks up techniques quickly. Encourage more drilling.", 3},
	}
	for _, o := range obs {
		ob := observation.Observation{
			ID:        uuid.New().String(),
			MemberID:  memberIDs[o.memberIdx],
			AuthorID:  coachAccountID,
			Content:   o.content,
			CreatedAt: now.AddDate(0, 0, -o.daysAgo),
			UpdatedAt: now.AddDate(0, 0, -o.daysAgo),
		}
		if err := deps.ObservationStore.Save(ctx, ob); err != nil {
			return fmt.Errorf("seed observation: %w", err)
		}
	}

	// --- Notices ---
	notices := []notice.Notice{
		{ID: uuid.New().String(), Type: notice.TypeSchoolWide, Status: notice.StatusPublished, Title: "Grading Day — Saturday 15 Feb", Content: "Belt grading for all eligible members. Arrive 30 min early in clean gi.", CreatedBy: adminAccountID, PublishedBy: adminAccountID, CreatedAt: now.AddDate(0, 0, -5), PublishedAt: now.AddDate(0, 0, -5)},
		{ID: uuid.New().String(), Type: notice.TypeSchoolWide, Status: notice.StatusPublished, Title: "New Competition Class", Content: "Starting next week, Tuesday & Thursday 6pm. Open to blue belt and above.", CreatedBy: adminAccountID, PublishedBy: adminAccountID, CreatedAt: now.AddDate(0, 0, -3), PublishedAt: now.AddDate(0, 0, -3)},
		{ID: uuid.New().String(), Type: notice.TypeSchoolWide, Status: notice.StatusPublished, Title: "Open Mat Every Friday", Content: "Friday open mat 4-5:30pm. All ranks welcome. Bring water and mouthguard.", CreatedBy: adminAccountID, PublishedBy: adminAccountID, CreatedAt: now.AddDate(0, 0, -1), PublishedAt: now.AddDate(0, 0, -1)},
		{ID: uuid.New().String(), Type: notice.TypeClassSpecific, Status: notice.StatusPublished, Title: "Kids Grading Prep", Content: "Kids grading prep this Saturday. Parents welcome to watch.", CreatedBy: adminAccountID, PublishedBy: adminAccountID, CreatedAt: now.AddDate(0, 0, -2), PublishedAt: now.AddDate(0, 0, -2)},
		{ID: uuid.New().String(), Type: notice.TypeSchoolWide, Status: notice.StatusDraft, Title: "Gym Closure — Easter Break", Content: "Gym closed Good Friday through Easter Monday. Normal schedule resumes Tuesday.", CreatedBy: adminAccountID, CreatedAt: now},
	}
	for _, n := range notices {
		if err := deps.NoticeStore.Save(ctx, n); err != nil {
			return fmt.Errorf("seed notice: %w", err)
		}
	}

	// --- Grading configs ---
	existingConfigs, _ := deps.GradingConfigStore.List(ctx)
	if len(existingConfigs) <= 1 {
		configs := []grading.Config{
			{ID: uuid.New().String(), Program: "adults", Belt: grading.BeltBlue, FlightTimeHours: 150, StripeCount: 4},
			{ID: uuid.New().String(), Program: "adults", Belt: grading.BeltPurple, FlightTimeHours: 300, StripeCount: 4},
			{ID: uuid.New().String(), Program: "adults", Belt: grading.BeltBrown, FlightTimeHours: 500, StripeCount: 4},
			{ID: uuid.New().String(), Program: "adults", Belt: grading.BeltBlack, FlightTimeHours: 750, StripeCount: 4},
			{ID: uuid.New().String(), Program: "kids", Belt: grading.BeltGrey, FlightTimeHours: 0, AttendancePct: 80, StripeCount: 4},
			{ID: uuid.New().String(), Program: "kids", Belt: grading.BeltYellow, FlightTimeHours: 0, AttendancePct: 80, StripeCount: 4},
			{ID: uuid.New().String(), Program: "kids", Belt: grading.BeltOrange, FlightTimeHours: 0, AttendancePct: 80, StripeCount: 4},
		}
		for _, c := range configs {
			if err := deps.GradingConfigStore.Save(ctx, c); err != nil {
				return fmt.Errorf("seed grading config: %w", err)
			}
		}
	}

	// --- Grading records (historical promotions) ---
	records := []struct {
		memberIdx int
		belt      string
		daysAgo   int
	}{
		{0, grading.BeltBlue, 730},    // Marcus: white→blue 2yr ago
		{0, grading.BeltPurple, 180},  // Marcus: blue→purple 6mo ago
		{1, grading.BeltBlue, 365},    // Sarah: white→blue 1yr ago
		{2, grading.BeltBlue, 400},    // Tane: white→blue
		{6, grading.BeltBlue, 300},    // Dave: white→blue
		{8, grading.BeltBlue, 600},    // Liam: white→blue 1.5yr ago
		{8, grading.BeltPurple, 120},  // Liam: blue→purple 4mo ago
		{10, grading.BeltGrey, 200},   // Ruby: white→grey
		{10, grading.BeltYellow, 60},  // Ruby: grey→yellow
		{11, grading.BeltGrey, 90},    // Finn: white→grey
		{12, grading.BeltGrey, 300},   // Aiden: white→grey
		{12, grading.BeltYellow, 150}, // Aiden: grey→yellow
		{12, grading.BeltOrange, 30},  // Aiden: yellow→orange
	}
	for _, rec := range records {
		r := grading.Record{
			ID:         uuid.New().String(),
			MemberID:   memberIDs[rec.memberIdx],
			Belt:       rec.belt,
			Stripe:     0,
			PromotedAt: now.AddDate(0, 0, -rec.daysAgo),
			ProposedBy: coachAccountID,
			ApprovedBy: adminAccountID,
			Method:     grading.MethodStandard,
		}
		if err := deps.GradingRecordStore.Save(ctx, r); err != nil {
			return fmt.Errorf("seed grading record: %w", err)
		}
	}

	// --- Pending grading proposals ---
	proposals := []struct {
		memberIdx  int
		targetBelt string
		notes      string
	}{
		{1, grading.BeltPurple, "Sarah has 310+ mat hours, strong guard game, competed twice this year."},
		{2, grading.BeltPurple, "Tane is technically ready but needs more competition experience."},
		{3, grading.BeltBlue, "Emily has trained consistently 2x/week for 18 months. Solid fundamentals."},
	}
	for _, p := range proposals {
		pr := grading.Proposal{
			ID:         uuid.New().String(),
			MemberID:   memberIDs[p.memberIdx],
			TargetBelt: p.targetBelt,
			Notes:      p.notes,
			ProposedBy: coachAccountID,
			Status:     grading.ProposalPending,
			CreatedAt:  now.AddDate(0, 0, -2),
		}
		if err := deps.GradingProposalStore.Save(ctx, pr); err != nil {
			return fmt.Errorf("seed grading proposal: %w", err)
		}
	}

	// --- Messages (admin to members) ---
	msgs := []struct {
		memberIdx int
		subject   string
		content   string
		daysAgo   int
		read      bool
	}{
		{3, "Welcome to the school!", "Hi Emily, welcome aboard! Let us know if you have any questions about the schedule or need help with anything.", 30, true},
		{1, "Grading nomination", "Hi Sarah, you've been nominated for purple belt grading on Feb 15. Please confirm you're available.", 3, false},
		{4, "Competition results", "Great effort at the Auckland Open, James! Two close matches. Let's work on your guard passing this week.", 8, true},
		{7, "We miss you!", "Hi Mika, we noticed you haven't trained in a while. Everything okay? We'd love to see you back on the mats.", 5, false},
		{0, "Instructor opportunity", "Marcus, we'd like to discuss you taking on a fundamentals teaching role. Interested?", 1, false},
		{5, "Fee reminder", "Hi Aroha, just a friendly reminder that your February membership fee is due. Please reach out if you need to discuss.", 2, false},
	}
	for _, m := range msgs {
		msg := message.Message{
			ID:         uuid.New().String(),
			SenderID:   adminAccountID,
			ReceiverID: memberIDs[m.memberIdx],
			Subject:    m.subject,
			Content:    m.content,
			CreatedAt:  now.AddDate(0, 0, -m.daysAgo),
		}
		if m.read {
			msg.ReadAt = now.AddDate(0, 0, -(m.daysAgo - 1))
		}
		if err := deps.MessageStore.Save(ctx, msg); err != nil {
			return fmt.Errorf("seed message: %w", err)
		}
	}

	// --- Milestones ---
	existingMs, _ := deps.MilestoneStore.List(ctx)
	if len(existingMs) == 0 {
		milestones := []milestone.Milestone{
			{ID: uuid.New().String(), Name: "First Class", Metric: milestone.MetricClasses, Threshold: 1, BadgeIcon: "🥋"},
			{ID: uuid.New().String(), Name: "Dedicated 10", Metric: milestone.MetricClasses, Threshold: 10, BadgeIcon: "⭐"},
			{ID: uuid.New().String(), Name: "Century Club", Metric: milestone.MetricClasses, Threshold: 100, BadgeIcon: "💯"},
			{ID: uuid.New().String(), Name: "Iron Will 250", Metric: milestone.MetricClasses, Threshold: 250, BadgeIcon: "🏆"},
			{ID: uuid.New().String(), Name: "50 Mat Hours", Metric: milestone.MetricMatHours, Threshold: 50, BadgeIcon: "🔥"},
			{ID: uuid.New().String(), Name: "200 Mat Hours", Metric: milestone.MetricMatHours, Threshold: 200, BadgeIcon: "💪"},
			{ID: uuid.New().String(), Name: "Month Streak", Metric: milestone.MetricStreakWeeks, Threshold: 4, BadgeIcon: "📅"},
			{ID: uuid.New().String(), Name: "Quarter Streak", Metric: milestone.MetricStreakWeeks, Threshold: 13, BadgeIcon: "🎯"},
		}
		for _, ms := range milestones {
			if err := deps.MilestoneStore.Save(ctx, ms); err != nil {
				return fmt.Errorf("seed milestone: %w", err)
			}
		}
	}

	// --- Training goals ---
	goals := []struct {
		memberIdx int
		target    int
		period    string
	}{
		{0, 5, traininggoal.PeriodWeekly},
		{1, 4, traininggoal.PeriodWeekly},
		{3, 3, traininggoal.PeriodWeekly},
		{4, 3, traininggoal.PeriodWeekly},
		{6, 3, traininggoal.PeriodWeekly},
	}
	for _, g := range goals {
		goal := traininggoal.TrainingGoal{
			ID:        uuid.New().String(),
			MemberID:  memberIDs[g.memberIdx],
			Target:    g.target,
			Period:    g.period,
			CreatedAt: now.AddDate(0, 0, -14),
			Active:    true,
		}
		if err := deps.TrainingGoalStore.Save(ctx, goal); err != nil {
			return fmt.Errorf("seed training goal: %w", err)
		}
	}

	slog.Info("seed_event", "event", "synthetic_seeded",
		"members", len(roster),
		"attendance_records", "~800",
		"observations", len(obs),
		"messages", len(msgs),
		"notices", len(notices),
	)
	return nil
}

func dayOfWeek(w time.Weekday) string {
	switch w {
	case time.Monday:
		return "monday"
	case time.Tuesday:
		return "tuesday"
	case time.Wednesday:
		return "wednesday"
	case time.Thursday:
		return "thursday"
	case time.Friday:
		return "friday"
	case time.Saturday:
		return "saturday"
	case time.Sunday:
		return "sunday"
	}
	return ""
}

// seedThemesAndClips creates realistic BJJ technical themes and study clips.
// PRE: deps.ThemeStore and deps.ClipStore are non-nil
// POST: 4 themes and 7 clips are persisted
func seedThemesAndClips(ctx context.Context, deps SyntheticSeedDeps, now time.Time, createdBy string) error {
	themeData := []struct {
		name   string
		desc   string
		prog   string
		offset int
	}{
		{"Leg Lasso Series", "Controlling distance and off-balancing from open guard using the leg lasso grip system.", "adults", -14},
		{"Closed Guard Attacks", "Cross-collar chokes, arm bars, and triangle setups from closed guard.", "adults", 14},
		{"Half Guard Sweeps", "Underhook recovery, deep half transitions, and the Lucas Leite sweep series.", "adults", -42},
		{"Animal Movements", "Fun warmups and coordination drills: bear crawls, shrimping, and technical stand-ups.", "kids", -7},
	}
	themeIDs := make([]string, len(themeData))
	for i, td := range themeData {
		themeIDs[i] = uuid.New().String()
		start := now.AddDate(0, 0, td.offset)
		t := theme.Theme{
			ID:          themeIDs[i],
			Name:        td.name,
			Description: td.desc,
			Program:     td.prog,
			StartDate:   start,
			EndDate:     start.AddDate(0, 0, 27),
			CreatedBy:   createdBy,
			CreatedAt:   now.AddDate(0, 0, td.offset-1),
		}
		if err := deps.ThemeStore.Save(ctx, t); err != nil {
			return fmt.Errorf("seed theme: %w", err)
		}
	}

	clipData := []struct {
		themeIdx int
		title    string
		url      string
		start    int
		end      int
		notes    string
		promoted bool
	}{
		{0, "Lachlan Giles Leg Lasso Sweep", "https://www.youtube.com/watch?v=JxDl1yvMLj0", 45, 78, "Watch the hip angle when entering lasso", true},
		{0, "Leg Lasso to Omoplata", "https://www.youtube.com/watch?v=JxDl1yvMLj0", 120, 155, "Key detail: control the sleeve before inverting", true},
		{0, "Lasso Guard Retention", "https://www.youtube.com/watch?v=JxDl1yvMLj0", 200, 230, "Re-lasso after guard pass attempt", false},
		{2, "Deep Half Entry from Half Guard", "https://www.youtube.com/watch?v=eVDCKbWRnOA", 30, 60, "Duck under the crossface and scoop the far leg", true},
		{2, "Lucas Leite Sweep", "https://www.youtube.com/watch?v=eVDCKbWRnOA", 90, 125, "Classic sweep from deep half — waiter sweep variation", true},
		{1, "Roger Gracie Cross-Collar Choke", "https://www.youtube.com/watch?v=2o-YDvHbfl4", 15, 50, "Posture break, first grip deep, elbow to the mat", false},
		{3, "Bear Crawl Technique", "https://www.youtube.com/watch?v=CLJGMi3NFWI", 10, 35, "Keep hips low, opposite hand and foot", true},
	}
	for _, cd := range clipData {
		c := clip.Clip{
			ID:           uuid.New().String(),
			ThemeID:      themeIDs[cd.themeIdx],
			Title:        cd.title,
			YouTubeURL:   cd.url,
			StartSeconds: cd.start,
			EndSeconds:   cd.end,
			Notes:        cd.notes,
			CreatedBy:    createdBy,
			Promoted:     cd.promoted,
			CreatedAt:    now.AddDate(0, 0, -3),
		}
		if cd.promoted {
			c.PromotedBy = createdBy
		}
		_ = c.ExtractYouTubeID()
		if err := deps.ClipStore.Save(ctx, c); err != nil {
			return fmt.Errorf("seed clip: %w", err)
		}
	}
	slog.Info("seed_event", "event", "themes_clips_seeded", "themes", len(themeData), "clips", len(clipData))
	return nil
}
