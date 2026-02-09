package browser_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"

	gradingDomain "workshop/internal/domain/grading"
	injuryDomain "workshop/internal/domain/injury"
	memberDomain "workshop/internal/domain/member"
	scheduleDomain "workshop/internal/domain/schedule"
)

// TestCoachAttendance_ShowsTodayCheckins tests US-3.1.2: Coach views today's attendance.
// Verifies that the /attendance page shows names, belt icons, mat hours, and injury flags.
func TestCoachAttendance_ShowsTodayCheckins(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	// Seed class type ID from seeded data
	cts, err := app.Stores.ClassTypeStore.List(ctx)
	if err != nil || len(cts) == 0 {
		t.Fatal("no seeded class types found")
	}
	ctID := cts[0].ID

	// Seed a 90-minute schedule
	sched := scheduleDomain.Schedule{
		ID: uuid.New().String(), ClassTypeID: ctID,
		Day: scheduleDomain.Monday, StartTime: "18:00", EndTime: "19:30",
	}
	if err := app.Stores.ScheduleStore.Save(ctx, sched); err != nil {
		t.Fatalf("failed to seed schedule: %v", err)
	}

	// Seed members
	alice := memberDomain.Member{
		ID: uuid.New().String(), Name: "Alice Tester", Email: "alice@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	bob := memberDomain.Member{
		ID: uuid.New().String(), Name: "Bob Tester", Email: "bob@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	for _, m := range []memberDomain.Member{alice, bob} {
		if err := app.Stores.MemberStore.Save(ctx, m); err != nil {
			t.Fatalf("failed to seed member %s: %v", m.Name, err)
		}
	}

	// Give Alice a blue belt with 2 stripes
	gradingRecord := gradingDomain.Record{
		ID: uuid.New().String(), MemberID: alice.ID,
		Belt: gradingDomain.BeltBlue, Stripe: 2,
		PromotedAt: time.Now().Add(-30 * 24 * time.Hour),
		ProposedBy: "coach-1", ApprovedBy: "admin-1", Method: gradingDomain.MethodStandard,
	}
	if err := app.Stores.GradingRecordStore.Save(ctx, gradingRecord); err != nil {
		t.Fatalf("failed to seed grading record: %v", err)
	}

	// Give Bob an injury (recent)
	inj := injuryDomain.Injury{
		ID: uuid.New().String(), MemberID: bob.ID,
		BodyPart: "left knee", Description: "sprain",
		ReportedAt: time.Now(),
	}
	if err := app.Stores.InjuryStore.Save(ctx, inj); err != nil {
		t.Fatalf("failed to seed injury: %v", err)
	}

	// Check in both members via API
	for _, m := range []memberDomain.Member{alice, bob} {
		body := fmt.Sprintf(`{"MemberID":"%s","ScheduleID":"%s"}`, m.ID, sched.ID)
		resp, err := http.Post(app.BaseURL+"/checkin", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("failed to check in %s: %v", m.Name, err)
		}
		resp.Body.Close()
	}

	// Navigate to attendance page
	if _, err := page.Goto(app.BaseURL + "/attendance"); err != nil {
		t.Fatal(err)
	}

	// Verify member count
	countText := page.Locator("p:has-text('member(s) checked in today')")
	text, err := countText.TextContent()
	if err != nil {
		t.Fatal("could not find attendance count text")
	}
	if !strings.Contains(text, "2") {
		t.Errorf("expected 2 members checked in, got: %s", text)
	}

	// Verify Alice's name appears
	aliceRow := page.Locator("tr:has-text('Alice Tester')")
	if visible, _ := aliceRow.IsVisible(); !visible {
		t.Error("Alice Tester should be visible in attendance table")
	}

	// Verify Alice's belt icon
	aliceBelt := aliceRow.Locator("[data-belt='blue']")
	if visible, _ := aliceBelt.IsVisible(); !visible {
		t.Error("Alice should have a blue belt icon")
	}

	// Verify Alice's stripe dots (2)
	aliceStripes := aliceRow.Locator(".stripe-dot")
	stripeCount, _ := aliceStripes.Count()
	if stripeCount != 2 {
		t.Errorf("expected 2 stripe dots for Alice, got %d", stripeCount)
	}

	// Verify mat hours column shows 1.5h (90 min)
	aliceMatHours := aliceRow.Locator("td:has-text('1.5h')")
	if visible, _ := aliceMatHours.IsVisible(); !visible {
		t.Error("Alice should show 1.5h mat hours")
	}

	// Verify Bob's name appears
	bobRow := page.Locator("tr:has-text('Bob Tester')")
	if visible, _ := bobRow.IsVisible(); !visible {
		t.Error("Bob Tester should be visible in attendance table")
	}

	// Verify Bob has injury flag
	bobInjury := bobRow.Locator("text=left knee")
	if visible, _ := bobInjury.IsVisible(); !visible {
		t.Error("Bob should show injury flag for 'left knee'")
	}
}

// TestCoachAttendance_EmptyState tests the empty state when no one has checked in.
func TestCoachAttendance_EmptyState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/attendance"); err != nil {
		t.Fatal(err)
	}

	// Verify empty state message
	emptyMsg := page.Locator("text=No check-ins recorded today")
	if err := emptyMsg.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}); err != nil {
		t.Error("expected 'No check-ins recorded today' message")
	}
}

// TestCoachAttendance_JSONEndpoint tests the JSON API returns attendees with belt/mat hours.
func TestCoachAttendance_JSONEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	cts, _ := app.Stores.ClassTypeStore.List(ctx)
	sched := scheduleDomain.Schedule{
		ID: uuid.New().String(), ClassTypeID: cts[0].ID,
		Day: scheduleDomain.Monday, StartTime: "06:00", EndTime: "07:00",
	}
	app.Stores.ScheduleStore.Save(ctx, sched)

	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "JSON Tester", Email: "json@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	app.Stores.MemberStore.Save(ctx, member)

	// Seed purple belt
	app.Stores.GradingRecordStore.Save(ctx, gradingDomain.Record{
		ID: uuid.New().String(), MemberID: member.ID,
		Belt: gradingDomain.BeltPurple, Stripe: 1,
		PromotedAt: time.Now().Add(-10 * 24 * time.Hour),
		ProposedBy: "coach-1", ApprovedBy: "admin-1", Method: gradingDomain.MethodStandard,
	})

	// Check in
	body := fmt.Sprintf(`{"MemberID":"%s","ScheduleID":"%s"}`, member.ID, sched.ID)
	resp, _ := http.Post(app.BaseURL+"/checkin", "application/json", strings.NewReader(body))
	resp.Body.Close()

	// Fetch JSON
	req, _ := http.NewRequest("GET", app.BaseURL+"/attendance", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Parse JSON response - verify it contains expected fields
	rawBody, _ := io.ReadAll(resp.Body)
	jsonStr := string(rawBody)

	if !strings.Contains(jsonStr, "JSON Tester") {
		t.Error("JSON response should contain member name")
	}
	if !strings.Contains(jsonStr, "purple") {
		t.Error("JSON response should contain belt color 'purple'")
	}
	if !strings.Contains(jsonStr, "MatHours") {
		t.Error("JSON response should contain MatHours field")
	}
}
