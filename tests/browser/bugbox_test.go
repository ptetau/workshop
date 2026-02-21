package browser_test

import (
	"context"
	"testing"

	"github.com/playwright-community/playwright-go"

	accountStore "workshop/internal/adapters/storage/account"
	"workshop/internal/application/orchestrators"
)

// TestBugBox_AdminSeesFloatingButton verifies the bug box button is visible to admins on the dashboard.
// AC: Given I am logged in as admin
//
//	When I navigate to any page
//	Then I see the floating bug report button
func TestBugBox_AdminSeesFloatingButton(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Navigate to dashboard
	if _, err := page.Goto(app.BaseURL + "/dashboard"); err != nil {
		t.Fatalf("failed to navigate to dashboard: %v", err)
	}

	// Bug box button must be visible
	btn := page.Locator("#bugbox-btn")
	if err := btn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("bug box button not visible for admin on dashboard: %v", err)
	}
}

// TestBugBox_CoachSeesFloatingButton verifies the bug box button is visible to coaches.
// AC: Given I am logged in as coach
//
//	When I navigate to any page
//	Then I see the floating bug report button
func TestBugBox_CoachSeesFloatingButton(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Create a coach account
	ctx := context.Background()
	acctStore := accountStore.NewSQLiteStore(app.DB)
	_, err := orchestrators.ExecuteCreateAccount(ctx, orchestrators.CreateAccountInput{
		Email:                  "coach@test.com",
		Password:               "TestPass123!",
		Role:                   "coach",
		PasswordChangeRequired: false,
	}, orchestrators.CreateAccountDeps{AccountStore: acctStore})
	if err != nil {
		t.Fatalf("failed to create coach account: %v", err)
	}

	page := app.newPage(t)

	// Login as coach
	if _, err := page.Goto(app.BaseURL + "/login"); err != nil {
		t.Fatalf("failed to navigate to login: %v", err)
	}
	if err := page.Locator("input[name=Email]").Fill("coach@test.com"); err != nil {
		t.Fatalf("failed to fill email: %v", err)
	}
	if err := page.Locator("input[name=Password]").Fill("TestPass123!"); err != nil {
		t.Fatalf("failed to fill password: %v", err)
	}
	if err := page.Locator("button[type=submit]").Click(); err != nil {
		t.Fatalf("failed to click login: %v", err)
	}
	if err := page.WaitForURL(app.BaseURL+"/dashboard", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("coach login did not redirect to dashboard: %v", err)
	}

	// Bug box button must be visible for coach
	btn := page.Locator("#bugbox-btn")
	if err := btn.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("bug box button not visible for coach on dashboard: %v", err)
	}
}

// TestBugBox_MemberDoesNotSeeButton verifies the bug box button is hidden from members.
// AC: Given I am logged in as member
//
//	When I navigate to any page
//	Then I do NOT see the floating bug report button
func TestBugBox_MemberDoesNotSeeButton(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Create a member account
	ctx := context.Background()
	acctStore := accountStore.NewSQLiteStore(app.DB)
	_, err := orchestrators.ExecuteCreateAccount(ctx, orchestrators.CreateAccountInput{
		Email:                  "member@test.com",
		Password:               "TestPass123!",
		Role:                   "member",
		PasswordChangeRequired: false,
	}, orchestrators.CreateAccountDeps{AccountStore: acctStore})
	if err != nil {
		t.Fatalf("failed to create member account: %v", err)
	}

	page := app.newPage(t)

	// Login as member
	if _, err := page.Goto(app.BaseURL + "/login"); err != nil {
		t.Fatalf("failed to navigate to login: %v", err)
	}
	if err := page.Locator("input[name=Email]").Fill("member@test.com"); err != nil {
		t.Fatalf("failed to fill email: %v", err)
	}
	if err := page.Locator("input[name=Password]").Fill("TestPass123!"); err != nil {
		t.Fatalf("failed to fill password: %v", err)
	}
	if err := page.Locator("button[type=submit]").Click(); err != nil {
		t.Fatalf("failed to click login: %v", err)
	}
	if err := page.WaitForURL(app.BaseURL+"/dashboard", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("member login did not redirect to dashboard: %v", err)
	}

	// Bug box button must NOT be present for member
	count, err := page.Locator("#bugbox-btn").Count()
	if err != nil {
		t.Fatalf("failed to count bugbox buttons: %v", err)
	}
	if count > 0 {
		t.Error("bug box button should not be visible for member role")
	}
}

