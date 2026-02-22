package browser_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// pageInfo holds health check info for a page
type pageInfo struct {
	path        string
	role        string
	wantStatus  int
	wantTitle   string
	criticalCSS []string // CSS selectors that must exist
}

// TestSmoke_NavigationCrawl verifies all major routes load without errors
func TestSmoke_NavigationCrawl(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Define all routes to test by role
	routes := []pageInfo{
		// Public routes (no auth)
		{path: "/login", role: "", wantStatus: 200, wantTitle: "Login", criticalCSS: []string{"form", "input[name=Email]"}},

		// Admin routes
		{path: "/dashboard", role: "admin", wantStatus: 200, wantTitle: "Dashboard", criticalCSS: []string{"nav", "main"}},
		{path: "/members", role: "admin", wantStatus: 200, wantTitle: "Members", criticalCSS: []string{".member-list, .members-grid, table", "h1"}},
		{path: "/attendance", role: "admin", wantStatus: 200, wantTitle: "Attendance", criticalCSS: []string{"#datePicker", "h1"}},
		{path: "/grading", role: "admin", wantStatus: 200, wantTitle: "Grading", criticalCSS: []string{"h1", "main"}},
		{path: "/accounts", role: "admin", wantStatus: 200, wantTitle: "Accounts", criticalCSS: []string{"table, .accounts-list", "h1"}},
		{path: "/curriculum", role: "admin", wantStatus: 200, wantTitle: "Curriculum", criticalCSS: []string{"h1", "main"}},
		{path: "/calendar", role: "admin", wantStatus: 200, wantTitle: "Calendar", criticalCSS: []string{".calendar-container, #calendar", "h1"}},
		{path: "/class-types", role: "admin", wantStatus: 200, wantTitle: "Class Types", criticalCSS: []string{"h1", "main"}},
		{path: "/injuries", role: "admin", wantStatus: 200, wantTitle: "Injuries", criticalCSS: []string{"h1", "main"}},
		{path: "/waivers", role: "admin", wantStatus: 200, wantTitle: "Waivers", criticalCSS: []string{"h1", "main"}},
		{path: "/themes", role: "admin", wantStatus: 200, wantTitle: "Themes", criticalCSS: []string{"h1", "main"}},
		{path: "/kiosk", role: "admin", wantStatus: 200, wantTitle: "Kiosk", criticalCSS: []string{".kiosk-container, #kiosk", "h1, .kiosk-header"}},
		{path: "/email", role: "admin", wantStatus: 200, wantTitle: "Email", criticalCSS: []string{"h1", "form"}},
		{path: "/library", role: "admin", wantStatus: 200, wantTitle: "Library", criticalCSS: []string{"h1", ".clip-grid, .library-container"}},
		{path: "/goals", role: "admin", wantStatus: 200, wantTitle: "Goals", criticalCSS: []string{"h1", "main"}},

		// Coach routes
		{path: "/dashboard", role: "coach", wantStatus: 200, wantTitle: "Dashboard", criticalCSS: []string{"nav", "main"}},
		{path: "/members", role: "coach", wantStatus: 200, wantTitle: "Members", criticalCSS: []string{".member-list, .members-grid, table", "h1"}},
		{path: "/attendance", role: "coach", wantStatus: 200, wantTitle: "Attendance", criticalCSS: []string{"#datePicker", "h1"}},
		{path: "/grading", role: "coach", wantStatus: 200, wantTitle: "Grading", criticalCSS: []string{"h1", "main"}},
		{path: "/calendar", role: "coach", wantStatus: 200, wantTitle: "Calendar", criticalCSS: []string{".calendar-container, #calendar", "h1"}},
		{path: "/injuries", role: "coach", wantStatus: 200, wantTitle: "Injuries", criticalCSS: []string{"h1", "main"}},
		{path: "/library", role: "coach", wantStatus: 200, wantTitle: "Library", criticalCSS: []string{"h1", ".clip-grid, .library-container"}},

		// Member routes
		{path: "/dashboard", role: "member", wantStatus: 200, wantTitle: "Dashboard", criticalCSS: []string{"nav", "main"}},
		{path: "/attendance", role: "member", wantStatus: 200, wantTitle: "My Attendance", criticalCSS: []string{"h1", "main"}},
		{path: "/goals", role: "member", wantStatus: 200, wantTitle: "Goals", criticalCSS: []string{"h1", "main"}},
		{path: "/library", role: "member", wantStatus: 200, wantTitle: "Library", criticalCSS: []string{"h1", ".clip-grid, .library-container"}},
	}

	// Test each route
	for _, route := range routes {
		route := route // capture range variable
		t.Run(fmt.Sprintf("%s_as_%s", route.path, route.role), func(t *testing.T) {
			page := app.newPage(t)

			// Authenticate if needed
			if route.role != "" {
				app.login(t, page)
				if route.role != "admin" {
					app.impersonate(t, page, route.role)
				}
			}

			// Navigate to page
			resp, err := page.Goto(app.BaseURL + route.path)
			if err != nil {
				t.Errorf("failed to navigate to %s: %v", route.path, err)
				return
			}

			// Check status
			if resp.Status() != route.wantStatus {
				t.Errorf("%s: got status %d, want %d", route.path, resp.Status(), route.wantStatus)
			}

			// Check for critical CSS selectors
			for _, selector := range route.criticalCSS {
				element := page.Locator(selector)
				count, err := element.Count()
				if err != nil || count == 0 {
					t.Errorf("%s: missing critical element %s", route.path, selector)
				}
			}

			// Check page has h1 (except login)
			if route.path != "/login" {
				h1 := page.Locator("h1")
				if visible, _ := h1.IsVisible(); !visible {
					t.Errorf("%s: no visible h1 found", route.path)
				}
			}
		})
	}
}

