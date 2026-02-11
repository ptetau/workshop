package browser_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"

	attendanceDomain "workshop/internal/domain/attendance"
	estimatedHoursDomain "workshop/internal/domain/estimatedhours"
	memberDomain "workshop/internal/domain/member"
)

// TestEstimatedHours_OverlapAdd tests US-3.4.3: choosing "Add on Top" keeps attendance and saves estimate.
func TestEstimatedHours_OverlapAdd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	// Seed a member
	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "Overlap Add Tester", Email: "overlapadd@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, member); err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	// Seed attendance records in the overlap range
	att1 := attendanceDomain.Attendance{
		ID: uuid.New().String(), MemberID: member.ID,
		CheckInTime: time.Date(2026, 2, 5, 18, 0, 0, 0, time.UTC),
		MatHours:    1.5,
	}
	att2 := attendanceDomain.Attendance{
		ID: uuid.New().String(), MemberID: member.ID,
		CheckInTime: time.Date(2026, 2, 12, 18, 0, 0, 0, time.UTC),
		MatHours:    2.0,
	}
	if err := app.Stores.AttendanceStore.Save(ctx, att1); err != nil {
		t.Fatalf("failed to seed attendance 1: %v", err)
	}
	if err := app.Stores.AttendanceStore.Save(ctx, att2); err != nil {
		t.Fatalf("failed to seed attendance 2: %v", err)
	}

	// Navigate to the member profile
	if _, err := page.Goto(app.BaseURL + "/members/profile?id=" + member.ID); err != nil {
		t.Fatalf("failed to navigate to member profile: %v", err)
	}

	// Open the "Add Estimated Hours" details
	if err := page.Locator("summary").Filter(playwright.LocatorFilterOptions{HasText: "Add Estimated Hours"}).Click(); err != nil {
		t.Fatalf("failed to open estimated hours form: %v", err)
	}

	// Fill the form
	if err := page.Locator("#estStart").Fill("2026-02-01"); err != nil {
		t.Fatalf("failed to fill start date: %v", err)
	}
	if err := page.Locator("#estEnd").Fill("2026-04-30"); err != nil {
		t.Fatalf("failed to fill end date: %v", err)
	}
	if err := page.Locator("#estWeekly").Fill("3"); err != nil {
		t.Fatalf("failed to fill weekly hours: %v", err)
	}
	if err := page.Locator("#estNote").Fill("extra training at partner gym"); err != nil {
		t.Fatalf("failed to fill note: %v", err)
	}

	// Submit the form — this should trigger overlap check
	if err := page.Locator("#estHoursForm button[type=submit]").Click(); err != nil {
		t.Fatalf("failed to submit form: %v", err)
	}

	// Wait for overlap warning to appear
	overlapWarning := page.Locator("#overlapWarning")
	if err := overlapWarning.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("overlap warning did not appear: %v", err)
	}

	// Verify the "Add on Top" button exists
	addBtn := page.Locator("button:has-text('Add on Top')")
	if visible, _ := addBtn.IsVisible(); !visible {
		t.Fatal("expected 'Add on Top' button to be visible in overlap warning")
	}

	// Click "Add on Top"
	if err := addBtn.Click(); err != nil {
		t.Fatalf("failed to click Add on Top: %v", err)
	}

	// Wait for success message
	estMsg := page.Locator("#estMsg")
	if err := estMsg.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("success message did not appear: %v", err)
	}
	msgText, _ := estMsg.TextContent()
	if !strings.Contains(msgText, "Added") {
		t.Errorf("expected success message containing 'Added', got %q", msgText)
	}

	// Verify: attendance records are still present (not deleted)
	records, err := app.Stores.AttendanceStore.ListByMemberID(ctx, member.ID)
	if err != nil {
		t.Fatalf("failed to list attendance: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 attendance records preserved after 'Add on Top', got %d", len(records))
	}

	// Verify: estimated hours entry was saved
	estimates, err := app.Stores.EstimatedHoursStore.ListByMemberID(ctx, member.ID)
	if err != nil {
		t.Fatalf("failed to list estimated hours: %v", err)
	}
	if len(estimates) != 1 {
		t.Fatalf("expected 1 estimated hours entry, got %d", len(estimates))
	}
	if estimates[0].Note != "extra training at partner gym" {
		t.Errorf("note = %q, want %q", estimates[0].Note, "extra training at partner gym")
	}
	if estimates[0].TotalHours == 0 {
		t.Error("expected non-zero TotalHours")
	}
}