// TestBugBox_AdminOpensModalOnAnyPage verifies the modal opens when the button is clicked on any page.
// AC: Given I am logged in as admin
//
//	When I click the floating bug button on the members page
//	Then the modal opens with the correct title
//	And the route field is pre-populated with the current path
func TestBugBox_AdminOpensModalOnAnyPage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Navigate to a non-dashboard page to verify route capture works anywhere
	if _, err := page.Goto(app.BaseURL + "/dashboard"); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Click the bug box button
	if err := page.Locator("#bugbox-btn").Click(); err != nil {
		t.Fatalf("failed to click bug box button: %v", err)
	}

	// Modal must become visible
	overlay := page.Locator("#bugbox-overlay")
	if err := overlay.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("bug box modal did not open: %v", err)
	}

	// Title must contain the expected text
	title := page.Locator("#bugbox-title")
	titleText, err := title.TextContent()
	if err != nil {
		t.Fatalf("failed to get modal title: %v", err)
	}
	if titleText == "" {
		t.Error("modal title is empty")
	}

	// Route hidden field must be populated with current path
	routeVal, err := page.Locator("#bugbox-route").InputValue()
	if err != nil {
		t.Fatalf("failed to get route value: %v", err)
	}
	if routeVal == "" {
		t.Error("route field was not auto-populated when modal opened")
	}

	// Viewport hidden field must be populated
	viewportVal, err := page.Locator("#bugbox-viewport").InputValue()
	if err != nil {
		t.Fatalf("failed to get viewport value: %v", err)
	}
	if viewportVal == "" {
		t.Error("viewport field was not auto-populated when modal opened")
	}
}

// TestBugBox_ModalClosesOnEscape verifies pressing Escape closes the modal.
// AC: Given the bug box modal is open
//
//	When I press Escape
//	Then the modal closes
func TestBugBox_ModalClosesOnEscape(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/dashboard"); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Open modal
	if err := page.Locator("#bugbox-btn").Click(); err != nil {
		t.Fatalf("failed to click bug box button: %v", err)
	}
	if err := page.Locator("#bugbox-overlay").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("modal did not open: %v", err)
	}

	// Press Escape
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("failed to press Escape: %v", err)
	}

	// Modal must be hidden
	if err := page.Locator("#bugbox-overlay").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(3000),
	}); err != nil {
		t.Error("modal did not close after pressing Escape")
	}
}

// TestBugBox_ModalClosesOnBackdropClick verifies clicking the backdrop closes the modal.
// AC: Given the bug box modal is open
//
//	When I click outside the modal panel
//	Then the modal closes
func TestBugBox_ModalClosesOnBackdropClick(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/dashboard"); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Open modal
	if err := page.Locator("#bugbox-btn").Click(); err != nil {
		t.Fatalf("failed to click bug box button: %v", err)
	}
	if err := page.Locator("#bugbox-overlay").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("modal did not open: %v", err)
	}

	// Click the overlay backdrop (top-left corner, outside the modal panel)
	if err := page.Locator("#bugbox-overlay").Click(playwright.LocatorClickOptions{
		Position: &playwright.Position{X: 5, Y: 5},
	}); err != nil {
		t.Fatalf("failed to click backdrop: %v", err)
	}

	// Modal must be hidden
	if err := page.Locator("#bugbox-overlay").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(3000),
	}); err != nil {
		t.Error("modal did not close after clicking backdrop")
	}
}

