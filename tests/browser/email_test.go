package browser_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"

	memberDomain "workshop/internal/domain/member"
)

// seedTestMember creates a member directly in the store for use in email tests.
func seedTestMember(t *testing.T, app *testApp, name, email string) string {
	t.Helper()
	id := uuid.New().String()
	m := memberDomain.Member{
		ID:        id,
		AccountID: app.AdminID,
		Email:     email,
		Name:      name,
		Program:   memberDomain.ProgramAdults,
		Status:    memberDomain.StatusActive,
	}
	if err := app.Stores.MemberStore.Save(context.Background(), m); err != nil {
		t.Fatalf("failed to seed test member: %v", err)
	}
	return id
}

// TestEmail_AdminEmailsPageLoads verifies the admin emails management page loads.
func TestEmail_AdminEmailsPageLoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	_, err := page.Goto(app.BaseURL + "/admin/emails")
	if err != nil {
		t.Fatalf("failed to navigate to admin emails: %v", err)
	}

	// Page should have the Email Management heading
	err = page.Locator("h1:has-text('Email Management')").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("Email Management heading not found")
	}

	// Should have a Compose Email link
	err = page.Locator("a:has-text('Compose Email')").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("Compose Email link not found")
	}
}

// TestEmail_ComposePageLoads verifies the compose email page loads.
func TestEmail_ComposePageLoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	_, err := page.Goto(app.BaseURL + "/admin/emails/compose")
	if err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Page should have compose heading
	err = page.Locator("h1:has-text('Compose Email')").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("Compose Email heading not found")
	}

	// Should have subject and body fields
	if count, _ := page.Locator("#emailSubject").Count(); count == 0 {
		t.Error("subject input not found")
	}
	if count, _ := page.Locator("#emailBody").Count(); count == 0 {
		t.Error("body textarea not found")
	}
}

// TestEmail_RecipientSearch tests searching for members to add as recipients.
// Covers #107: Admin can search for members by name and add them as recipients.
func TestEmail_RecipientSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Seed test members
	seedTestMember(t, app, "Marcus Almeida", "marcus@test.com")
	seedTestMember(t, app, "Yuki Nakai", "yuki@test.com")

	page := app.newPage(t)
	app.login(t, page)

	_, err := page.Goto(app.BaseURL + "/admin/emails/compose")
	if err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Type in the search box
	if err := page.Locator("#recipientSearch").Fill("Marcus"); err != nil {
		t.Fatalf("failed to fill recipient search: %v", err)
	}

	// Wait for search results to appear
	err = page.Locator("#searchResults >> text=Marcus Almeida").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("search results did not show Marcus Almeida: %v", err)
	}

	// Click on the result to add as recipient
	if err := page.Locator("#searchResults >> text=Marcus Almeida").Click(); err != nil {
		t.Fatalf("failed to click search result: %v", err)
	}

	// Verify the recipient appears in selected recipients
	err = page.Locator("#selectedRecipients >> text=Marcus Almeida").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("Marcus Almeida not shown in selected recipients after clicking")
	}

	// Verify count updates
	err = page.Locator("#recipientCount >> text=1 selected").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("recipient count did not update to '1 selected'")
	}
}

// TestEmail_SaveDraft tests composing and saving an email draft.
// Covers #113: Admin can compose and save a draft email.
func TestEmail_SaveDraft(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Seed a test member
	seedTestMember(t, app, "Roger Gracie", "roger@test.com")

	page := app.newPage(t)
	app.login(t, page)

	_, err := page.Goto(app.BaseURL + "/admin/emails/compose")
	if err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Fill subject and body
	if err := page.Locator("#emailSubject").Fill("Grading Day Reminder"); err != nil {
		t.Fatalf("failed to fill subject: %v", err)
	}
	if err := page.Locator("#emailBody").Fill("Don't forget grading is this Saturday at 10am."); err != nil {
		t.Fatalf("failed to fill body: %v", err)
	}

	// Search and add a recipient
	if err := page.Locator("#recipientSearch").Fill("Roger"); err != nil {
		t.Fatalf("failed to fill recipient search: %v", err)
	}
	err = page.Locator("#searchResults >> text=Roger Gracie").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("Roger Gracie not found in search: %v", err)
	}
	if err := page.Locator("#searchResults >> text=Roger Gracie").Click(); err != nil {
		t.Fatalf("failed to select Roger: %v", err)
	}

	// Click Save Draft
	if err := page.Locator("button:has-text('Save Draft')").Click(); err != nil {
		t.Fatalf("failed to click Save Draft: %v", err)
	}

	// Wait for success message
	err = page.Locator("#formMsg >> text=Draft saved").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("draft saved confirmation message not shown")
	}

	// Navigate to emails list and verify draft appears
	_, err = page.Goto(app.BaseURL + "/admin/emails")
	if err != nil {
		t.Fatalf("failed to navigate to emails list: %v", err)
	}

	err = page.Locator("#emailList >> text=Grading Day Reminder").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("saved draft 'Grading Day Reminder' not found in email list")
	}

	// Verify draft status badge
	err = page.Locator("#emailList >> text=draft").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("draft status badge not visible")
	}
}

// TestEmail_SendEmail tests composing and sending an email.
// Covers #113: Admin can compose and send an email (uses noop sender in tests).
func TestEmail_SendEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Seed a test member
	seedTestMember(t, app, "Andre Galvao", "andre@test.com")

	page := app.newPage(t)
	app.login(t, page)

	_, err := page.Goto(app.BaseURL + "/admin/emails/compose")
	if err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Fill subject and body
	if err := page.Locator("#emailSubject").Fill("Schedule Change"); err != nil {
		t.Fatalf("failed to fill subject: %v", err)
	}
	if err := page.Locator("#emailBody").Fill("Monday class moves to Tuesday."); err != nil {
		t.Fatalf("failed to fill body: %v", err)
	}

	// Search and add recipient
	if err := page.Locator("#recipientSearch").Fill("Andre"); err != nil {
		t.Fatalf("failed to fill search: %v", err)
	}
	err = page.Locator("#searchResults >> text=Andre Galvao").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("Andre Galvao not found in search: %v", err)
	}
	if err := page.Locator("#searchResults >> text=Andre Galvao").Click(); err != nil {
		t.Fatalf("failed to select Andre: %v", err)
	}

	// Click Send Now
	if err := page.Locator("button:has-text('Send Now')").Click(); err != nil {
		t.Fatalf("failed to click Send Now: %v", err)
	}

	// Should redirect to emails list after send
	err = page.WaitForURL(app.BaseURL+"/admin/emails", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		// May still be on compose page with success message — check for "sent" in list
		_, _ = page.Goto(app.BaseURL + "/admin/emails")
	}

	// Wait for email list to load and show the sent email
	time.Sleep(500 * time.Millisecond) // Allow JS to render
	err = page.Locator("#emailList >> text=Schedule Change").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("sent email 'Schedule Change' not found in email list")
	}

	// Verify it shows as sent
	err = page.Locator("#emailList >> text=sent").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("sent status badge not visible for sent email")
	}
}

// TestEmail_DashboardEmailLink verifies the Emails link appears on the admin dashboard.
func TestEmail_DashboardEmailLink(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Should already be on dashboard after login
	emailLink := page.Locator("a:has-text('Emails')")
	err := emailLink.First().WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("Emails link not found on admin dashboard")
	}
}
