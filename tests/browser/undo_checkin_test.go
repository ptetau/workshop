package browser_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	attendanceDomain "workshop/internal/domain/attendance"
	memberDomain "workshop/internal/domain/member"
)

// TestUndoCheckIn_HappyPath tests US-2.5.1: Undo accidental check-in.
// Member checks in, sees the check-in listed, taps Un-Check-In, and it's removed.
func TestUndoCheckIn_HappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	// Seed a member
	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "Undo Tester", Email: "undo@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, member); err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	// Create a check-in for today via API
	checkinBody := fmt.Sprintf(`{"MemberID":"%s"}`, member.ID)
	resp, err := http.Post(app.BaseURL+"/checkin", "application/json", strings.NewReader(checkinBody))
	if err != nil {
		t.Fatalf("failed to check in: %v", err)
	}
	resp.Body.Close()

	// Verify attendance exists via API
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
	attendanceID := checkins[0].ID

	// Undo the check-in via API
	undoBody := fmt.Sprintf(`{"AttendanceID":"%s"}`, attendanceID)
	req, _ := http.NewRequest("DELETE", app.BaseURL+"/api/attendance/undo", strings.NewReader(undoBody))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to undo: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify attendance is gone
	resp, err = http.Get(app.BaseURL + "/api/attendance/member?member_id=" + member.ID)
	if err != nil {
		t.Fatalf("failed to get attendance: %v", err)
	}
	json.NewDecoder(resp.Body).Decode(&checkins)
	resp.Body.Close()

	if len(checkins) != 0 {
		t.Errorf("expected 0 check-ins after undo, got %d", len(checkins))
	}
}

// TestUndoCheckIn_TodayOnly tests US-2.5.2: Un-check-in limited to today.
// Yesterday's check-ins cannot be undone.
func TestUndoCheckIn_TodayOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	// Seed a member
	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "Yesterday Tester", Email: "yesterday@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, member); err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	// Create a check-in for yesterday directly in the store
	yesterday := time.Now().AddDate(0, 0, -1)
	oldCheckin := attendanceDomain.Attendance{
		ID:          uuid.New().String(),
		MemberID:    member.ID,
		CheckInTime: yesterday,
	}
	if err := app.Stores.AttendanceStore.Save(ctx, oldCheckin); err != nil {
		t.Fatalf("failed to save yesterday's check-in: %v", err)
	}

	// Try to undo yesterday's check-in — should fail
	undoBody := fmt.Sprintf(`{"AttendanceID":"%s"}`, oldCheckin.ID)
	req, _ := http.NewRequest("DELETE", app.BaseURL+"/api/attendance/undo", strings.NewReader(undoBody))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to call undo: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for yesterday's check-in, got %d", resp.StatusCode)
	}

	// Verify the check-in still exists
	records, err := app.Stores.AttendanceStore.ListByMemberID(ctx, member.ID)
	if err != nil {
		t.Fatalf("failed to list attendance: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected yesterday's check-in to still exist, got %d records", len(records))
	}

	// Also verify that today's check-ins list doesn't include yesterday
	resp2, err := http.Get(app.BaseURL + "/api/attendance/member?member_id=" + member.ID)
	if err != nil {
		t.Fatalf("failed to get today's attendance: %v", err)
	}
	var todayCheckins []attendanceDomain.Attendance
	json.NewDecoder(resp2.Body).Decode(&todayCheckins)
	resp2.Body.Close()

	if len(todayCheckins) != 0 {
		t.Errorf("expected 0 today's check-ins (yesterday's should not appear), got %d", len(todayCheckins))
	}
}

// TestUndoCheckIn_KioskUI tests the kiosk UI flow for un-check-in.
func TestUndoCheckIn_KioskUI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	// Seed a member
	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "Kiosk Undo", Email: "kioskundo@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, member); err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	// Create a check-in for today
	checkin := attendanceDomain.Attendance{
		ID:          uuid.New().String(),
		MemberID:    member.ID,
		CheckInTime: time.Now(),
	}
	if err := app.Stores.AttendanceStore.Save(ctx, checkin); err != nil {
		t.Fatalf("failed to save check-in: %v", err)
	}

	// Launch kiosk
	_, err := page.Goto(app.BaseURL + "/kiosk")
	if err != nil {
		t.Fatalf("failed to navigate to kiosk: %v", err)
	}

	// Search for the member
	nameInput := page.Locator("#nameInput")
	if err := nameInput.Fill("Kiosk Undo"); err != nil {
		t.Fatalf("failed to fill search: %v", err)
	}
	// Wait for search results
	page.WaitForTimeout(500)

	// Click the member in results
	memberResult := page.Locator("#memberResults li:has-text('Kiosk Undo')")
	if err := memberResult.Click(); err != nil {
		t.Fatalf("failed to click member: %v", err)
	}
	// Wait for check-in list to load
	page.WaitForTimeout(500)

	// Verify "Already checked in today" section is visible
	todayDiv := page.Locator("#todayCheckins")
	visible, err := todayDiv.IsVisible()
	if err != nil {
		t.Fatalf("failed to check todayCheckins visibility: %v", err)
	}
	if !visible {
		t.Error("todayCheckins section should be visible when member has today's check-ins")
	}

	// Verify Un-Check-In button exists
	undoBtn := page.Locator(".undo-btn")
	count, err := undoBtn.Count()
	if err != nil {
		t.Fatalf("failed to count undo buttons: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 undo button, got %d", count)
	}

	// Click Un-Check-In
	if err := undoBtn.Click(); err != nil {
		t.Fatalf("failed to click undo: %v", err)
	}
	// Wait for the undo to complete and view to refresh
	page.WaitForTimeout(1000)

	// After undo, the todayCheckins section should be hidden (no more check-ins)
	visible, _ = todayDiv.IsVisible()
	if visible {
		t.Error("todayCheckins section should be hidden after undoing the only check-in")
	}

	// Verify attendance is actually deleted
	records, err := app.Stores.AttendanceStore.ListByMemberID(ctx, member.ID)
	if err != nil {
		t.Fatalf("failed to list attendance: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 attendance records after undo, got %d", len(records))
	}
}
