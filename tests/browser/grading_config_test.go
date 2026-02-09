package browser_test

import (
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

// TestGradingConfig_AddViaDropdown tests #191: admin adds grading config via belt dropdown.
func TestGradingConfig_AddViaDropdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Navigate to grading admin page
	if _, err := page.Goto(app.BaseURL+"/admin/grading", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	page.WaitForTimeout(500)

	// Select program "adults", belt "purple" via dropdown, fill hours and stripes
	page.SelectOption("#cfgProgram", playwright.SelectOptionValues{Values: &[]string{"adults"}})
	page.SelectOption("#cfgBelt", playwright.SelectOptionValues{Values: &[]string{"purple"}})
	page.Fill("#cfgHours", "200")
	page.Fill("#cfgStripes", "4")

	// Click Add Config
	page.Click("button:has-text('Add Config')")
	page.WaitForTimeout(1000)

	// Verify success message appears
	msg := page.Locator("#cfgMsg")
	msgText, _ := msg.TextContent()
	if msgText != "Created!" {
		t.Errorf("expected 'Created!' message, got %q", msgText)
	}

	// Verify config appears in the table
	configTable := page.Locator("#configList table")
	err := configTable.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatal("config table did not appear after adding config")
	}

	tableText, _ := configTable.TextContent()
	if !containsAll(tableText, "adults", "purple", "200") {
		t.Errorf("config table should contain adults/purple/200, got: %s", tableText)
	}
}

// TestGradingConfig_ErrorFeedback tests that server errors are displayed to the user.
func TestGradingConfig_ErrorFeedback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL+"/admin/grading", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	page.WaitForTimeout(500)

	// Verify belt field is a dropdown (select), not a text input
	beltSelect := page.Locator("select#cfgBelt")
	err := beltSelect.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	})
	if err != nil {
		t.Fatal("belt field should be a <select> dropdown, not found")
	}

	// Verify all 9 belt options exist
	options := page.Locator("select#cfgBelt option")
	count, _ := options.Count()
	if count != 9 {
		t.Errorf("expected 9 belt options, got %d", count)
	}

	// Submit with empty program by injecting empty value to trigger server error
	page.Evaluate(`document.getElementById('cfgProgram').innerHTML = '<option value="">empty</option>'`)
	page.SelectOption("#cfgProgram", playwright.SelectOptionValues{Values: &[]string{""}})
	page.Click("button:has-text('Add Config')")

	// Wait for error message
	time.Sleep(1 * time.Second)
	msg := page.Locator("#cfgMsg")
	msgText, _ := msg.TextContent()
	if msgText == "" || msgText == "Created!" {
		t.Errorf("expected error message, got %q", msgText)
	}
}

func containsAll(s string, substrings ...string) bool {
	for _, sub := range substrings {
		found := false
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
