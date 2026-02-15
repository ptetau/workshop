package browser_test

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// TestLibrary_CreateClipAndPlayMuted verifies a user can create a clip using mm:ss timestamps
// and that playback starts muted by default.
func TestLibrary_CreateClipAndPlayMuted(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	seedSyntheticData(t, app)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/library"); err != nil {
		t.Fatalf("failed to navigate to library: %v", err)
	}

	// Wait for themes/clips JS to load.
	page.WaitForTimeout(1000)

	// Open "Add a Clip".
	if err := page.Locator("summary:has-text('Add a Clip')").Click(); err != nil {
		t.Fatalf("failed to open add clip section: %v", err)
	}

	// Select the first real theme option.
	clipTheme := page.Locator("#clipTheme")
	options := clipTheme.Locator("option")
	count, err := options.Count()
	if err != nil {
		t.Fatalf("failed to count theme options: %v", err)
	}
	if count <= 1 {
		t.Skip("no themes available to create a clip")
	}
	val, _ := options.Nth(1).GetAttribute("value")
	if val == "" {
		t.Skip("theme option had empty value")
	}
	if _, err := clipTheme.SelectOption(playwright.SelectOptionValues{Values: &[]string{val}}); err != nil {
		t.Fatalf("failed to select theme: %v", err)
	}

	title := "Test Clip - 12:34 to 12:44"
	if err := page.Locator("#clipTitle").Fill(title); err != nil {
		t.Fatalf("failed to fill title: %v", err)
	}
	if err := page.Locator("#clipURL").Fill("https://www.youtube.com/watch?v=dQw4w9WgXcQ"); err != nil {
		t.Fatalf("failed to fill url: %v", err)
	}
	if err := page.Locator("#clipStart").Fill("12:34"); err != nil {
		t.Fatalf("failed to fill start: %v", err)
	}
	if err := page.Locator("#clipEnd").Fill("12:44"); err != nil {
		t.Fatalf("failed to fill end: %v", err)
	}
	if err := page.Locator("#clipNotes").Fill("notes"); err != nil {
		t.Fatalf("failed to fill notes: %v", err)
	}

	if err := page.Locator("#createClipForm button[type=submit]").Click(); err != nil {
		t.Fatalf("failed to submit create clip form: %v", err)
	}

	// Wait for the created clip to appear in the grid.
	clipTitle := page.Locator("#clipGrid strong").Filter(playwright.LocatorFilterOptions{HasText: title})
	if err := clipTitle.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(5000)}); err != nil {
		t.Fatalf("created clip did not appear in grid: %v", err)
	}

	// Click it to open player.
	if err := clipTitle.Click(); err != nil {
		t.Fatalf("failed to click created clip: %v", err)
	}

	player := page.Locator("#clipPlayer")
	if err := player.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(5000)}); err != nil {
		t.Fatalf("player did not open: %v", err)
	}

	src, err := page.Locator("#ytPlayer").GetAttribute("src")
	if err != nil {
		t.Fatalf("failed to get iframe src: %v", err)
	}
	if src == "" {
		t.Fatalf("expected iframe src to be populated")
	}
	// 12:34 -> 754s, 12:44 -> 764s.
	for _, want := range []string{"mute=1", "start=754", "end=764", "loop=1"} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected iframe src to contain %q, got %q", want, src)
		}
	}
}

// TestLibrary_TrialCannotAccess verifies Trial role cannot access /library per PRD.
func TestLibrary_TrialCannotAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	seedSyntheticData(t, app)
	page := app.newPage(t)
	app.login(t, page)

	app.impersonate(t, page, "trial")

	if _, err := page.Goto(app.BaseURL + "/library"); err != nil {
		t.Fatalf("failed to navigate to library as trial: %v", err)
	}
	if err := page.WaitForURL(app.BaseURL+"/dashboard", playwright.PageWaitForURLOptions{Timeout: playwright.Float(5000)}); err != nil {
		t.Fatalf("expected trial to be redirected to dashboard: %v", err)
	}
}
