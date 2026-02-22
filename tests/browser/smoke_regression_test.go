package browser_test

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// TestSmoke_NavigationCrawl verifies all major routes load without errors
func TestSmoke_NavigationCrawl(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)

	// Define all routes to test by role - using actual routes from routes.go
	routes := []struct {
		path       string
		role       string
		wantStatus int
	}{
		// Public routes (no auth)
		{path: "/login", role: "", wantStatus: 200},

		// Admin routes
		{path: "/dashboard", role: "admin", wantStatus: 200},
		{path: "/members", role: "admin", wantStatus: 200},
		{path: "/attendance", role: "admin", wantStatus: 200},
		{path: "/admin/grading", role: "admin", wantStatus: 200},
		{path: "/admin/accounts", role: "admin", wantStatus: 200},
		{path: "/curriculum", role: "admin", wantStatus: 200},
		{path: "/calendar", role: "admin", wantStatus: 200},
		{path: "/admin/class-types", role: "admin", wantStatus: 200},
		{path: "/admin/notices", role: "admin", wantStatus: 200},
		{path: "/admin/schedules", role: "admin", wantStatus: 200},
		{path: "/admin/holidays", role: "admin", wantStatus: 200},
		{path: "/admin/terms", role: "admin", wantStatus: 200},
		{path: "/admin/features", role: "admin", wantStatus: 200},
		{path: "/admin/inactive", role: "admin", wantStatus: 200},
		{path: "/admin/milestones", role: "admin", wantStatus: 200},
		{path: "/admin/emails", role: "admin", wantStatus: 200},
		{path: "/themes", role: "admin", wantStatus: 200},
		{path: "/kiosk", role: "admin", wantStatus: 200},
		{path: "/library", role: "admin", wantStatus: 200},
		{path: "/training-log", role: "admin", wantStatus: 200},
		{path: "/messages", role: "admin", wantStatus: 200},

		// Coach routes
		{path: "/dashboard", role: "coach", wantStatus: 200},
		{path: "/members", role: "coach", wantStatus: 200},
		{path: "/attendance", role: "coach", wantStatus: 200},
		{path: "/calendar", role: "coach", wantStatus: 200},
		{path: "/library", role: "coach", wantStatus: 200},
		{path: "/training-log", role: "coach", wantStatus: 200},
		{path: "/messages", role: "coach", wantStatus: 200},

		// Member routes
		{path: "/dashboard", role: "member", wantStatus: 200},
		{path: "/attendance", role: "member", wantStatus: 200},
		{path: "/library", role: "member", wantStatus: 200},
		{path: "/training-log", role: "member", wantStatus: 200},
		{path: "/messages", role: "member", wantStatus: 200},
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
		page.WaitForTimeout(500)
	}

	if len(errors) > 0 {
		t.Errorf("console errors found: %v", errors)
	}
}
