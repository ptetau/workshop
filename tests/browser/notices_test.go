package browser_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"

	noticeDomain "workshop/internal/domain/notice"
)

// TestNotice_CoachDraftsNotice covers US-8.1.1: Coach drafts a notice.
// AC: Given Open Mat is cancelled this Friday
//
//	When I create a notice with type school_wide
//	Then the notice is saved as a draft
//	And Admin receives it for review and publishing
func TestNotice_CoachDraftsNotice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Navigate to admin notices page
	_, err := page.Goto(app.BaseURL + "/admin/notices")
	if err != nil {
		t.Fatalf("failed to navigate to notices: %v", err)
	}

	// Fill in the notice form (template uses id= selectors, not name=)
	if err := page.Locator("#noticeTitle").Fill("Open Mat cancelled this Friday"); err != nil {
		t.Fatalf("failed to fill title: %v", err)
	}
	if err := page.Locator("#noticeContent").Fill("Open Mat is cancelled this Friday due to a grading event."); err != nil {
		t.Fatalf("failed to fill content: %v", err)
	}

	// Type defaults to school_wide, which is what we want

	// Click "Save as Draft" button (JS-driven, not form submit)
	if err := page.Locator("button:has-text('Save as Draft')").Click(); err != nil {
		t.Fatalf("failed to click Save as Draft: %v", err)
	}

	// Wait for the notice to appear in the list (loaded via fetch)
	err = page.Locator("#noticeList >> text=Open Mat cancelled this Friday").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("notice did not appear after creation: %v", err)
	}

	// Verify it's in draft status (shown as a status badge in the list)
	err = page.Locator("#noticeList >> text=draft").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("draft status badge not visible in notice list")
	}
}

// TestNotice_AdminPublishesNotice covers US-8.1.2: Admin publishes a notice.
// AC: Given a coach drafted "Open Mat cancelled this Friday"
//
//	When I review and click "Publish"
//	Then the notice appears on all member dashboards and the kiosk
func TestNotice_AdminPublishesNotice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// First create a draft notice via store
	createNoticeViaStore(t, app, "Grading this Saturday", "Reminder: grading event this Saturday at 10am.")

	// Navigate to admin notices page
	_, err := page.Goto(app.BaseURL + "/admin/notices")
	if err != nil {
		t.Fatalf("failed to navigate to notices: %v", err)
	}

	// Wait for the notice to appear in the JS-loaded list
	err = page.Locator("#noticeList >> text=Grading this Saturday").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("draft notice not found on admin page: %v", err)
	}

	// Verify it's currently in draft status
	err = page.Locator("#noticeList >> text=draft").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Fatalf("notice should be in draft status before publish: %v", err)
	}

	// Click the Publish button (the template renders a small green button per draft notice)
	if err := page.Locator("#noticeList button:has-text('Publish')").First().Click(); err != nil {
		t.Fatalf("failed to click publish: %v", err)
	}

	// After publish, the list reloads via JS. Wait for "published" status badge to appear
	err = page.Locator("#noticeList >> text=published").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("published status badge not visible after clicking Publish")
	}

	// The Publish button should be gone (only shown for drafts)
	publishBtns := page.Locator("#noticeList button:has-text('Publish')")
	count, _ := publishBtns.Count()
	if count > 0 {
		t.Error("Publish button still visible after publishing — notice may still be in draft")
	}
}

// TestNotice_AdminNoticesPageLoads verifies the admin notices management page loads correctly.
func TestNotice_AdminNoticesPageLoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	_, err := page.Goto(app.BaseURL + "/admin/notices")
	if err != nil {
		t.Fatalf("failed to navigate to admin notices: %v", err)
	}

	// Page should contain a notices heading or form
	title := page.Locator("h1, h2, h3")
	if count, _ := title.Count(); count == 0 {
		t.Error("admin notices page has no headings")
	}
}

// createNoticeViaStore creates a draft notice directly via the notice store.
func createNoticeViaStore(t *testing.T, app *testApp, title, content string) {
	t.Helper()
	n := noticeDomain.Notice{
		ID:        uuid.New().String(),
		Type:      noticeDomain.TypeSchoolWide,
		Status:    noticeDomain.StatusDraft,
		Title:     title,
		Content:   content,
		CreatedBy: app.AdminID,
		Color:     noticeDomain.ColorOrange,
		CreatedAt: time.Now(),
	}
	if err := app.Stores.NoticeStore.Save(context.Background(), n); err != nil {
		t.Fatalf("failed to create test notice: %v", err)
	}
}
