package browser_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"

	memberDomain "workshop/internal/domain/member"
)

// TestListView_BookmarkFilteredView tests US-1.5.6: Bookmark a filtered view.
// Opening a URL with query params pre-set should display the matching filtered/sorted view.
func TestListView_BookmarkFilteredView(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Seed specific members
	ctx := context.Background()
	members := []memberDomain.Member{
		{ID: uuid.New().String(), Name: "Alice Smith", Email: "alice@test.com", Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly"},
		{ID: uuid.New().String(), Name: "Bob Jones", Email: "bob@test.com", Program: "Kids", Status: "active", Fee: 80, Frequency: "monthly"},
		{ID: uuid.New().String(), Name: "Charlie Smith", Email: "charlie@test.com", Program: "Adults", Status: "trial", Fee: 100, Frequency: "monthly"},
		{ID: uuid.New().String(), Name: "Diana Lee", Email: "diana@test.com", Program: "Kids", Status: "archived", Fee: 80, Frequency: "monthly"},
	}
	for _, m := range members {
		if err := app.Stores.MemberStore.Save(ctx, m); err != nil {
			t.Fatalf("failed to seed: %v", err)
		}
	}

	// Test 1: Bookmark with search filter — open URL with q=Smith directly
	_, err := page.Goto(app.BaseURL + "/members?q=Smith")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	summary := page.Locator("#results-summary")
	text, _ := summary.TextContent()
	if !strings.Contains(text, "of 2") {
		t.Errorf("bookmarked search q=Smith: expected 2 results, got: %s", strings.TrimSpace(text))
	}
	// Verify search input is pre-filled
	searchVal, _ := page.Locator("#search-input").InputValue()
	if searchVal != "Smith" {
		t.Errorf("expected search input to show 'Smith', got: %q", searchVal)
	}

	// Test 2: Bookmark with program + status filters
	_, err = page.Goto(app.BaseURL + "/members?program=Adults&status=active")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	text, _ = summary.TextContent()
	if !strings.Contains(text, "of 1") {
		t.Errorf("bookmarked Adults+active: expected 1 result, got: %s", strings.TrimSpace(text))
	}
	// Verify dropdowns are pre-selected
	progVal, _ := page.Locator("#program-filter").InputValue()
	if progVal != "Adults" {
		t.Errorf("expected program dropdown to show Adults, got: %q", progVal)
	}
	statusVal, _ := page.Locator("#status-filter").InputValue()
	if statusVal != "active" {
		t.Errorf("expected status dropdown to show active, got: %q", statusVal)
	}

	// Test 3: Bookmark with sort params — sort by name desc
	_, err = page.Goto(app.BaseURL + "/members?sort=name&dir=desc")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	firstRow := page.Locator("table tbody tr:first-child td:first-child")
	text, _ = firstRow.TextContent()
	if !strings.Contains(text, "Diana") {
		t.Errorf("bookmarked sort=name&dir=desc: expected Diana first, got: %s", strings.TrimSpace(text))
	}

	// Test 4: Bookmark with per_page — use valid option (10)
	_, err = page.Goto(app.BaseURL + "/members?per_page=10")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	text, _ = summary.TextContent()
	if !strings.Contains(text, "1-4 of 4") {
		t.Errorf("bookmarked per_page=10: expected '1-4 of 4', got: %s", strings.TrimSpace(text))
	}
	perPageVal, _ := page.Locator("#per-page-select").InputValue()
	if perPageVal != "10" {
		t.Errorf("expected per_page dropdown to show 10, got: %q", perPageVal)
	}

	// Test 5: Full bookmark with all params combined
	_, err = page.Goto(app.BaseURL + "/members?q=Smith&program=Adults&sort=name&dir=asc&per_page=10")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	text, _ = summary.TextContent()
	if !strings.Contains(text, "of 2") {
		t.Errorf("full bookmark: expected 2 results (Smith+Adults), got: %s", strings.TrimSpace(text))
	}
	// First row should be Alice (name asc, both Smiths are Adults)
	firstRow = page.Locator("table tbody tr:first-child td:first-child")
	text, _ = firstRow.TextContent()
	if !strings.Contains(text, "Alice") {
		t.Errorf("full bookmark: expected Alice first (name asc), got: %s", strings.TrimSpace(text))
	}

	// Test 6: Browser back/forward preserves state
	_, err = page.Goto(app.BaseURL + "/members")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	_, err = page.Goto(app.BaseURL + "/members?program=Kids")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	text, _ = summary.TextContent()
	if !strings.Contains(text, "of 2") {
		t.Errorf("Kids filter: expected 2, got: %s", strings.TrimSpace(text))
	}
	// Go back
	page.GoBack()
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})
	text, _ = summary.TextContent()
	if !strings.Contains(text, "of 4") {
		t.Errorf("after back: expected all 4 members, got: %s", strings.TrimSpace(text))
	}
	// Go forward
	page.GoForward()
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})
	text, _ = summary.TextContent()
	if !strings.Contains(text, "of 2") {
		t.Errorf("after forward: expected 2 Kids, got: %s", strings.TrimSpace(text))
	}
}

