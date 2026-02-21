package browser_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"

	attendanceDomain "workshop/internal/domain/attendance"
	memberDomain "workshop/internal/domain/member"
)

// TestTrainingLog_AttendanceVolumeGraph verifies the training volume graph renders and responds to range/compare UI.
func TestTrainingLog_AttendanceVolumeGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()
	memberID := uuid.New().String()
	// IMPORTANT: training-log page resolves MemberID by session email.
	m := memberDomain.Member{
		ID:        memberID,
		Name:      "Training Volume Member",
		Email:     "admin@test.com",
		Program:   "Adults",
		Status:    "active",
		Fee:       100,
		Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, m); err != nil {
		t.Fatalf("seed member: %v", err)
	}

	now := time.Now()
	att1 := attendanceDomain.Attendance{
		ID:          uuid.New().String(),
		MemberID:    memberID,
		ScheduleID:  "sched-1",
		ClassDate:   now.AddDate(0, 0, -1).Format("2006-01-02"),
		CheckInTime: now.AddDate(0, 0, -1),
	}
	att2 := attendanceDomain.Attendance{
		ID:          uuid.New().String(),
		MemberID:    memberID,
		ScheduleID:  "sched-1",
		ClassDate:   now.AddDate(0, 0, -2).Format("2006-01-02"),
		CheckInTime: now.AddDate(0, 0, -2),
	}
	if err := app.Stores.AttendanceStore.Save(ctx, att1); err != nil {
		t.Fatalf("seed attendance 1: %v", err)
	}
	if err := app.Stores.AttendanceStore.Save(ctx, att2); err != nil {
		t.Fatalf("seed attendance 2: %v", err)
	}

	// Switch to member role (uses same email, so we can still resolve the member record).
	app.impersonate(t, page, "member")

	if _, err := page.Goto(app.BaseURL + "/training-log"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	// Graph should render.
	svg := page.Locator("#volumeGraphInner svg")
	if err := svg.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("volume graph svg not visible: %v", err)
	}

	// Default legend includes "You" + "All members avg".
	if err := page.Locator("#volumeLegend:has-text('You')").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("missing You legend: %v", err)
	}
	if err := page.Locator("#volumeLegend:has-text('All members avg')").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("missing avg legend: %v", err)
	}

	// Switch to year range disables compare.
	rangeSel := page.Locator("#volumeRange")
	if _, err := rangeSel.SelectOption(playwright.SelectOptionValues{Values: &[]string{"year"}}); err != nil {
		t.Fatalf("select year range: %v", err)
	}
	cmp := page.Locator("#volumeCompare")
	if err := cmp.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateAttached}); err != nil {
		t.Fatalf("compare checkbox missing: %v", err)
	}
	disabled, err := cmp.IsDisabled()
	if err != nil {
		t.Fatalf("compare IsDisabled: %v", err)
	}
	if !disabled {
		t.Fatalf("compare should be disabled for year range")
	}

	// Switch back to month enables compare and allows extra series.
	if _, err := rangeSel.SelectOption(playwright.SelectOptionValues{Values: &[]string{"month"}}); err != nil {
		t.Fatalf("select month range: %v", err)
	}
	disabled, err = cmp.IsDisabled()
	if err != nil {
		t.Fatalf("compare IsDisabled after month: %v", err)
	}
	if disabled {
		t.Fatalf("compare should be enabled for month range")
	}

	if err := cmp.Click(); err != nil {
		t.Fatalf("enable compare: %v", err)
	}

	if err := page.Locator("#volumeLegend:has-text('Last month')").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("missing Last month legend: %v", err)
	}
	if err := page.Locator("#volumeLegend:has-text('Same month last year')").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("missing Same month last year legend: %v", err)
	}

	// Should have >= 4 polylines (you + avg + comparisons).
	polys := page.Locator("#volumeGraphInner polyline")
	count, err := polys.Count()
	if err != nil {
		t.Fatalf("polyline count: %v", err)
	}
	if count < 4 {
		t.Fatalf("polyline count=%d, want >= 4", count)
	}
}
