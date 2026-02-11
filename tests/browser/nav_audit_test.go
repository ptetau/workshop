package browser_test

import (
	"context"
	"fmt"
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

// TestNavAudit_AdminNavLinks verifies admin nav has core links visible and grouped links in More dropdown.
func TestNavAudit_AdminNavLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/dashboard"); err != nil {
		t.Fatal(err)
	}

	nav := page.Locator(".nav-links")

	// Core nav links should be directly visible
	for _, link := range []struct{ text, href string }{
		{"Members", "/members"},
		{"Attendance", "/attendance"},
		{"Grading", "/admin/grading"},
		{"Emails", "/admin/emails"},
	} {
		loc := nav.Locator(fmt.Sprintf("a[href='%s']", link.href))
		if visible, _ := loc.IsVisible(); !visible {
			t.Errorf("admin core nav missing link: %s (%s)", link.text, link.href)
		}
	}

	// More dropdown should exist
	moreToggle := nav.Locator(".nav-more summary")
	if visible, _ := moreToggle.IsVisible(); !visible {
		t.Fatal("admin nav missing More dropdown")
	}

	// Open the More dropdown
	if err := moreToggle.Click(); err != nil {
		t.Fatal("failed to open More dropdown:", err)
	}

	// Links inside More should now be visible
	for _, link := range []struct{ text, href string }{
		{"Schedules", "/admin/schedules"},
		{"Milestones", "/admin/milestones"},
		{"Curriculum", "/curriculum"},
		{"Themes", "/themes"},
		{"Library", "/library"},
		{"Notices", "/admin/notices"},
		{"Accounts", "/admin/accounts"},
		{"Terms", "/admin/terms"},
		{"Holidays", "/admin/holidays"},
		{"Inactive Radar", "/admin/inactive"},
	} {
		loc := nav.Locator(fmt.Sprintf(".nav-more-menu a[href='%s']", link.href))
		if visible, _ := loc.IsVisible(); !visible {
			t.Errorf("admin More menu missing link: %s (%s)", link.text, link.href)
		}
	}

	// Verify grouped sections exist
	for _, label := range []string{"Training", "Content", "Settings"} {
		loc := nav.Locator(fmt.Sprintf(".nav-more-label:has-text('%s')", label))
		if visible, _ := loc.IsVisible(); !visible {
			t.Errorf("admin More menu missing group label: %s", label)
		}
	}
}

// TestNavAudit_CoachNavLinks verifies coach nav has core links visible and overflow in More dropdown.
func TestNavAudit_CoachNavLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Impersonate coach
	app.impersonate(t, page, "coach")

	nav := page.Locator(".nav-links")

	// Core nav links should be directly visible
	for _, link := range []struct{ text, href string }{
		{"Members", "/members"},
		{"Attendance", "/attendance"},
		{"Grading", "/admin/grading"},
		{"Curriculum", "/curriculum"},
	} {
		loc := nav.Locator(fmt.Sprintf("a[href='%s']", link.href))
		if visible, _ := loc.IsVisible(); !visible {
			t.Errorf("coach core nav missing link: %s (%s)", link.text, link.href)
		}
	}

	// Open More dropdown
	moreToggle := nav.Locator(".nav-more summary")
	if visible, _ := moreToggle.IsVisible(); !visible {
		t.Fatal("coach nav missing More dropdown")
	}
	if err := moreToggle.Click(); err != nil {
		t.Fatal("failed to open coach More dropdown:", err)
	}

	// Links inside More should now be visible
	for _, link := range []struct{ text, href string }{
		{"Schedules", "/admin/schedules"},
		{"Notices", "/admin/notices"},
		{"Themes", "/themes"},
		{"Library", "/library"},
		{"Kiosk", "/kiosk"},
	} {
		loc := nav.Locator(fmt.Sprintf(".nav-more-menu a[href='%s']", link.href))
		if visible, _ := loc.IsVisible(); !visible {
			t.Errorf("coach More menu missing link: %s (%s)", link.text, link.href)
		}
	}

	// Coach should NOT have admin-only links
	for _, href := range []string{"/admin/emails", "/admin/accounts", "/admin/terms"} {
		loc := nav.Locator(fmt.Sprintf("a[href='%s']", href))
		count, _ := loc.Count()
		if count > 0 {
			t.Errorf("coach nav should NOT contain %s", href)
		}
	}
}

