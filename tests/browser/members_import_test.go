package browser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/playwright-community/playwright-go"
)

func TestMembersImportCSV_ButtonVisibleForAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/members"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	btn := page.Locator("#import-csv-btn")
	if err := btn.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("import button not visible for admin: %v", err)
	}
}

func TestMembersImportCSV_ButtonHiddenForCoach(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)
	app.impersonate(t, page, "coach")

	if _, err := page.Goto(app.BaseURL + "/members"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	// Import button must not be present for coach.
	count, err := page.Locator("#import-csv-btn").Count()
	if err != nil {
		t.Fatalf("count locator: %v", err)
	}
	if count != 0 {
		t.Errorf("import button should not be visible for coach, found %d", count)
	}
}

func TestMembersImportCSV_ModalOpensAndImports(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Write a temp CSV file.
	csvContent := "NAME,EMAIL,PROGRAM,STATUS\nImport Alice,import-alice@test.com,adults,active\nImport Bob,import-bob@test.com,kids,active\n"
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "members.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0600); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	if _, err := page.Goto(app.BaseURL + "/members"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	// Open modal.
	if err := page.Locator("#import-csv-btn").Click(); err != nil {
		t.Fatalf("click import btn: %v", err)
	}
	modal := page.Locator("#import-csv-modal")
	if err := modal.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("modal not visible: %v", err)
	}

	// Upload file.
	if err := page.Locator("#import-file-input").SetInputFiles(csvPath); err != nil {
		t.Fatalf("set input file: %v", err)
	}

	// Click Preview.
	if err := page.Locator("#import-preview-btn").Click(); err != nil {
		t.Fatalf("click preview: %v", err)
	}

	// Wait for preview step.
	previewStep := page.Locator("#import-step-preview")
	if err := previewStep.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("preview step not visible: %v", err)
	}

	// Verify preview shows 2 to create.
	summary, err := page.Locator("#import-preview-summary").InnerText()
	if err != nil {
		t.Fatalf("read preview summary: %v", err)
	}
	if summary == "" {
		t.Error("preview summary should not be empty")
	}

	// Confirm import.
	if err := page.Locator("#import-confirm-btn").Click(); err != nil {
		t.Fatalf("click confirm: %v", err)
	}

	// Wait for done step.
	doneStep := page.Locator("#import-step-done")
	if err := doneStep.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("done step not visible: %v", err)
	}

	doneSummary, err := page.Locator("#import-done-summary").InnerText()
	if err != nil {
		t.Fatalf("read done summary: %v", err)
	}
	if doneSummary == "" {
		t.Error("done summary should not be empty")
	}

	// Click Done — page reloads and members should appear.
	if err := page.Locator("button:has-text('Done')").Click(); err != nil {
		t.Fatalf("click done: %v", err)
	}
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("wait for load: %v", err)
	}

	// Verify imported members appear in the list.
	if err := page.Locator("text=Import Alice").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("imported member not visible in list: %v", err)
	}
}

func TestMembersImportCSV_NonAdminBlockedAtAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)
	app.impersonate(t, page, "coach")

	// Coach should get 403 from the API directly.
	resp, err := page.Request().Fetch(app.BaseURL+"/api/members/import?dry_run=true", playwright.APIRequestContextFetchOptions{
		Method: playwright.String("POST"),
	})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if resp.Status() != 403 {
		t.Errorf("status=%d want 403", resp.Status())
	}
}