// TestEstimatedHours_OverlapAddAPI tests the "add" overlap mode via direct API call.
func TestEstimatedHours_OverlapAddAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	// Seed a member
	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "API Add Tester", Email: "apiadd@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, member); err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	// Seed attendance
	att := attendanceDomain.Attendance{
		ID: uuid.New().String(), MemberID: member.ID,
		CheckInTime: time.Date(2026, 3, 1, 18, 0, 0, 0, time.UTC),
		MatHours:    1.5,
	}
	if err := app.Stores.AttendanceStore.Save(ctx, att); err != nil {
		t.Fatalf("failed to seed attendance: %v", err)
	}

	// POST estimated hours with OverlapMode=add
	body := fmt.Sprintf(`{"MemberID":"%s","StartDate":"2026-02-01","EndDate":"2026-04-30","WeeklyHours":3,"Note":"API add test","OverlapMode":"add"}`, member.ID)
	resp, err := page.Request().Post(app.BaseURL+"/api/estimated-hours", playwright.APIRequestContextPostOptions{
		Data:    body,
		Headers: map[string]string{"Content-Type": "application/json"},
	})
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	if resp.Status() != 201 {
		respBody, _ := resp.Text()
		t.Fatalf("expected 201, got %d: %s", resp.Status(), respBody)
	}

	var result estimatedHoursDomain.EstimatedHours
	respBody, _ := resp.Body()
	json.Unmarshal(respBody, &result)
	if result.TotalHours == 0 {
		t.Error("expected non-zero TotalHours")
	}

	// Verify attendance NOT deleted
	records, _ := app.Stores.AttendanceStore.ListByMemberID(ctx, member.ID)
	if len(records) != 1 {
		t.Errorf("expected 1 attendance record preserved, got %d", len(records))
	}

	// Verify estimate saved
	estimates, _ := app.Stores.EstimatedHoursStore.ListByMemberID(ctx, member.ID)
	if len(estimates) != 1 {
		t.Errorf("expected 1 estimate, got %d", len(estimates))
	}
}

// TestEstimatedHours_TrainingLogShowsBulkHours tests that bulk estimated hours appear in the training log.
func TestEstimatedHours_TrainingLogShowsBulkHours(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	// Seed a member
	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "Bulk Hours Viewer", Email: "bulkhours@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, member); err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	// Seed an attendance record (needed so training log projection doesn't return early)
	att := attendanceDomain.Attendance{
		ID: uuid.New().String(), MemberID: member.ID,
		CheckInTime: time.Date(2026, 2, 5, 18, 0, 0, 0, time.UTC),
		MatHours:    1.5,
	}
	if err := app.Stores.AttendanceStore.Save(ctx, att); err != nil {
		t.Fatalf("failed to seed attendance: %v", err)
	}

	// Seed a bulk estimated hours entry
	est := estimatedHoursDomain.EstimatedHours{
		ID: uuid.New().String(), MemberID: member.ID,
		StartDate: "2026-01-01", EndDate: "2026-03-31",
		WeeklyHours: 3, TotalHours: 39,
		Source: "estimate", Status: "approved",
		Note: "partner gym", CreatedBy: app.AdminID,
		CreatedAt: time.Now(),
	}
	if err := app.Stores.EstimatedHoursStore.Save(ctx, est); err != nil {
		t.Fatalf("failed to seed estimated hours: %v", err)
	}

	// GET training log via authenticated page request
	resp, err := page.Request().Get(app.BaseURL + "/api/training-log?member_id=" + member.ID)
	if err != nil {
		t.Fatalf("failed to get training log: %v", err)
	}
	if resp.Status() != 200 {
		respText, _ := resp.Text()
		t.Fatalf("training log API returned %d: %s", resp.Status(), respText)
	}

	var logResult struct {
		TotalMatHours      float64 `json:"TotalMatHours"`
		BulkEstimatedHours float64 `json:"BulkEstimatedHours"`
	}
	respBody, _ := resp.Body()
	json.Unmarshal(respBody, &logResult)

	if logResult.BulkEstimatedHours != 39 {
		t.Errorf("BulkEstimatedHours = %v, want 39", logResult.BulkEstimatedHours)
	}
	if logResult.TotalMatHours < 39 {
		t.Errorf("TotalMatHours = %v, want >= 39", logResult.TotalMatHours)
	}
}
