package browser_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	attendanceDomain "workshop/internal/domain/attendance"
	memberDomain "workshop/internal/domain/member"
	scheduleDomain "workshop/internal/domain/schedule"
)

// firstClassTypeID returns the ID of the first seeded class type.
func firstClassTypeID(t *testing.T, app *testApp) string {
	t.Helper()
	cts, err := app.Stores.ClassTypeStore.List(context.Background())
	if err != nil || len(cts) == 0 {
		t.Fatal("no seeded class types found")
	}
	return cts[0].ID
}

// TestMatHours_CheckInCreatesMatHours tests US-3.1.1: Attendance creates mat hours.
// Checking into a 90-minute class should record 1.5 mat hours.
func TestMatHours_CheckInCreatesMatHours(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	// Seed a member
	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "Mat Hours Tester", Email: "mathours@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, member); err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	// Seed a schedule (90-minute class: 18:00 - 19:30)
	sched := scheduleDomain.Schedule{
		ID:          uuid.New().String(),
		ClassTypeID: firstClassTypeID(t, app),
		Day:         scheduleDomain.Monday,
		StartTime:   "18:00",
		EndTime:     "19:30",
	}
	if err := app.Stores.ScheduleStore.Save(ctx, sched); err != nil {
		t.Fatalf("failed to seed schedule: %v", err)
	}

	// Check in via API with the schedule ID
	checkinBody := fmt.Sprintf(`{"MemberID":"%s","ScheduleID":"%s"}`, member.ID, sched.ID)
	resp, err := http.Post(app.BaseURL+"/checkin", "application/json", strings.NewReader(checkinBody))
	if err != nil {
		t.Fatalf("failed to check in: %v", err)
	}
	resp.Body.Close()

	// Verify mat hours on the attendance record
	resp, err = http.Get(app.BaseURL + "/api/attendance/member?member_id=" + member.ID)
	if err != nil {
		t.Fatalf("failed to get attendance: %v", err)
	}
	var checkins []attendanceDomain.Attendance
	json.NewDecoder(resp.Body).Decode(&checkins)
	resp.Body.Close()

	if len(checkins) != 1 {
		t.Fatalf("expected 1 check-in, got %d", len(checkins))
	}

	// 90 minutes = 1.5 hours
	if math.Abs(checkins[0].MatHours-1.5) > 0.01 {
		t.Errorf("expected 1.5 mat hours, got %.2f", checkins[0].MatHours)
	}
}

// TestMatHours_CheckInWithoutSchedule tests that check-in without a schedule records 0 mat hours.
func TestMatHours_CheckInWithoutSchedule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	// Seed a member
	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "No Schedule Tester", Email: "nosched@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, member); err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	// Check in without a schedule ID
	checkinBody := fmt.Sprintf(`{"MemberID":"%s"}`, member.ID)
	resp, err := http.Post(app.BaseURL+"/checkin", "application/json", strings.NewReader(checkinBody))
	if err != nil {
		t.Fatalf("failed to check in: %v", err)
	}
	resp.Body.Close()

	// Verify mat hours = 0
	resp, err = http.Get(app.BaseURL + "/api/attendance/member?member_id=" + member.ID)
	if err != nil {
		t.Fatalf("failed to get attendance: %v", err)
	}
	var checkins []attendanceDomain.Attendance
	json.NewDecoder(resp.Body).Decode(&checkins)
	resp.Body.Close()

	if len(checkins) != 1 {
		t.Fatalf("expected 1 check-in, got %d", len(checkins))
	}

	if checkins[0].MatHours != 0 {
		t.Errorf("expected 0 mat hours without schedule, got %.2f", checkins[0].MatHours)
	}
}

// TestMatHours_60MinuteClass tests a 60-minute class produces 1.0 mat hours.
func TestMatHours_60MinuteClass(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "Hour Tester", Email: "hour@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, member); err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	// 60-minute class: 6:00 PM - 7:00 PM (matches acceptance criteria)
	sched := scheduleDomain.Schedule{
		ID:          uuid.New().String(),
		ClassTypeID: firstClassTypeID(t, app),
		Day:         scheduleDomain.Monday,
		StartTime:   "18:00",
		EndTime:     "19:00",
	}
	if err := app.Stores.ScheduleStore.Save(ctx, sched); err != nil {
		t.Fatalf("failed to seed schedule: %v", err)
	}

	checkinBody := fmt.Sprintf(`{"MemberID":"%s","ScheduleID":"%s"}`, member.ID, sched.ID)
	resp, err := http.Post(app.BaseURL+"/checkin", "application/json", strings.NewReader(checkinBody))
	if err != nil {
		t.Fatalf("failed to check in: %v", err)
	}
	resp.Body.Close()

	resp, err = http.Get(app.BaseURL + "/api/attendance/member?member_id=" + member.ID)
	if err != nil {
		t.Fatalf("failed to get attendance: %v", err)
	}
	var checkins []attendanceDomain.Attendance
	json.NewDecoder(resp.Body).Decode(&checkins)
	resp.Body.Close()

	if len(checkins) != 1 {
		t.Fatalf("expected 1 check-in, got %d", len(checkins))
	}

	// 60 minutes = 1.0 hour — matches acceptance criteria:
	// "Given I check into '6:00 PM Nuts and Bolts' (60 min), Then 1 hour is added"
	if math.Abs(checkins[0].MatHours-1.0) > 0.01 {
		t.Errorf("expected 1.0 mat hours, got %.2f", checkins[0].MatHours)
	}
}