// TestNavAudit_MemberNavLinks verifies member nav contains expected links and no More dropdown.
func TestNavAudit_MemberNavLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Impersonate member
	app.impersonate(t, page, "member")

	nav := page.Locator(".nav-links")
	for _, link := range []struct{ text, href string }{
		{"Training Log", "/training-log"},
		{"Messages", "/messages"},
		{"Curriculum", "/curriculum"},
		{"Library", "/library"},
	} {
		loc := nav.Locator(fmt.Sprintf("a[href='%s']", link.href))
		if visible, _ := loc.IsVisible(); !visible {
			t.Errorf("member nav missing link: %s (%s)", link.text, link.href)
		}
	}

	// Member should NOT have More dropdown
	moreToggle := nav.Locator(".nav-more summary")
	count, _ := moreToggle.Count()
	if count > 0 {
		t.Error("member nav should NOT have a More dropdown")
	}
}

// TestNavAudit_TrialNavLinks verifies trial nav does NOT contain Curriculum, Themes, or Library.
func TestNavAudit_TrialNavLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Impersonate trial
	app.impersonate(t, page, "trial")

	nav := page.Locator(".nav-links")

	// Trial SHOULD have these links
	for _, link := range []struct{ text, href string }{
		{"Training Log", "/training-log"},
		{"Messages", "/messages"},
	} {
		loc := nav.Locator(fmt.Sprintf("a[href='%s']", link.href))
		if visible, _ := loc.IsVisible(); !visible {
			t.Errorf("trial nav missing link: %s (%s)", link.text, link.href)
		}
	}

	// Trial should NOT have these links
	for _, href := range []string{"/curriculum", "/themes", "/library"} {
		loc := nav.Locator(fmt.Sprintf("a[href='%s']", href))
		count, _ := loc.Count()
		if count > 0 {
			if visible, _ := loc.IsVisible(); visible {
				t.Errorf("trial nav should NOT contain %s", href)
			}
		}
	}
}

// TestNavAudit_TrialCurriculumRedirects verifies trial users are redirected away from /curriculum.
func TestNavAudit_TrialCurriculumRedirects(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	app.impersonate(t, page, "trial")

	resp, err := page.Goto(app.BaseURL + "/curriculum")
	if err != nil {
		t.Fatal(err)
	}
	// Should redirect to dashboard (303 -> 200 after redirect)
	url := page.URL()
	if !strings.Contains(url, "/dashboard") {
		t.Errorf("trial accessing /curriculum should redirect to /dashboard, got %s (status %d)", url, resp.Status())
	}
}

// TestNavAudit_CurriculumLinkNavigates verifies the Curriculum nav link works via the More dropdown (admin).
func TestNavAudit_CurriculumLinkNavigates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/dashboard"); err != nil {
		t.Fatal(err)
	}

	// For admin, Curriculum is inside the More dropdown — open it first
	if err := page.Locator(".nav-more summary").Click(); err != nil {
		t.Fatal("failed to open More dropdown:", err)
	}

	if err := page.Locator(".nav-more-menu a[href='/curriculum']").Click(); err != nil {
		t.Fatal("failed to click Curriculum link:", err)
	}

	if err := page.WaitForURL("**/curriculum"); err != nil {
		t.Error("expected to navigate to /curriculum")
	}
}

// TestNavAudit_CoachDashboardBeltAndMatHours verifies coach dashboard shows belt icons and mat hours.
func TestNavAudit_CoachDashboardBeltAndMatHours(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()

	// Seed data
	cts, _ := app.Stores.ClassTypeStore.List(ctx)
	sched := scheduleDomain.Schedule{
		ID: uuid.New().String(), ClassTypeID: cts[0].ID,
		Day: scheduleDomain.Monday, StartTime: "18:00", EndTime: "19:30",
	}
	app.Stores.ScheduleStore.Save(ctx, sched)

	member := memberDomain.Member{
		ID: uuid.New().String(), Name: "Belt Tester", Email: "belt@test.com",
		Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly",
	}
	app.Stores.MemberStore.Save(ctx, member)

	// Give blue belt
	app.Stores.GradingRecordStore.Save(ctx, gradingDomain.Record{
		ID: uuid.New().String(), MemberID: member.ID,
		Belt: gradingDomain.BeltBlue, Stripe: 3,
		PromotedAt: time.Now().Add(-10 * 24 * time.Hour),
		ProposedBy: "coach-1", ApprovedBy: "admin-1", Method: gradingDomain.MethodStandard,
	})

	// Report injury
	app.Stores.InjuryStore.Save(ctx, injuryDomain.Injury{
		ID: uuid.New().String(), MemberID: member.ID,
		BodyPart: "right shoulder", Description: "strain",
		ReportedAt: time.Now(),
	})

	// Check in via API
	body := fmt.Sprintf(`{"MemberID":"%s","ScheduleID":"%s"}`, member.ID, sched.ID)
	resp, _ := http.Post(app.BaseURL+"/checkin", "application/json", strings.NewReader(body))
	resp.Body.Close()

	// Impersonate coach and view dashboard
	app.impersonate(t, page, "coach")

	// Verify attendance table has Belt and Mat Hrs columns
	beltHeader := page.Locator("th:has-text('Belt')")
	if visible, _ := beltHeader.IsVisible(); !visible {
		t.Error("coach dashboard attendance should have Belt column header")
	}

	matHeader := page.Locator("th:has-text('Mat Hrs')")
	if visible, _ := matHeader.IsVisible(); !visible {
		t.Error("coach dashboard attendance should have Mat Hrs column header")
	}

	// Verify member row shows mat hours
	memberRow := page.Locator("tr:has-text('Belt Tester')")
	if visible, _ := memberRow.IsVisible(); !visible {
		t.Error("Belt Tester should appear in coach dashboard attendance")
	}

	matHoursCell := memberRow.Locator("td:has-text('1.5h')")
	if visible, _ := matHoursCell.IsVisible(); !visible {
		t.Error("should show 1.5h mat hours for 90-minute class")
	}

	// Verify injury flag
	injuryFlag := memberRow.Locator("text=right shoulder")
	if visible, _ := injuryFlag.IsVisible(); !visible {
		t.Error("should show injury flag for 'right shoulder'")
	}
}

