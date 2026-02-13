package browser_test

import (
	"encoding/json"
	"fmt"
	"testing"
)

// TestCalendar_ViewAndCRUD tests viewing the calendar and creating/deleting events via API.
func TestCalendar_ViewAndCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Create a club event via API
	event1 := apiPost(t, page, app.BaseURL+"/api/calendar/events", map[string]interface{}{
		"title":      "End of Year BBQ",
		"type":       "event",
		"location":   "Workshop HQ",
		"start_date": "2026-03-14",
		"end_date":   "2026-03-14",
	})
	event1ID := event1.(map[string]interface{})["ID"].(string)
	if event1ID == "" {
		t.Fatal("expected event ID")
	}

	// Create a competition via API
	event2 := apiPost(t, page, app.BaseURL+"/api/calendar/events", map[string]interface{}{
		"title":            "Grappling Industries Wellington",
		"type":             "competition",
		"start_date":       "2026-03-15",
		"registration_url": "https://grapplingindustries.com/register",
	})
	event2ID := event2.(map[string]interface{})["ID"].(string)
	if event2ID == "" {
		t.Fatal("expected competition ID")
	}

	// List events for March 2026
	events := apiGet(t, page, app.BaseURL+"/api/calendar/events?from=2026-03-01&to=2026-03-31")
	eventList := events.([]interface{})
	if len(eventList) != 2 {
		t.Fatalf("expected 2 events in March, got %d", len(eventList))
	}

	// Verify event details
	e1 := eventList[0].(map[string]interface{})
	if e1["Title"].(string) != "End of Year BBQ" {
		t.Fatalf("expected title 'End of Year BBQ', got %s", e1["Title"])
	}
	if e1["Type"].(string) != "event" {
		t.Fatalf("expected type 'event', got %s", e1["Type"])
	}

	e2 := eventList[1].(map[string]interface{})
	if e2["Title"].(string) != "Grappling Industries Wellington" {
		t.Fatalf("expected title 'Grappling Industries Wellington', got %s", e2["Title"])
	}
	if e2["Type"].(string) != "competition" {
		t.Fatalf("expected type 'competition', got %s", e2["Type"])
	}

	// Delete event1
	deleteResp, err := page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s/api/calendar/events?id=%s', {method:'DELETE', headers:{'Content-Type':'application/json'}});
		return r.status;
	}`, app.BaseURL, event1ID))
	if err != nil {
		t.Fatal(err)
	}
	if toInt(deleteResp) != 204 {
		t.Fatalf("expected 204 for delete, got %v", deleteResp)
	}

	// Verify only 1 event remains
	remaining := apiGet(t, page, app.BaseURL+"/api/calendar/events?from=2026-03-01&to=2026-03-31")
	remainingList := remaining.([]interface{})
	if len(remainingList) != 1 {
		t.Fatalf("expected 1 event after delete, got %d", len(remainingList))
	}
}

// TestCalendar_PageLoads tests that the calendar page renders for authenticated users.
func TestCalendar_PageLoads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Navigate to calendar page
	_, err := page.Goto(app.BaseURL + "/calendar")
	if err != nil {
		t.Fatal(err)
	}

	// Verify the page title is visible
	title := page.Locator("h1")
	text, err := title.TextContent()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Club Calendar" {
		t.Fatalf("expected 'Club Calendar' heading, got %q", text)
	}

	// Verify month navigation exists
	monthTitle := page.Locator("#monthTitle")
	monthText, err := monthTitle.TextContent()
	if err != nil {
		t.Fatal(err)
	}
	if monthText == "" {
		t.Fatal("expected month title to be populated")
	}
}

// TestCalendar_ValidationErrors tests that the API rejects invalid events.
func TestCalendar_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Missing title should fail
	bodyJSON, _ := json.Marshal(map[string]interface{}{
		"type":       "event",
		"start_date": "2026-04-01",
	})
	status, err := page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s/api/calendar/events', {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: '%s'
		});
		return r.status;
	}`, app.BaseURL, string(bodyJSON)))
	if err != nil {
		t.Fatal(err)
	}
	if toInt(status) != 400 {
		t.Fatalf("expected 400 for missing title, got %v", status)
	}

	// Missing start_date should fail
	bodyJSON2, _ := json.Marshal(map[string]interface{}{
		"title": "Test Event",
		"type":  "event",
	})
	status2, err := page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s/api/calendar/events', {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: '%s'
		});
		return r.status;
	}`, app.BaseURL, string(bodyJSON2)))
	if err != nil {
		t.Fatal(err)
	}
	if toInt(status2) != 400 {
		t.Fatalf("expected 400 for missing start_date, got %v", status2)
	}
}
