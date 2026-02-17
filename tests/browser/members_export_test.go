package browser_test

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

func TestMembersExportCSV_LinkVisible_AndEndpointReturnsCSV(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	seedMembers(t, app, 10)

	// Use some params so we can verify the export link preserves them.
	_, err := page.Goto(app.BaseURL + "/members?q=Member+001&program=Adults&sort=name&dir=asc")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	exportLink := page.Locator("#export-members-csv")
	if err := exportLink.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("export link not visible: %v", err)
	}

	href, err := exportLink.GetAttribute("href")
	if err != nil {
		t.Fatalf("failed to read export href: %v", err)
	}
	if !strings.HasPrefix(href, "/api/members/export") {
		t.Fatalf("unexpected export href=%q", href)
	}
	if !strings.Contains(href, "program=Adults") || !strings.Contains(href, "sort=name") || !strings.Contains(href, "dir=asc") {
		t.Fatalf("export href does not include expected params: %q", href)
	}

	// Use the authenticated API context to fetch the CSV (avoids flaky browser download handling).
	resp, err := page.Request().Get(app.BaseURL + href)
	if err != nil {
		t.Fatalf("GET export failed: %v", err)
	}
	if resp.Status() != 200 {
		body, _ := resp.Text()
		t.Fatalf("export status=%d body=%s", resp.Status(), body)
	}
	body, _ := resp.Text()
	if !strings.Contains(body, "ID,AccountID,Name,Email,Program,Status,Fee,Frequency,GradingMetric") {
		t.Fatalf("export body missing header row: %q", body)
	}

	// Coach should also see the link.
	app.impersonate(t, page, "coach")
	if _, err := page.Goto(app.BaseURL + "/members"); err != nil {
		t.Fatalf("failed to navigate as coach: %v", err)
	}
	visible, _ := exportLink.IsVisible()
	if !visible {
		t.Fatalf("export link should be visible for coach")
	}
}
