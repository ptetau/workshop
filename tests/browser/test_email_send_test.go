package browser_test

import (
	"testing"

	"github.com/playwright-community/playwright-go"
)

// TestTestEmailSend_ComposeAndTestSend verifies the admin can compose a draft and send a test email.
func TestTestEmailSend_ComposeAndTestSend(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Navigate to compose page
	if _, err := page.Goto(app.BaseURL + "/admin/emails/compose"); err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Fill in subject and body
	if err := page.Locator("#emailSubject").Fill("Test Subject"); err != nil {
		t.Fatalf("failed to fill subject: %v", err)
	}
	if err := page.Locator("#emailBody").Fill("<p>Test body content</p>"); err != nil {
		t.Fatalf("failed to fill body: %v", err)
	}

	// Verify test send UI elements exist
	testInput := page.Locator("#testSendAddress")
	if visible, _ := testInput.IsVisible(); !visible {
		t.Fatal("test send address input not visible")
	}
	testBtn := page.Locator("#testSendBtn")
	if visible, _ := testBtn.IsVisible(); !visible {
		t.Fatal("test send button not visible")
	}

	// Fill test address and click send test
	if err := testInput.Fill("admin@test.com"); err != nil {
		t.Fatalf("failed to fill test address: %v", err)
	}
	if err := testBtn.Click(); err != nil {
		t.Fatalf("failed to click send test: %v", err)
	}

	// Wait for success message (the noop sender is configured in tests)
	msg := page.Locator("#formMsg")
	if err := msg.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("success message did not appear: %v", err)
	}
	text, _ := msg.TextContent()
	if text == "" {
		t.Error("expected a feedback message after test send")
	}

	// Verify draft was saved (emailID should now be set) but page still shows compose
	emailID, _ := page.Locator("#emailID").InputValue()
	if emailID == "" {
		t.Error("expected emailID to be set after test send (draft auto-saved)")
	}
}

// TestTestEmailSend_EmptyAddressDisabled verifies that test send with empty address shows error.
func TestTestEmailSend_EmptyAddressDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Navigate to compose page
	if _, err := page.Goto(app.BaseURL + "/admin/emails/compose"); err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Fill subject and body but leave test address empty
	if err := page.Locator("#emailSubject").Fill("Test Subject"); err != nil {
		t.Fatalf("failed to fill subject: %v", err)
	}
	if err := page.Locator("#emailBody").Fill("<p>Body</p>"); err != nil {
		t.Fatalf("failed to fill body: %v", err)
	}

	// Click send test with empty address
	if err := page.Locator("#testSendBtn").Click(); err != nil {
		t.Fatalf("failed to click send test: %v", err)
	}

	// Should show error message
	msg := page.Locator("#formMsg")
	if err := msg.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}); err != nil {
		t.Fatalf("error message did not appear: %v", err)
	}
	text, _ := msg.TextContent()
	if text != "Enter a test email address" {
		t.Errorf("expected 'Enter a test email address', got %q", text)
	}

	// Verify no draft was created (emailID should be empty)
	emailID, _ := page.Locator("#emailID").InputValue()
	if emailID != "" {
		t.Error("expected emailID to remain empty when test send fails validation")
	}
}