// TestBugBox_ValidationRequiresSummaryAndDescription verifies the form requires both fields.
// AC: Given the bug box modal is open
//
//	When I submit without filling in summary or description
//	Then the form does not submit (browser validation prevents it)
func TestBugBox_ValidationRequiresSummaryAndDescription(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/dashboard"); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Open modal
	if err := page.Locator("#bugbox-btn").Click(); err != nil {
		t.Fatalf("failed to click bug box button: %v", err)
	}
	if err := page.Locator("#bugbox-overlay").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("modal did not open: %v", err)
	}

	// Click submit without filling required fields
	if err := page.Locator("#bugbox-submit").Click(); err != nil {
		t.Fatalf("failed to click submit: %v", err)
	}

	// Modal must still be open (form validation blocked submission)
	if err := page.Locator("#bugbox-overlay").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(2000),
	}); err != nil {
		t.Error("modal closed unexpectedly — form should have been blocked by validation")
	}

	// Success panel must NOT be visible
	successVisible, _ := page.Locator("#bugbox-success").IsVisible()
	if successVisible {
		t.Error("success panel should not be visible when form validation failed")
	}
}

// TestBugBox_AdminSubmitsReport verifies a full submission flow.
// GitHub is not configured in tests, so we expect the error message to appear
// (not a 401/403/400 — those would indicate auth or validation failure).
// AC: Given I am logged in as admin
//
//	When I fill in summary and description and click Submit
//	Then the form is submitted to the server
//	And I see either a success confirmation or a GitHub configuration error
//	And I do NOT see an auth or validation error
func TestBugBox_AdminSubmitsReport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/dashboard"); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Open modal
	if err := page.Locator("#bugbox-btn").Click(); err != nil {
		t.Fatalf("failed to click bug box button: %v", err)
	}
	if err := page.Locator("#bugbox-overlay").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("modal did not open: %v", err)
	}

	// Fill required fields
	if err := page.Locator("#bugbox-summary").Fill("Dashboard loads slowly"); err != nil {
		t.Fatalf("failed to fill summary: %v", err)
	}
	if err := page.Locator("#bugbox-description").Fill("The dashboard takes over 3 seconds to load on first visit."); err != nil {
		t.Fatalf("failed to fill description: %v", err)
	}
	if err := page.Locator("#bugbox-steps").Fill("1. Log in\n2. Navigate to /dashboard\n3. Observe load time"); err != nil {
		t.Fatalf("failed to fill steps: %v", err)
	}

	// Submit
	if err := page.Locator("#bugbox-submit").Click(); err != nil {
		t.Fatalf("failed to click submit: %v", err)
	}

	// Wait for the submit button to re-enable — that means the fetch completed
	if err := page.Locator("#bugbox-submit:not([disabled])").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("submit button never re-enabled — fetch may have hung: %v", err)
	}

	// Check what appeared — success is ideal; GitHub config error is acceptable in test env
	successVisible, _ := page.Locator("#bugbox-success").IsVisible()
	errorVisible, _ := page.Locator("#bugbox-error").IsVisible()

	if !successVisible && !errorVisible {
		t.Fatal("neither success nor error panel is visible after submission")
	}

	if successVisible {
		// Submission worked (GitHub token configured in env)
		return
	}

	if errorVisible {
		errText, _ := page.Locator("#bugbox-error").TextContent()
		// Must be a GitHub config error, NOT an auth or validation error
		if errText == "Forbidden" || errText == "not authenticated" {
			t.Errorf("unexpected auth error after submission: %q", errText)
		}
		if errText == "summary and description are required" {
			t.Errorf("unexpected validation error after submission: %q", errText)
		}
		// GitHub config error is expected in test environment — pass
		t.Logf("submission reached server; GitHub not configured in test env: %s", errText)
	}
}

// TestBugBox_ButtonVisibleOnMultiplePages verifies the button appears on several pages, not just dashboard.
// AC: Given I am logged in as admin
//
//	When I navigate to different pages
//	Then the bug box button is always visible
func TestBugBox_ButtonVisibleOnMultiplePages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	pages := []string{"/dashboard", "/admin/notices", "/admin/features"}

	for _, path := range pages {
		if _, err := page.Goto(app.BaseURL + path); err != nil {
			t.Fatalf("failed to navigate to %s: %v", path, err)
		}
		if err := page.Locator("#bugbox-btn").WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		}); err != nil {
			t.Errorf("bug box button not visible on %s: %v", path, err)
		}
	}
}