// TestSmoke_APILiveness verifies API endpoints respond correctly
func TestSmoke_APILiveness(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Test authenticated API endpoints
	endpoints := []struct {
		path   string
		method string
		auth   bool
	}{
		{"/api/members", "GET", true},
		{"/api/attendance/today", "GET", true},
		{"/api/class-types", "GET", true},
		{"/api/calendar/events", "GET", true},
	}

	page := app.newPage(t)
	app.login(t, page)

	for _, ep := range endpoints {
		ep := ep
		t.Run(fmt.Sprintf("%s_%s", ep.method, ep.path), func(t *testing.T) {
			var resp *playwright.APIResponse
			var err error

			switch ep.method {
			case "GET":
				resp, err = page.Context().APIRequest().Get(app.BaseURL + ep.path)
			case "POST":
				resp, err = page.Context().APIRequest().Post(app.BaseURL+ep.path, playwright.APIRequestContextPostOptions{
					Headers: map[string]string{"Content-Type": "application/json"},
					Data:    map[string]interface{}{},
				})
			}

			if err != nil {
				t.Errorf("%s %s failed: %v", ep.method, ep.path, err)
				return
			}

			status := resp.Status()
			// Accept 200 OK or various valid responses
			if status != http.StatusOK && status != http.StatusUnauthorized && status != http.StatusBadRequest {
				t.Errorf("%s %s: unexpected status %d", ep.method, ep.path, status)
			}
		})
	}
}

// TestSmoke_NoConsoleErrors verifies pages load without JavaScript errors
func TestSmoke_NoConsoleErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)

	// Collect console messages
	var errors []string
	page.On("console", func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			errors = append(errors, msg.Text())
		}
	})

	// Navigate to key pages and check for console errors
	pages := []string{
		"/login",
		"/dashboard",
	}

	app.login(t, page)

	for _, path := range pages {
		page.Goto(app.BaseURL + path)
		page.WaitForLoadState(playwright.PageWaitForLoadStateNetworkIdle)
	}

	if len(errors) > 0 {
		t.Errorf("console errors found: %v", errors)
	}
}
