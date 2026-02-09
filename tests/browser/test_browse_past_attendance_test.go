package browser_test

import (
	"context"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"

	"workshop/internal/application/orchestrators"
	"workshop/internal/domain/attendance"

	"github.com/google/uuid"
)

// seedAttendanceForDate creates an attendance record for a specific member on a specific date.
func seedAttendanceForDate(t *testing.T, app *testApp, memberID, scheduleID string, date time.Time) {
	t.Helper()
	ctx := context.Background()
	a := attendance.Attendance{
		ID:          uuid.New().String(),
		MemberID:    memberID,
		ScheduleID:  scheduleID,
		ClassDate:   date.Format("2006-01-02"),
		CheckInTime: date.Add(6 * time.Hour),
	}
	if err := app.Stores.AttendanceStore.Save(ctx, a); err != nil {
		t.Fatalf("failed to seed attendance: %v", err)
	}
}

// TestBrowsePastAttendance_DateNavVisible verifies the date navigation is visible on the attendance page.
func TestBrowsePastAttendance_DateNavVisible(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/attendance"); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Date picker should be visible
	datePicker := page.Locator("#datePicker")
	if visible, _ := datePicker.IsVisible(); !visible {
		t.Error("date picker not visible")
	}

	// Prev/next arrows should be visible
	prevArrow := page.Locator("a[title='Previous day']")
	if visible, _ := prevArrow.IsVisible(); !visible {
		t.Error("previous day arrow not visible")
	}
	nextArrow := page.Locator("a[title='Next day']")
	if visible, _ := nextArrow.IsVisible(); !visible {
		t.Error("next day arrow not visible")
	}

	// "Back to Today" should NOT be visible when viewing today
	backToToday := page.Locator("a:has-text('Back to Today')")
	count, _ := backToToday.Count()
	if count > 0 {
		if visible, _ := backToToday.IsVisible(); visible {
			t.Error("Back to Today should not be visible when viewing today")
		}
	}
}

// TestBrowsePastAttendance_NavigateToPastDate verifies navigating to a past date shows attendance and read-only banner.
func TestBrowsePastAttendance_NavigateToPastDate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)

	// Seed synthetic data so there are members and attendance
	ctx := context.Background()
	deps := orchestrators.SyntheticSeedDeps{
		AccountStore:         app.Stores.AccountStore,
		MemberStore:          app.Stores.MemberStore,
		WaiverStore:          app.Stores.WaiverStore,
		InjuryStore:          app.Stores.InjuryStore,
		AttendanceStore:      app.Stores.AttendanceStore,
		ScheduleStore:        app.Stores.ScheduleStore,
		TermStore:            app.Stores.TermStore,
		HolidayStore:         app.Stores.HolidayStore,
		NoticeStore:          app.Stores.NoticeStore,
		GradingRecordStore:   app.Stores.GradingRecordStore,
		GradingConfigStore:   app.Stores.GradingConfigStore,
		GradingProposalStore: app.Stores.GradingProposalStore,
		MessageStore:         app.Stores.MessageStore,
		ObservationStore:     app.Stores.ObservationStore,
		MilestoneStore:       app.Stores.MilestoneStore,
		TrainingGoalStore:    app.Stores.TrainingGoalStore,
		ClassTypeStore:       app.Stores.ClassTypeStore,
		ThemeStore:           app.Stores.ThemeStore,
		ClipStore:            app.Stores.ClipStore,
	}
	if err := orchestrators.ExecuteSeedSynthetic(ctx, deps, app.AdminID); err != nil {
		t.Fatalf("failed to seed synthetic: %v", err)
	}

	page := app.newPage(t)
	app.login(t, page)

	// Navigate to a date 7 days ago (synthetic data has attendance in last 60 days)
	pastDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	if _, err := page.Goto(app.BaseURL + "/attendance?date=" + pastDate); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Read-only banner should be visible
	banner := page.Locator(".read-only-banner")
	if err := banner.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatal("read-only banner not visible on past date")
	}

	// "Back to Today" should be visible
	backToToday := page.Locator("a:has-text('Back to Today')")
	if visible, _ := backToToday.IsVisible(); !visible {
		t.Error("Back to Today button not visible on past date")
	}

	// Date picker should show the past date
	val, _ := page.Locator("#datePicker").InputValue()
	if val != pastDate {
		t.Errorf("date picker shows %q, want %q", val, pastDate)
	}

	// Title should say "Attendance" not "Today's Attendance"
	title, _ := page.Locator("h1").TextContent()
	if title == "📊 Today's Attendance" {
		t.Error("title should not say 'Today's Attendance' on a past date")
	}
}

// TestBrowsePastAttendance_PrevNextArrows verifies clicking prev arrow navigates to the previous day.
func TestBrowsePastAttendance_PrevNextArrows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Start on today
	if _, err := page.Goto(app.BaseURL + "/attendance"); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Click prev arrow
	prevArrow := page.Locator("a[title='Previous day']")
	if err := prevArrow.Click(); err != nil {
		t.Fatalf("failed to click prev arrow: %v", err)
	}

	// Wait for navigation
	page.WaitForTimeout(1000)

	// Should now be on yesterday's date
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	val, _ := page.Locator("#datePicker").InputValue()
	if val != yesterday {
		t.Errorf("after clicking prev, date picker shows %q, want %q", val, yesterday)
	}

	// Read-only banner should be visible
	banner := page.Locator(".read-only-banner")
	if visible, _ := banner.IsVisible(); !visible {
		t.Error("read-only banner should be visible on yesterday")
	}
}

// TestBrowsePastAttendance_BackToTodayButton verifies the Back to Today button returns to today.
func TestBrowsePastAttendance_BackToTodayButton(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Navigate to a past date
	pastDate := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
	if _, err := page.Goto(app.BaseURL + "/attendance?date=" + pastDate); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Click Back to Today
	backToToday := page.Locator("a:has-text('Back to Today')")
	if err := backToToday.Click(); err != nil {
		t.Fatalf("failed to click Back to Today: %v", err)
	}

	// Wait for navigation
	page.WaitForTimeout(1000)

	// Should now be on today
	today := time.Now().Format("2006-01-02")
	val, _ := page.Locator("#datePicker").InputValue()
	if val != today {
		t.Errorf("after Back to Today, date picker shows %q, want %q", val, today)
	}

	// Read-only banner should NOT be visible
	banner := page.Locator(".read-only-banner")
	count, _ := banner.Count()
	if count > 0 {
		if visible, _ := banner.IsVisible(); visible {
			t.Error("read-only banner should not be visible on today")
		}
	}

	// Title should say "Today's Attendance"
	title, _ := page.Locator("h1").TextContent()
	if title != "📊 Today's Attendance" {
		t.Errorf("title should be 'Today's Attendance', got %q", title)
	}
}

// TestBrowsePastAttendance_PastDateReadOnly verifies Check In link is hidden on past dates.
func TestBrowsePastAttendance_PastDateReadOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Navigate to a date far in the past (no attendance expected)
	pastDate := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
	if _, err := page.Goto(app.BaseURL + "/attendance?date=" + pastDate); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// The "Check In Member" button should NOT be visible on a past date with no attendance
	checkinBtn := page.Locator("a:has-text('Check In Member')")
	count, _ := checkinBtn.Count()
	if count > 0 {
		if visible, _ := checkinBtn.IsVisible(); visible {
			t.Error("Check In Member link should not be visible on past date")
		}
	}
}
