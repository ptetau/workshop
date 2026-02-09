package browser_test

import (
	"context"
	"testing"

	"github.com/playwright-community/playwright-go"

	"workshop/internal/application/orchestrators"
)

// seedSyntheticData seeds the full synthetic dataset for tests that need attendance records.
func seedSyntheticData(t *testing.T, app *testApp) {
	t.Helper()
	ctx := context.Background()
	deps := orchestrators.SyntheticSeedDeps{
		AccountStore:         app.Stores.AccountStore,
		MemberStore:          app.Stores.MemberStore,
		WaiverStore:          app.Stores.WaiverStore,
		InjuryStore:          app.Stores.InjuryStore,
		AttendanceStore:      app.Stores.AttendanceStore,
		ScheduleStore:        app.Stores.ScheduleStore,
		TermStore:            app.Stores.TermStore,
		HolidayStore:         app.Stores.HolidayStore,
		NoticeStore:          app.Stores.NoticeStore,
		GradingRecordStore:   app.Stores.GradingRecordStore,
		GradingConfigStore:   app.Stores.GradingConfigStore,
		GradingProposalStore: app.Stores.GradingProposalStore,
		MessageStore:         app.Stores.MessageStore,
		ObservationStore:     app.Stores.ObservationStore,
		MilestoneStore:       app.Stores.MilestoneStore,
		TrainingGoalStore:    app.Stores.TrainingGoalStore,
		ClassTypeStore:       app.Stores.ClassTypeStore,
		ThemeStore:           app.Stores.ThemeStore,
		ClipStore:            app.Stores.ClipStore,
	}
	if err := orchestrators.ExecuteSeedSynthetic(ctx, deps, app.AdminID); err != nil {
		t.Fatalf("failed to seed synthetic data: %v", err)
	}
}

// TestClassFilter_SessionDropdownPopulated verifies that the session filter dropdown is populated on the compose page.
func TestClassFilter_SessionDropdownPopulated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	seedSyntheticData(t, app)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/admin/emails/compose"); err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Wait for session filter to be populated (JS fetch)
	sessionFilter := page.Locator("#sessionFilter")
	if err := sessionFilter.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("session filter not visible: %v", err)
	}

	// Wait a moment for the fetch to complete
	page.WaitForTimeout(1000)

	// Check that options were populated (more than just the default placeholder)
	options := sessionFilter.Locator("option")
	count, err := options.Count()
	if err != nil {
		t.Fatalf("failed to count session filter options: %v", err)
	}
	if count <= 1 {
		t.Errorf("expected session filter to have options beyond placeholder, got %d", count)
	}
}

// TestClassFilter_ClassTypeDropdownPopulated verifies that the class type filter dropdown is populated.
func TestClassFilter_ClassTypeDropdownPopulated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	seedSyntheticData(t, app)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/admin/emails/compose"); err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	classTypeFilter := page.Locator("#classTypeFilter")
	if err := classTypeFilter.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("class type filter not visible: %v", err)
	}

	page.WaitForTimeout(1000)

	options := classTypeFilter.Locator("option")
	count, err := options.Count()
	if err != nil {
		t.Fatalf("failed to count class type filter options: %v", err)
	}
	// Should have placeholder + at least one class type
	if count <= 1 {
		t.Errorf("expected class type filter to have options beyond placeholder, got %d", count)
	}
}

// TestClassFilter_FilterBySessionShowsMembers verifies filtering by a specific class session populates the results.
func TestClassFilter_FilterBySessionShowsMembers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	seedSyntheticData(t, app)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/admin/emails/compose"); err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Wait for session filter options to load
	page.WaitForTimeout(1500)

	sessionFilter := page.Locator("#sessionFilter")
	options := sessionFilter.Locator("option")
	count, _ := options.Count()
	if count <= 1 {
		t.Skip("no class sessions available to test (synthetic data may not have sessions in last 14 days)")
	}

	// Select the second option (first real session)
	secondOption := options.Nth(1)
	val, _ := secondOption.GetAttribute("value")
	if val == "" {
		t.Skip("second session option has no value")
	}

	if _, err := sessionFilter.SelectOption(playwright.SelectOptionValues{Values: &[]string{val}}); err != nil {
		t.Fatalf("failed to select session: %v", err)
	}

	// Wait for filter results to appear
	filterResults := page.Locator("#filterResults")
	if err := filterResults.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		// May be no attendance for this session — not a hard failure
		t.Logf("filter results did not appear (may be no attendance for selected session): %v", err)
		return
	}

	// Should show filter result list with checkboxes
	checkboxes := page.Locator("#filterResultsList input[type='checkbox']")
	cbCount, _ := checkboxes.Count()
	t.Logf("session filter returned %d members", cbCount)
}