// TestNavAudit_DashboardQuickLinks verifies grouped dashboard links resolve without error.
func TestNavAudit_DashboardQuickLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/dashboard"); err != nil {
		t.Fatal(err)
	}

	// Admin dashboard now has Training, Content, Settings sections instead of Quick Links
	for _, section := range []string{"Training", "Content", "Settings"} {
		heading := page.Locator(fmt.Sprintf("h2:has-text('%s')", section))
		if visible, _ := heading.IsVisible(); !visible {
			t.Errorf("admin dashboard missing section: %s", section)
		}
	}

	// Collect all link hrefs from grouped sections
	links := page.Locator("main .card h2 ~ div a")
	count, _ := links.Count()
	if count == 0 {
		t.Fatal("no dashboard links found")
	}

	var hrefs []string
	for i := 0; i < count; i++ {
		href, _ := links.Nth(i).GetAttribute("href")
		if href != "" {
			hrefs = append(hrefs, href)
		}
	}

	// Navigate to each link using the browser (shares auth session)
	for _, href := range hrefs {
		resp, err := page.Goto(app.BaseURL + href)
		if err != nil {
			t.Errorf("dashboard link %s failed to navigate: %v", href, err)
			continue
		}
		status := resp.Status()
		if status >= 400 {
			t.Errorf("dashboard link %s returned %d", href, status)
		}
	}
}

// TestNavAudit_MobileHamburger verifies the hamburger menu toggle works on mobile viewport.
func TestNavAudit_MobileHamburger(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)

	// Set mobile viewport
	page.SetViewportSize(375, 667)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/dashboard"); err != nil {
		t.Fatal(err)
	}

	// Hamburger should be visible
	hamburger := page.Locator(".nav-toggle")
	if err := hamburger.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}); err != nil {
		t.Fatal("hamburger toggle should be visible on mobile viewport")
	}

	// Nav links should be hidden initially
	navLinks := page.Locator(".nav-links")
	visible, _ := navLinks.IsVisible()
	if visible {
		t.Error("nav links should be hidden on mobile before toggling")
	}

	// Click hamburger to open
	if err := hamburger.Click(); err != nil {
		t.Fatal("failed to click hamburger:", err)
	}

	// Nav links should now be visible
	if err := navLinks.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}); err != nil {
		t.Error("nav links should be visible after hamburger click")
	}

	// Click hamburger again to close
	if err := hamburger.Click(); err != nil {
		t.Fatal("failed to click hamburger second time:", err)
	}

	// Wait briefly for toggle
	page.WaitForTimeout(300)

	// Nav links should be hidden again
	visible, _ = navLinks.IsVisible()
	if visible {
		t.Error("nav links should be hidden after second hamburger click")
	}
}

// TestNavAudit_MobileNoOverflow verifies no horizontal overflow on mobile.
func TestNavAudit_MobileNoOverflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)

	page.SetViewportSize(375, 667)
	app.login(t, page)

	for _, path := range []string{"/dashboard", "/members", "/attendance"} {
		if _, err := page.Goto(app.BaseURL + path); err != nil {
			t.Fatal(err)
		}
		// Check body scrollWidth <= viewport width
		overflow, _ := page.Evaluate("() => document.body.scrollWidth > window.innerWidth")
		if overflow == true {
			t.Errorf("horizontal overflow detected on %s at 375px viewport", path)
		}
	}
}
