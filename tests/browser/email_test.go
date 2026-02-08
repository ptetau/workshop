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
	return seedTestMemberWithProgram(t, app, name, email, memberDomain.ProgramAdults)
}

// seedTestMemberWithProgram creates a member with a specific program.
func seedTestMemberWithProgram(t *testing.T, app *testApp, name, email, program string) string {
	t.Helper()
	id := uuid.New().String()
	m := memberDomain.Member{
		ID:        id,
		AccountID: app.AdminID,
		Email:     email,
		Name:      name,
		Program:   program,
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

// TestEmail_FilterByProgram tests filtering recipients by program.
// Covers #110: Admin can filter recipients by program (e.g. Kids, Adults).
func TestEmail_FilterByProgram(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Seed members in different programs
	seedTestMember(t, app, "Alice Adults", "alice@test.com")
	seedTestMemberWithProgram(t, app, "Bobby Kids", "bobby@test.com", memberDomain.ProgramKids)
	seedTestMemberWithProgram(t, app, "Charlie Kids", "charlie@test.com", memberDomain.ProgramKids)

	page := app.newPage(t)
	app.login(t, page)

	_, err := page.Goto(app.BaseURL + "/admin/emails/compose")
	if err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Select "Kids" from program filter
	if _, err := page.Locator("#programFilter").SelectOption(playwright.SelectOptionValues{Values: &[]string{"kids"}}); err != nil {
		t.Fatalf("failed to select Kids program: %v", err)
	}

	// Wait for filter results to appear
	err = page.Locator("#filterResultsList >> text=Bobby Kids").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("Bobby Kids not found in filter results: %v", err)
	}

	// Charlie Kids should also appear
	err = page.Locator("#filterResultsList >> text=Charlie Kids").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("Charlie Kids not found in filter results")
	}

	// Alice Adults should NOT appear in Kids filter
	count, _ := page.Locator("#filterResultsList >> text=Alice Adults").Count()
	if count > 0 {
		t.Error("Alice Adults should not appear in Kids program filter")
	}
}

// TestEmail_SelectAll tests the Select All button for filtered recipients.
// Covers #111: Admin can select all filtered members then deselect individuals.
func TestEmail_SelectAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Seed 3 kids members
	seedTestMemberWithProgram(t, app, "Dana Kids", "dana@test.com", memberDomain.ProgramKids)
	seedTestMemberWithProgram(t, app, "Eve Kids", "eve@test.com", memberDomain.ProgramKids)
	seedTestMemberWithProgram(t, app, "Finn Kids", "finn@test.com", memberDomain.ProgramKids)

	page := app.newPage(t)
	app.login(t, page)

	_, err := page.Goto(app.BaseURL + "/admin/emails/compose")
	if err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Filter by Kids
	if _, err := page.Locator("#programFilter").SelectOption(playwright.SelectOptionValues{Values: &[]string{"kids"}}); err != nil {
		t.Fatalf("failed to select Kids program: %v", err)
	}

	// Wait for filter results
	err = page.Locator("#filterResultsList >> text=Dana Kids").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("filter results did not load: %v", err)
	}

	// Click Select All
	if err := page.Locator("#selectAllBtn").Click(); err != nil {
		t.Fatalf("failed to click Select All: %v", err)
	}

	// Verify all 3 are selected
	err = page.Locator("#recipientCount >> text=3 selected").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("recipient count did not update to '3 selected' after Select All")
	}

	// Deselect one by unchecking
	cb := page.Locator("#filterResultsList input[data-member-name='Dana Kids']")
	if err := cb.Uncheck(); err != nil {
		t.Fatalf("failed to uncheck Dana Kids: %v", err)
	}

	// Verify count dropped to 2
	err = page.Locator("#recipientCount >> text=2 selected").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("recipient count did not update to '2 selected' after deselecting one")
	}
}

// TestEmail_InvertSelection tests the Invert Selection button.
// Covers #112: Admin can invert the selection to quickly switch included/excluded.
func TestEmail_InvertSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Seed 3 adults members
	seedTestMember(t, app, "Grace Adult", "grace@test.com")
	seedTestMember(t, app, "Hank Adult", "hank@test.com")
	seedTestMember(t, app, "Iris Adult", "iris@test.com")

	page := app.newPage(t)
	app.login(t, page)

	_, err := page.Goto(app.BaseURL + "/admin/emails/compose")
	if err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Filter by Adults
	if _, err := page.Locator("#programFilter").SelectOption(playwright.SelectOptionValues{Values: &[]string{"adults"}}); err != nil {
		t.Fatalf("failed to select Adults program: %v", err)
	}

	// Wait for filter results
	err = page.Locator("#filterResultsList >> text=Grace Adult").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("filter results did not load: %v", err)
	}

	// Select All first
	if err := page.Locator("#selectAllBtn").Click(); err != nil {
		t.Fatalf("failed to click Select All: %v", err)
	}

	// Verify 3 selected
	err = page.Locator("#recipientCount >> text=3 selected").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Fatalf("expected 3 selected after Select All: %v", err)
	}

	// Deselect Grace by unchecking
	cb := page.Locator("#filterResultsList input[data-member-name='Grace Adult']")
	if err := cb.Uncheck(); err != nil {
		t.Fatalf("failed to uncheck Grace Adult: %v", err)
	}

	// Now 2 selected (Hank + Iris)
	err = page.Locator("#recipientCount >> text=2 selected").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Fatalf("expected 2 selected: %v", err)
	}

	// Click Invert — should select Grace, deselect Hank + Iris → 1 selected
	if err := page.Locator("#invertBtn").Click(); err != nil {
		t.Fatalf("failed to click Invert Selection: %v", err)
	}

	err = page.Locator("#recipientCount >> text=1 selected").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("after invert, expected 1 selected (Grace only)")
	}

	// Grace should be in selected recipients
	err = page.Locator("#selectedRecipients >> text=Grace Adult").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Error("Grace Adult should be in selected recipients after invert")
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
