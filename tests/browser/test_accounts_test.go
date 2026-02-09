package browser_test

import (
	"context"
	"testing"

	"github.com/playwright-community/playwright-go"

	"workshop/internal/application/orchestrators"
)

// TestTestAccounts_CoachLogin verifies the test coach account can log in and see the coach dashboard.
func TestTestAccounts_CoachLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)

	// Seed test accounts for this test
	testAcctDeps := orchestrators.TestAccountSeedDeps{
		AccountStore: app.Stores.AccountStore,
		MemberStore:  app.Stores.MemberStore,
	}
	if err := orchestrators.ExecuteSeedTestAccounts(context.Background(), testAcctDeps); err != nil {
		t.Fatalf("failed to seed test accounts: %v", err)
	}

	page := app.newPage(t)

	// Log in as test coach
	if _, err := page.Goto(app.BaseURL + "/login"); err != nil {
		t.Fatalf("failed to navigate to login: %v", err)
	}
	if err := page.Locator("input[name=Email]").Fill("info+coach@workshopjiujitsu.co.nz"); err != nil {
		t.Fatalf("failed to fill email: %v", err)
	}
	if err := page.Locator("input[name=Password]").Fill("Umami+coach!"); err != nil {
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

	// Verify coach dashboard content
	heading := page.Locator("h1")
	text, _ := heading.TextContent()
	if text != "Coach Dashboard" {
		t.Errorf("expected 'Coach Dashboard' heading, got %q", text)
	}
}

// TestTestAccounts_MemberLogin verifies the test member account can log in and see the member dashboard.
func TestTestAccounts_MemberLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)

	// Seed test accounts for this test
	testAcctDeps := orchestrators.TestAccountSeedDeps{
		AccountStore: app.Stores.AccountStore,
		MemberStore:  app.Stores.MemberStore,
	}
	if err := orchestrators.ExecuteSeedTestAccounts(context.Background(), testAcctDeps); err != nil {
		t.Fatalf("failed to seed test accounts: %v", err)
	}

	page := app.newPage(t)

	// Log in as test member
	if _, err := page.Goto(app.BaseURL + "/login"); err != nil {
		t.Fatalf("failed to navigate to login: %v", err)
	}
	if err := page.Locator("input[name=Email]").Fill("info+member@workshopjiujitsu.co.nz"); err != nil {
		t.Fatalf("failed to fill email: %v", err)
	}
	if err := page.Locator("input[name=Password]").Fill("Umami+coach!"); err != nil {
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

	// Verify member dashboard content
	heading := page.Locator("h1")
	text, _ := heading.TextContent()
	if text == "" {
		t.Error("expected a heading on the member dashboard")
	}
}