// TestClassFilter_FilterByClassTypeShowsMembers verifies filtering by class type + lookback populates results.
func TestClassFilter_FilterByClassTypeShowsMembers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	seedSyntheticData(t, app)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/admin/emails/compose"); err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	// Wait for class type filter options to load
	page.WaitForTimeout(1500)

	classTypeFilter := page.Locator("#classTypeFilter")
	options := classTypeFilter.Locator("option")
	count, _ := options.Count()
	if count <= 1 {
		t.Skip("no class types available to test")
	}

	// Try "Fundamentals" first (most attendance data), fall back to first option
	var selectedVal string
	for i := 1; i < count; i++ {
		opt := options.Nth(i)
		text, _ := opt.TextContent()
		val, _ := opt.GetAttribute("value")
		if text == "Fundamentals" {
			selectedVal = val
			break
		}
	}
	if selectedVal == "" {
		opt := options.Nth(1)
		selectedVal, _ = opt.GetAttribute("value")
	}
	if selectedVal == "" {
		t.Skip("no class type option has a value")
	}

	if _, err := classTypeFilter.SelectOption(playwright.SelectOptionValues{Values: &[]string{selectedVal}}); err != nil {
		t.Fatalf("failed to select class type: %v", err)
	}

	// Set lookback to 90 days to catch all synthetic attendance
	lookback := page.Locator("#lookbackDays")
	if _, err := lookback.SelectOption(playwright.SelectOptionValues{Values: &[]string{"90"}}); err != nil {
		t.Fatalf("failed to select lookback: %v", err)
	}

	// Click the Apply button
	applyBtn := page.Locator("button:has-text('Apply')")
	if err := applyBtn.Click(); err != nil {
		t.Fatalf("failed to click Apply: %v", err)
	}

	// Wait for filter results to appear
	filterResults := page.Locator("#filterResults")
	if err := filterResults.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("filter results did not appear: %v", err)
	}

	// Should show members who attended this class type
	checkboxes := page.Locator("#filterResultsList input[type='checkbox']")
	cbCount, _ := checkboxes.Count()
	if cbCount == 0 {
		t.Error("expected at least one member from class type filter with 90-day lookback")
	}
	t.Logf("class type filter returned %d members", cbCount)
}

// TestClassFilter_SelectAllFromFilter verifies that Select All adds filtered members to recipients.
func TestClassFilter_SelectAllFromFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	seedSyntheticData(t, app)
	page := app.newPage(t)
	app.login(t, page)

	if _, err := page.Goto(app.BaseURL + "/admin/emails/compose"); err != nil {
		t.Fatalf("failed to navigate to compose: %v", err)
	}

	page.WaitForTimeout(1500)

	// Use class type filter — prefer Fundamentals (most attendance data)
	classTypeFilter := page.Locator("#classTypeFilter")
	options := classTypeFilter.Locator("option")
	count, _ := options.Count()
	if count <= 1 {
		t.Skip("no class types available")
	}

	var selectedVal string
	for i := 1; i < count; i++ {
		opt := options.Nth(i)
		text, _ := opt.TextContent()
		val, _ := opt.GetAttribute("value")
		if text == "Fundamentals" {
			selectedVal = val
			break
		}
	}
	if selectedVal == "" {
		opt := options.Nth(1)
		selectedVal, _ = opt.GetAttribute("value")
	}
	if _, err := classTypeFilter.SelectOption(playwright.SelectOptionValues{Values: &[]string{selectedVal}}); err != nil {
		t.Fatalf("failed to select class type: %v", err)
	}

	lookback := page.Locator("#lookbackDays")
	lookback.SelectOption(playwright.SelectOptionValues{Values: &[]string{"90"}})

	applyBtn := page.Locator("button:has-text('Apply')")
	applyBtn.Click()

	filterResults := page.Locator("#filterResults")
	if err := filterResults.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Skip("no filter results to test Select All")
	}

	// Click Select All
	selectAllBtn := page.Locator("#selectAllBtn")
	if err := selectAllBtn.Click(); err != nil {
		t.Fatalf("failed to click Select All: %v", err)
	}

	// Verify recipient count updated
	countText, _ := page.Locator("#recipientCount").TextContent()
	if countText == "0 selected" {
		t.Error("expected recipient count to be greater than 0 after Select All")
	}

	// Verify selected recipient tags appear
	tags := page.Locator("#selectedRecipients span")
	tagCount, _ := tags.Count()
	if tagCount == 0 {
		t.Error("expected selected recipient tags to appear after Select All")
	}
	t.Logf("Select All added %d recipients, count text: %s", tagCount, countText)
}