// seedMembers creates n test members with varied programs and statuses.
func seedMembers(t *testing.T, app *testApp, n int) {
	t.Helper()
	ctx := context.Background()
	programs := []string{"Adults", "Kids"}
	statuses := []string{"active", "trial", "archived"}

	for i := 0; i < n; i++ {
		m := memberDomain.Member{
			ID:        uuid.New().String(),
			Name:      fmt.Sprintf("Member %03d", i+1),
			Email:     fmt.Sprintf("member%03d@test.com", i+1),
			Program:   programs[i%2],
			Status:    statuses[i%3],
			Fee:       100,
			Frequency: "monthly",
		}
		if err := app.Stores.MemberStore.Save(ctx, m); err != nil {
			t.Fatalf("failed to seed member %d: %v", i, err)
		}
	}
}

// TestListView_Pagination tests US-1.5.1: Paginate a list
func TestListView_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Seed 25 members (more than default per_page of 20)
	seedMembers(t, app, 25)

	// Navigate to member list
	_, err := page.Goto(app.BaseURL + "/members")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Verify results summary shows "Showing 1-20 of 25 results"
	summary := page.Locator("#results-summary")
	text, err := summary.TextContent()
	if err != nil {
		t.Fatalf("failed to get summary text: %v", err)
	}
	if !strings.Contains(text, "1-20 of 25") {
		t.Errorf("expected '1-20 of 25' in summary, got: %s", strings.TrimSpace(text))
	}

	// Verify pagination is visible
	nav := page.Locator("nav[aria-label='Pagination']")
	visible, err := nav.IsVisible()
	if err != nil {
		t.Fatalf("failed to check pagination visibility: %v", err)
	}
	if !visible {
		t.Error("pagination should be visible with 25 members at per_page=20")
	}

	// Click "Next" to go to page 2
	nextLink := page.Locator("nav[aria-label='Pagination'] a:has-text('Next')")
	if err := nextLink.Click(); err != nil {
		t.Fatalf("failed to click Next: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	// Verify page 2 shows remaining 5 members
	text, _ = summary.TextContent()
	if !strings.Contains(text, "21-25 of 25") {
		t.Errorf("expected '21-25 of 25' on page 2, got: %s", strings.TrimSpace(text))
	}

	// Verify URL contains page=2
	if !strings.Contains(page.URL(), "page=2") {
		t.Errorf("expected page=2 in URL, got: %s", page.URL())
	}

	// Click "Previous" to go back to page 1
	prevLink := page.Locator("nav[aria-label='Pagination'] a:has-text('Previous')")
	if err := prevLink.Click(); err != nil {
		t.Fatalf("failed to click Previous: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	text, _ = summary.TextContent()
	if !strings.Contains(text, "1-20 of 25") {
		t.Errorf("expected '1-20 of 25' after clicking Previous, got: %s", strings.TrimSpace(text))
	}

	// Click "Last" to jump to the last page
	lastLink := page.Locator("nav[aria-label='Pagination'] a:has-text('Last')")
	if err := lastLink.Click(); err != nil {
		t.Fatalf("failed to click Last: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	text, _ = summary.TextContent()
	if !strings.Contains(text, "21-25 of 25") {
		t.Errorf("expected '21-25 of 25' on last page, got: %s", strings.TrimSpace(text))
	}

	// Click "First" to jump back
	firstLink := page.Locator("nav[aria-label='Pagination'] a:has-text('First')")
	if err := firstLink.Click(); err != nil {
		t.Fatalf("failed to click First: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	text, _ = summary.TextContent()
	if !strings.Contains(text, "1-20 of 25") {
		t.Errorf("expected '1-20 of 25' after clicking First, got: %s", strings.TrimSpace(text))
	}

	// Verify pagination hidden when total <= per_page
	_, err = page.Goto(app.BaseURL + "/members?per_page=50")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}
	visible, _ = nav.IsVisible()
	if visible {
		t.Error("pagination should be hidden when total (25) <= per_page (50)")
	}
}

// TestListView_RowCount tests US-1.5.2: Change row count
func TestListView_RowCount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Seed 25 members
	seedMembers(t, app, 25)

	// Navigate to member list
	_, err := page.Goto(app.BaseURL + "/members")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Default should show 20 per page
	summary := page.Locator("#results-summary")
	text, _ := summary.TextContent()
	if !strings.Contains(text, "1-20 of 25") {
		t.Errorf("expected default per_page=20, got: %s", strings.TrimSpace(text))
	}

	// Change to 10 per page
	perPageSelect := page.Locator("#per-page-select")
	if _, err := perPageSelect.SelectOption(playwright.SelectOptionValues{Values: &[]string{"10"}}); err != nil {
		t.Fatalf("failed to select per_page=10: %v", err)
	}
	if err := page.WaitForURL("**/members?*per_page=10*", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("failed to wait for per_page=10 navigation: %v", err)
	}

	summary = page.Locator("#results-summary")
	text, _ = summary.TextContent()
	if !strings.Contains(text, "1-10 of 25") {
		t.Errorf("expected '1-10 of 25' with per_page=10, got: %s", strings.TrimSpace(text))
	}

	// Verify URL contains per_page=10
	if !strings.Contains(page.URL(), "per_page=10") {
		t.Errorf("expected per_page=10 in URL, got: %s", page.URL())
	}

	// Change to 50 per page — should show all 25 and hide pagination
	perPageSelect = page.Locator("#per-page-select")
	if _, err := perPageSelect.SelectOption(playwright.SelectOptionValues{Values: &[]string{"50"}}); err != nil {
		t.Fatalf("failed to select per_page=50: %v", err)
	}
	if err := page.WaitForURL("**/members?*per_page=50*", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("failed to wait for per_page=50 navigation: %v", err)
	}

	summary = page.Locator("#results-summary")
	text, _ = summary.TextContent()
	if !strings.Contains(text, "1-25 of 25") {
		t.Errorf("expected '1-25 of 25' with per_page=50, got: %s", strings.TrimSpace(text))
	}

	// Changing row count should reset to page 1
	_, err = page.Goto(app.BaseURL + "/members?page=2&per_page=10")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Now change per_page — form submits without page param, so it defaults to page 1
	perPageSelect = page.Locator("#per-page-select")
	if _, err := perPageSelect.SelectOption(playwright.SelectOptionValues{Values: &[]string{"20"}}); err != nil {
		t.Fatalf("failed to select per_page=20: %v", err)
	}
	if err := page.WaitForURL("**/members?*per_page=20*", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("failed to wait for per_page=20 navigation: %v", err)
	}

	// Should be on page 1
	summary = page.Locator("#results-summary")
	text, _ = summary.TextContent()
	if !strings.Contains(text, "1-20 of 25") {
		t.Errorf("expected page reset to 1 after changing per_page, got: %s", strings.TrimSpace(text))
	}
}

// TestListView_SortByColumn tests US-1.5.3: Sort by column
func TestListView_SortByColumn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Seed members with known names for predictable sort order
	ctx := context.Background()
	names := []string{"Charlie", "Alice", "Bob"}
	for i, name := range names {
		m := memberDomain.Member{
			ID:        uuid.New().String(),
			Name:      name,
			Email:     fmt.Sprintf("%s@test.com", strings.ToLower(name)),
			Program:   "Adults",
			Status:    "active",
			Fee:       100,
			Frequency: "monthly",
		}
		_ = i
		if err := app.Stores.MemberStore.Save(ctx, m); err != nil {
			t.Fatalf("failed to seed member: %v", err)
		}
	}

	// Navigate — default sort is name ASC
	_, err := page.Goto(app.BaseURL + "/members")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// First row should be Alice (alphabetically first)
	firstRow := page.Locator("table tbody tr:first-child td:first-child")
	text, err := firstRow.TextContent()
	if err != nil {
		t.Fatalf("failed to get first row: %v", err)
	}
	if !strings.Contains(text, "Alice") {
		t.Errorf("expected Alice first (name ASC default), got: %s", strings.TrimSpace(text))
	}

	// Click "Name" sort header to sort name ASC (already default, so this should toggle to desc)
	nameHeader := page.Locator("th a:has-text('Name')")
	if err := nameHeader.Click(); err != nil {
		t.Fatalf("failed to click Name header: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	// Verify URL has sort=name
	if !strings.Contains(page.URL(), "sort=name") {
		t.Errorf("expected sort=name in URL, got: %s", page.URL())
	}

	// Click again to toggle to desc
	nameHeader = page.Locator("th a:has-text('Name')")
	if err := nameHeader.Click(); err != nil {
		t.Fatalf("failed to click Name header again: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	// Verify URL has dir=desc
	if !strings.Contains(page.URL(), "dir=desc") {
		t.Errorf("expected dir=desc in URL, got: %s", page.URL())
	}

	// First row should now be Charlie (name DESC)
	firstRow = page.Locator("table tbody tr:first-child td:first-child")
	text, _ = firstRow.TextContent()
	if !strings.Contains(text, "Charlie") {
		t.Errorf("expected Charlie first (name DESC), got: %s", strings.TrimSpace(text))
	}

	// Sort by email column
	emailHeader := page.Locator("th a:has-text('Email')")
	if err := emailHeader.Click(); err != nil {
		t.Fatalf("failed to click Email header: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	if !strings.Contains(page.URL(), "sort=email") {
		t.Errorf("expected sort=email in URL, got: %s", page.URL())
	}
}

// TestListView_SearchAndFilter tests US-1.5.4: Search and filter
func TestListView_SearchAndFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Seed specific members for testing search and filters
	ctx := context.Background()
	members := []memberDomain.Member{
		{ID: uuid.New().String(), Name: "Alice Smith", Email: "alice@test.com", Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly"},
		{ID: uuid.New().String(), Name: "Bob Jones", Email: "bob@test.com", Program: "Kids", Status: "active", Fee: 80, Frequency: "monthly"},
		{ID: uuid.New().String(), Name: "Charlie Smith", Email: "charlie@test.com", Program: "Adults", Status: "trial", Fee: 100, Frequency: "monthly"},
		{ID: uuid.New().String(), Name: "Diana Lee", Email: "diana@test.com", Program: "Kids", Status: "archived", Fee: 80, Frequency: "monthly"},
	}
	for _, m := range members {
		if err := app.Stores.MemberStore.Save(ctx, m); err != nil {
			t.Fatalf("failed to seed member: %v", err)
		}
	}

	// Navigate to member list
	_, err := page.Goto(app.BaseURL + "/members")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Test search by name — submit the form manually (not relying on debounce timer)
	searchInput := page.Locator("#search-input")
	if err := searchInput.Fill("Smith"); err != nil {
		t.Fatalf("failed to fill search: %v", err)
	}
	submitBtn := page.Locator("button[type=submit]:has-text('Search')")
	if err := submitBtn.Click(); err != nil {
		t.Fatalf("failed to click Search: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	// Should find Alice Smith and Charlie Smith
	summary := page.Locator("#results-summary")
	text, _ := summary.TextContent()
	if !strings.Contains(text, "of 2") {
		t.Errorf("expected 2 results for 'Smith' search, got: %s", strings.TrimSpace(text))
	}

	// Verify URL contains q=Smith
	if !strings.Contains(page.URL(), "q=Smith") {
		t.Errorf("expected q=Smith in URL, got: %s", page.URL())
	}

	// Clear search and test program filter
	_, err = page.Goto(app.BaseURL + "/members")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Select "Adults" program filter
	programFilter := page.Locator("#program-filter")
	if _, err := programFilter.SelectOption(playwright.SelectOptionValues{Values: &[]string{"Adults"}}); err != nil {
		t.Fatalf("failed to select Adults: %v", err)
	}
	// Selecting a filter submits the form and navigates. Waiting on load state alone can be flaky
	// (it may already be satisfied from the current page), so wait for the URL to change.
	if err := page.WaitForURL("**/members?*program=Adults*", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("failed to wait for Adults filter navigation: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateDomcontentloaded})

	text, _ = summary.TextContent()
	if !strings.Contains(text, "of 2") {
		t.Errorf("expected 2 Adults, got: %s", strings.TrimSpace(text))
	}

	// Test status filter
	_, err = page.Goto(app.BaseURL + "/members")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	statusFilter := page.Locator("#status-filter")
	if _, err := statusFilter.SelectOption(playwright.SelectOptionValues{Values: &[]string{"active"}}); err != nil {
		t.Fatalf("failed to select active: %v", err)
	}
	if err := page.WaitForURL("**/members?*status=active*", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("failed to wait for status filter navigation: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateDomcontentloaded})

	text, _ = summary.TextContent()
	if !strings.Contains(text, "of 2") {
		t.Errorf("expected 2 active members, got: %s", strings.TrimSpace(text))
	}

	// Test combined filters: program=Adults AND status=active => Alice only
	_, err = page.Goto(app.BaseURL + "/members?program=Adults&status=active")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	text, _ = summary.TextContent()
	if !strings.Contains(text, "of 1") {
		t.Errorf("expected 1 result for Adults+active, got: %s", strings.TrimSpace(text))
	}

	// Verify filter resets to page 1: seed extra members so page 2 exists
	for i := 10; i < 25; i++ {
		m := memberDomain.Member{
			ID: uuid.New().String(), Name: fmt.Sprintf("Extra %d", i),
			Email: fmt.Sprintf("extra%d@test.com", i), Program: "Adults",
			Status: "active", Fee: 100, Frequency: "monthly",
		}
		if err := app.Stores.MemberStore.Save(ctx, m); err != nil {
			t.Fatalf("failed to seed extra member: %v", err)
		}
	}

	_, err = page.Goto(app.BaseURL + "/members?page=2&per_page=10")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Confirm we're on page 2
	if !strings.Contains(page.URL(), "page=2") {
		t.Fatalf("expected page=2 in URL before filter change, got: %s", page.URL())
	}

	programFilter = page.Locator("#program-filter")
	if _, err := programFilter.SelectOption(playwright.SelectOptionValues{Values: &[]string{"Kids"}}); err != nil {
		t.Fatalf("failed to select Kids filter: %v", err)
	}

	// Wait for the form to submit and page to navigate
	if err := page.WaitForURL("**/members?*program=Kids*", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("failed to wait for filter navigation: %v", err)
	}

	// The form submission shouldn't include page param, so defaults to 1
	if strings.Contains(page.URL(), "page=2") {
		t.Error("applying filter should reset to page 1")
	}
}

// TestListView_ClearFilters tests US-1.5.5: Clear filters
func TestListView_ClearFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Seed members
	ctx := context.Background()
	members := []memberDomain.Member{
		{ID: uuid.New().String(), Name: "Alice", Email: "alice@test.com", Program: "Adults", Status: "active", Fee: 100, Frequency: "monthly"},
		{ID: uuid.New().String(), Name: "Bob", Email: "bob@test.com", Program: "Kids", Status: "active", Fee: 80, Frequency: "monthly"},
		{ID: uuid.New().String(), Name: "Charlie", Email: "charlie@test.com", Program: "Adults", Status: "trial", Fee: 100, Frequency: "monthly"},
	}
	for _, m := range members {
		if err := app.Stores.MemberStore.Save(ctx, m); err != nil {
			t.Fatalf("failed to seed: %v", err)
		}
	}

	// Navigate with filters applied
	_, err := page.Goto(app.BaseURL + "/members?q=Alice&program=Adults&status=active")
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Verify Clear button is visible
	clearBtn := page.Locator("#clear-filters")
	visible, err := clearBtn.IsVisible()
	if err != nil {
		t.Fatalf("failed to check clear button visibility: %v", err)
	}
	if !visible {
		t.Error("Clear button should be visible when filters are active")
	}

	// Verify filtered results (Alice only)
	summary := page.Locator("#results-summary")
	text, _ := summary.TextContent()
	if !strings.Contains(text, "of 1") {
		t.Errorf("expected 1 filtered result, got: %s", strings.TrimSpace(text))
	}

	// Click Clear
	if err := clearBtn.Click(); err != nil {
		t.Fatalf("failed to click Clear: %v", err)
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	// Should be back to /members with no query params
	currentURL := page.URL()
	if strings.Contains(currentURL, "q=") || strings.Contains(currentURL, "program=") || strings.Contains(currentURL, "status=") {
		t.Errorf("expected clean URL after clear, got: %s", currentURL)
	}

	// All 3 members should be visible
	text, _ = summary.TextContent()
	if !strings.Contains(text, "of 3") {
		t.Errorf("expected all 3 members after clear, got: %s", strings.TrimSpace(text))
	}

	// Verify Clear button is hidden when no filters are active
	clearBtn = page.Locator("#clear-filters")
	visible, _ = clearBtn.IsVisible()
	if visible {
		t.Error("Clear button should be hidden when no filters are active")
	}
}
