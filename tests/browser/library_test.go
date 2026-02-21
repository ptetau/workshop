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

	// Mute toggle should exist and start muted (button shows "Unmute").
	muteBtn := page.Locator("#muteToggle")
	if err := muteBtn.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(5000)}); err != nil {
		t.Fatalf("mute toggle button not visible: %v", err)
	}
	label, err := muteBtn.InnerText()
	if err != nil {
		t.Fatalf("failed to read mute toggle label: %v", err)
	}
	if !strings.EqualFold(strings.TrimSpace(label), "unmute") {
		t.Fatalf("expected mute toggle label to equal %q (case-insensitive), got %q", "unmute", label)
	}

	// Toggle the UI (postMessage-based YouTube mute/unmute control).
	if err := muteBtn.Click(); err != nil {
		t.Fatalf("failed to click mute toggle: %v", err)
	}
	label2, err := muteBtn.InnerText()
	if err != nil {
		t.Fatalf("failed to read mute toggle label after click: %v", err)
	}
	if !strings.EqualFold(strings.TrimSpace(label2), "mute") {
		t.Fatalf("expected mute toggle label to equal %q after click (case-insensitive), got %q", "mute", label2)
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

// TestLibrary_FilterByTheme verifies that selecting a theme filter shows clips for that theme
// and hides clips from other themes.
func TestLibrary_FilterByTheme(t *testing.T) {
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

	clipTheme := page.Locator("#clipTheme")
	options := clipTheme.Locator("option")
	count, err := options.Count()
	if err != nil {
		t.Fatalf("failed to count theme options: %v", err)
	}
	if count <= 2 {
		t.Skipf("need at least 2 themes to test filter-by-theme, got %d", count-1)
	}

	// Pick two distinct themes.
	theme1, _ := options.Nth(1).GetAttribute("value")
	theme2, _ := options.Nth(2).GetAttribute("value")
	if theme1 == "" || theme2 == "" || theme1 == theme2 {
		t.Skip("theme options were not usable for testing")
	}

	createClip := func(themeID, title string) {
		t.Helper()
		if _, err := clipTheme.SelectOption(playwright.SelectOptionValues{Values: &[]string{themeID}}); err != nil {
			t.Fatalf("failed to select theme: %v", err)
		}
		if err := page.Locator("#clipTitle").Fill(title); err != nil {
			t.Fatalf("failed to fill title: %v", err)
		}
		if err := page.Locator("#clipURL").Fill("https://www.youtube.com/watch?v=dQw4w9WgXcQ"); err != nil {
			t.Fatalf("failed to fill url: %v", err)
		}
		if err := page.Locator("#clipStart").Fill("0:10"); err != nil {
			t.Fatalf("failed to fill start: %v", err)
		}
		if err := page.Locator("#clipEnd").Fill("0:20"); err != nil {
			t.Fatalf("failed to fill end: %v", err)
		}
		if err := page.Locator("#clipNotes").Fill("notes"); err != nil {
			t.Fatalf("failed to fill notes: %v", err)
		}
		if err := page.Locator("#createClipForm button[type=submit]").Click(); err != nil {
			t.Fatalf("failed to submit create clip form: %v", err)
		}

		clipTitle := page.Locator("#clipGrid strong").Filter(playwright.LocatorFilterOptions{HasText: title})
		if err := clipTitle.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(5000)}); err != nil {
			t.Fatalf("created clip did not appear in grid: %v", err)
		}
	}

	title1 := "Filter Theme Test - Clip A"
	title2 := "Filter Theme Test - Clip B"
	createClip(theme1, title1)
	createClip(theme2, title2)

	// Filter by theme1 and verify clip2 is not shown.
	filterTheme := page.Locator("#filterTheme")
	if _, err := filterTheme.SelectOption(playwright.SelectOptionValues{Values: &[]string{theme1}}); err != nil {
		t.Fatalf("failed to select filter theme: %v", err)
	}
	page.WaitForTimeout(1000)

	clip1 := page.Locator("#clipGrid strong").Filter(playwright.LocatorFilterOptions{HasText: title1})
	if err := clip1.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(5000)}); err != nil {
		t.Fatalf("expected clip1 to be visible after filtering by theme1: %v", err)
	}
	clip2 := page.Locator("#clipGrid strong").Filter(playwright.LocatorFilterOptions{HasText: title2})
	clip2Count, err := clip2.Count()
	if err != nil {
		t.Fatalf("failed to count clip2 after filtering: %v", err)
	}
	if clip2Count != 0 {
		t.Fatalf("expected clip2 to be hidden after filtering by theme1, but found %d matches", clip2Count)
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
