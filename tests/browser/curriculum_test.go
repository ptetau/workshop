package browser_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// TestCurriculum_RotorCRUD tests creating, listing, activating, and deleting rotors via API.
func TestCurriculum_RotorCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)

	// Login as admin
	app.login(t, page)

	// Get a class type to use
	classTypes := apiGet(t, page, app.BaseURL+"/api/class-types")
	ctList, ok := classTypes.([]interface{})
	if !ok || len(ctList) == 0 {
		t.Fatal("no class types found")
	}
	ct := ctList[0].(map[string]interface{})
	classTypeID := ct["ID"].(string)

	// Create a rotor
	createResp := apiPost(t, page, app.BaseURL+"/api/rotors", map[string]interface{}{
		"class_type_id": classTypeID,
		"name":          "Test Rotor v1",
	})
	rotorID := createResp.(map[string]interface{})["ID"].(string)
	if rotorID == "" {
		t.Fatal("expected rotor ID")
	}

	// List rotors for class
	listResp := apiGet(t, page, app.BaseURL+"/api/rotors?class_type_id="+classTypeID)
	rotors := listResp.([]interface{})
	if len(rotors) != 1 {
		t.Fatalf("expected 1 rotor, got %d", len(rotors))
	}
	if rotors[0].(map[string]interface{})["Status"].(string) != "draft" {
		t.Fatal("expected draft status")
	}

	// Activate the rotor
	apiPost(t, page, app.BaseURL+"/api/rotors/activate", map[string]interface{}{
		"id": rotorID,
	})

	// Verify it's active
	getResp := apiGet(t, page, app.BaseURL+"/api/rotors/by-id?id="+rotorID)
	if getResp.(map[string]interface{})["Status"].(string) != "active" {
		t.Fatal("expected active status after activation")
	}

	// Create a second rotor (draft)
	create2 := apiPost(t, page, app.BaseURL+"/api/rotors", map[string]interface{}{
		"class_type_id": classTypeID,
		"name":          "Test Rotor v2",
	})
	rotor2ID := create2.(map[string]interface{})["ID"].(string)

	// Delete draft rotor (should succeed)
	deleteResp, err := page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s/api/rotors/by-id?id=%s', {method:'DELETE', headers:{'Content-Type':'application/json'}});
		return r.status;
	}`, app.BaseURL, rotor2ID))
	if err != nil {
		t.Fatal(err)
	}
	statusCode := toInt(deleteResp)
	if statusCode != 204 {
		t.Fatalf("expected 204 for delete, got %d", statusCode)
	}

	// Try to delete active rotor (should fail)
	deleteActiveResp, err := page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s/api/rotors/by-id?id=%s', {method:'DELETE', headers:{'Content-Type':'application/json'}});
		return r.status;
	}`, app.BaseURL, rotorID))
	if err != nil {
		t.Fatal(err)
	}
	if toInt(deleteActiveResp) != 400 {
		t.Fatalf("expected 400 for deleting active rotor, got %v", deleteActiveResp)
	}
}

// TestCurriculum_ThemesAndTopics tests adding themes and topics to a rotor.
func TestCurriculum_ThemesAndTopics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Setup: get class type and create a draft rotor
	classTypes := apiGet(t, page, app.BaseURL+"/api/class-types")
	ctList := classTypes.([]interface{})
	classTypeID := ctList[0].(map[string]interface{})["ID"].(string)

	createResp := apiPost(t, page, app.BaseURL+"/api/rotors", map[string]interface{}{
		"class_type_id": classTypeID,
		"name":          "Theme Test Rotor",
	})
	rotorID := createResp.(map[string]interface{})["ID"].(string)

	// Add themes
	theme1 := apiPost(t, page, app.BaseURL+"/api/rotors/themes", map[string]interface{}{
		"rotor_id": rotorID, "name": "Standing", "position": 0,
	})
	theme1ID := theme1.(map[string]interface{})["ID"].(string)

	theme2 := apiPost(t, page, app.BaseURL+"/api/rotors/themes", map[string]interface{}{
		"rotor_id": rotorID, "name": "Guard", "position": 1,
	})
	theme2ID := theme2.(map[string]interface{})["ID"].(string)

	// List themes
	themes := apiGet(t, page, app.BaseURL+"/api/rotors/themes?rotor_id="+rotorID)
	themeList := themes.([]interface{})
	if len(themeList) != 2 {
		t.Fatalf("expected 2 themes, got %d", len(themeList))
	}

	// Add topics to theme1
	topic1 := apiPost(t, page, app.BaseURL+"/api/rotors/topics", map[string]interface{}{
		"rotor_theme_id": theme1ID, "name": "Single Leg", "duration_weeks": 1,
	})
	topic1ID := topic1.(map[string]interface{})["ID"].(string)

	apiPost(t, page, app.BaseURL+"/api/rotors/topics", map[string]interface{}{
		"rotor_theme_id": theme1ID, "name": "Double Leg", "duration_weeks": 2,
	})

	// Add topic to theme2
	apiPost(t, page, app.BaseURL+"/api/rotors/topics", map[string]interface{}{
		"rotor_theme_id": theme2ID, "name": "Closed Guard Attacks", "duration_weeks": 1,
	})

	// List topics for theme1
	topics := apiGet(t, page, app.BaseURL+"/api/rotors/topics?theme_id="+theme1ID)
	topicList := topics.([]interface{})
	if len(topicList) != 2 {
		t.Fatalf("expected 2 topics in theme1, got %d", len(topicList))
	}

	// Delete a topic
	_, err := page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s/api/rotors/topics?id=%s', {method:'DELETE', headers:{'Content-Type':'application/json'}});
		return r.status;
	}`, app.BaseURL, topic1ID))
	if err != nil {
		t.Fatal(err)
	}

	// Verify deletion
	topicsAfter := apiGet(t, page, app.BaseURL+"/api/rotors/topics?theme_id="+theme1ID)
	if len(topicsAfter.([]interface{})) != 1 {
		t.Fatal("expected 1 topic after deletion")
	}

	// Delete a theme (cascades topics)
	_, err = page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s/api/rotors/themes?id=%s', {method:'DELETE', headers:{'Content-Type':'application/json'}});
		return r.status;
	}`, app.BaseURL, theme2ID))
	if err != nil {
		t.Fatal(err)
	}

	themesAfter := apiGet(t, page, app.BaseURL+"/api/rotors/themes?rotor_id="+rotorID)
	if len(themesAfter.([]interface{})) != 1 {
		t.Fatal("expected 1 theme after deletion")
	}
}

// TestCurriculum_ScheduleAndAdvancement tests topic activation, extension, skipping, and completion.
func TestCurriculum_ScheduleAndAdvancement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Setup: class type, active rotor, theme, topics
	classTypes := apiGet(t, page, app.BaseURL+"/api/class-types")
	classTypeID := classTypes.([]interface{})[0].(map[string]interface{})["ID"].(string)

	rotorResp := apiPost(t, page, app.BaseURL+"/api/rotors", map[string]interface{}{
		"class_type_id": classTypeID, "name": "Schedule Test Rotor",
	})
	rotorID := rotorResp.(map[string]interface{})["ID"].(string)

	themeResp := apiPost(t, page, app.BaseURL+"/api/rotors/themes", map[string]interface{}{
		"rotor_id": rotorID, "name": "Standing", "position": 0,
	})
	themeID := themeResp.(map[string]interface{})["ID"].(string)

	topic1 := apiPost(t, page, app.BaseURL+"/api/rotors/topics", map[string]interface{}{
		"rotor_theme_id": themeID, "name": "Single Leg", "duration_weeks": 1,
	})
	topic1ID := topic1.(map[string]interface{})["ID"].(string)

	topic2 := apiPost(t, page, app.BaseURL+"/api/rotors/topics", map[string]interface{}{
		"rotor_theme_id": themeID, "name": "Double Leg", "duration_weeks": 1,
	})
	topic2ID := topic2.(map[string]interface{})["ID"].(string)

	// Activate rotor
	apiPost(t, page, app.BaseURL+"/api/rotors/activate", map[string]interface{}{"id": rotorID})

	// Activate topic1
	schedResp := apiPost(t, page, app.BaseURL+"/api/rotors/schedule/action", map[string]interface{}{
		"action": "activate", "topic_id": topic1ID, "rotor_theme_id": themeID,
	})
	if schedResp.(map[string]interface{})["Status"].(string) != "active" {
		t.Fatal("expected active schedule")
	}

	// Extend the current topic
	extendResp := apiPost(t, page, app.BaseURL+"/api/rotors/schedule/action", map[string]interface{}{
		"action": "extend", "rotor_theme_id": themeID, "extend_weeks": 1,
	})
	if extendResp.(map[string]interface{})["Status"].(string) != "active" {
		t.Fatal("expected still active after extend")
	}

	// Skip the current topic
	skipResp := apiPost(t, page, app.BaseURL+"/api/rotors/schedule/action", map[string]interface{}{
		"action": "skip", "rotor_theme_id": themeID,
	})
	if skipResp.(map[string]interface{})["Status"].(string) != "skipped" {
		t.Fatal("expected skipped status")
	}

	// Activate topic2
	sched2 := apiPost(t, page, app.BaseURL+"/api/rotors/schedule/action", map[string]interface{}{
		"action": "activate", "topic_id": topic2ID, "rotor_theme_id": themeID,
	})
	if sched2.(map[string]interface{})["Status"].(string) != "active" {
		t.Fatal("expected active schedule for topic2")
	}

	// Complete topic2 — should auto-advance to topic1 (cycling)
	completeResp := apiPost(t, page, app.BaseURL+"/api/rotors/schedule/action", map[string]interface{}{
		"action": "complete", "rotor_theme_id": themeID,
	})
	completeMap := completeResp.(map[string]interface{})
	completed := completeMap["completed"].(map[string]interface{})
	if completed["Status"].(string) != "completed" {
		t.Fatal("expected completed status")
	}
	// Auto-advance should have started the next topic (topic1, wrapping around)
	if completeMap["next_started"] == nil {
		t.Fatal("expected next_started from auto-advance")
	}
	nextStarted := completeMap["next_started"].(map[string]interface{})
	if nextStarted["TopicID"].(string) != topic1ID {
		t.Fatalf("expected auto-advance to topic1 (wrap-around), got topic %s", nextStarted["TopicID"])
	}
	if nextStarted["Status"].(string) != "active" {
		t.Fatal("expected next_started to be active")
	}

	// Verify curriculum view returns correct state
	viewResp := apiGet(t, page, app.BaseURL+"/api/curriculum/view?class_type_id="+classTypeID)
	viewMap := viewResp.(map[string]interface{})
	if viewMap["rotor"] == nil {
		t.Fatal("expected rotor in curriculum view")
	}
	viewThemes := viewMap["themes"].([]interface{})
	if len(viewThemes) != 1 {
		t.Fatalf("expected 1 theme in view, got %d", len(viewThemes))
	}
}

// TestCurriculum_Voting tests casting votes and the bump mechanism.
func TestCurriculum_Voting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	// Setup: active rotor with theme and topics
	classTypes := apiGet(t, page, app.BaseURL+"/api/class-types")
	classTypeID := classTypes.([]interface{})[0].(map[string]interface{})["ID"].(string)

	rotorResp := apiPost(t, page, app.BaseURL+"/api/rotors", map[string]interface{}{
		"class_type_id": classTypeID, "name": "Vote Test Rotor",
	})
	rotorID := rotorResp.(map[string]interface{})["ID"].(string)

	themeResp := apiPost(t, page, app.BaseURL+"/api/rotors/themes", map[string]interface{}{
		"rotor_id": rotorID, "name": "Guard", "position": 0,
	})
	themeID := themeResp.(map[string]interface{})["ID"].(string)

	topic1 := apiPost(t, page, app.BaseURL+"/api/rotors/topics", map[string]interface{}{
		"rotor_theme_id": themeID, "name": "DLR Sweeps", "duration_weeks": 1,
	})
	topic1ID := topic1.(map[string]interface{})["ID"].(string)

	topic2 := apiPost(t, page, app.BaseURL+"/api/rotors/topics", map[string]interface{}{
		"rotor_theme_id": themeID, "name": "Closed Guard Attacks", "duration_weeks": 1,
	})
	topic2ID := topic2.(map[string]interface{})["ID"].(string)

	apiPost(t, page, app.BaseURL+"/api/rotors/activate", map[string]interface{}{"id": rotorID})

	// Activate topic1
	apiPost(t, page, app.BaseURL+"/api/rotors/schedule/action", map[string]interface{}{
		"action": "activate", "topic_id": topic1ID, "rotor_theme_id": themeID,
	})

	// Vote for topic2
	voteResp := apiPost(t, page, app.BaseURL+"/api/votes", map[string]interface{}{
		"topic_id": topic2ID,
	})
	voteMap := voteResp.(map[string]interface{})
	if voteMap["status"].(string) != "voted" {
		t.Fatal("expected voted status")
	}
	votes := toInt(voteMap["votes"])
	if votes != 1 {
		t.Fatalf("expected 1 vote, got %d", votes)
	}

	// Verify vote count via GET
	countResp := apiGet(t, page, app.BaseURL+"/api/votes?topic_id="+topic2ID)
	if toInt(countResp.(map[string]interface{})["votes"]) != 1 {
		t.Fatal("expected 1 vote via GET")
	}

	// Duplicate vote should fail (409)
	dupStatus, err := page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s/api/votes', {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify({topic_id: '%s'})
		});
		return r.status;
	}`, app.BaseURL, topic2ID))
	if err != nil {
		t.Fatal(err)
	}
	if toInt(dupStatus) != 409 {
		t.Fatalf("expected 409 for duplicate vote, got %v", dupStatus)
	}

	// Bump topic2 (replaces current active topic1)
	bumpResp := apiPost(t, page, app.BaseURL+"/api/rotors/topics/bump", map[string]interface{}{
		"topic_id": topic2ID, "rotor_theme_id": themeID,
	})
	if bumpResp.(map[string]interface{})["Status"].(string) != "active" {
		t.Fatal("expected bumped topic to be active")
	}

	// Votes should be cleared after bump
	countAfter := apiGet(t, page, app.BaseURL+"/api/votes?topic_id="+topic2ID)
	if toInt(countAfter.(map[string]interface{})["votes"]) != 0 {
		t.Fatal("expected 0 votes after bump")
	}
}

// TestCurriculum_RotorVersioning tests creating a new version that archives the old one.
func TestCurriculum_RotorVersioning(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	classTypes := apiGet(t, page, app.BaseURL+"/api/class-types")
	classTypeID := classTypes.([]interface{})[0].(map[string]interface{})["ID"].(string)

	// Create and activate v1
	v1 := apiPost(t, page, app.BaseURL+"/api/rotors", map[string]interface{}{
		"class_type_id": classTypeID, "name": "Versioned Rotor",
	})
	v1ID := v1.(map[string]interface{})["ID"].(string)
	apiPost(t, page, app.BaseURL+"/api/rotors/activate", map[string]interface{}{"id": v1ID})

	// Create and activate v2 — should archive v1
	v2 := apiPost(t, page, app.BaseURL+"/api/rotors", map[string]interface{}{
		"class_type_id": classTypeID, "name": "Versioned Rotor v2",
	})
	v2ID := v2.(map[string]interface{})["ID"].(string)
	apiPost(t, page, app.BaseURL+"/api/rotors/activate", map[string]interface{}{"id": v2ID})

	// Verify v1 is archived
	v1After := apiGet(t, page, app.BaseURL+"/api/rotors/by-id?id="+v1ID)
	if v1After.(map[string]interface{})["Status"].(string) != "archived" {
		t.Fatal("expected v1 to be archived after v2 activation")
	}

	// Verify v2 is active
	v2After := apiGet(t, page, app.BaseURL+"/api/rotors/by-id?id="+v2ID)
	if v2After.(map[string]interface{})["Status"].(string) != "active" {
		t.Fatal("expected v2 to be active")
	}

	// List all rotors — should show both
	listResp := apiGet(t, page, app.BaseURL+"/api/rotors?class_type_id="+classTypeID)
	rotorList := listResp.([]interface{})
	if len(rotorList) != 2 {
		t.Fatalf("expected 2 rotors in history, got %d", len(rotorList))
	}
}

// TestCurriculum_PreviewToggle tests toggling the member preview setting.
func TestCurriculum_PreviewToggle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	classTypes := apiGet(t, page, app.BaseURL+"/api/class-types")
	classTypeID := classTypes.([]interface{})[0].(map[string]interface{})["ID"].(string)

	rotorResp := apiPost(t, page, app.BaseURL+"/api/rotors", map[string]interface{}{
		"class_type_id": classTypeID, "name": "Preview Rotor",
	})
	rotorID := rotorResp.(map[string]interface{})["ID"].(string)
	apiPost(t, page, app.BaseURL+"/api/rotors/activate", map[string]interface{}{"id": rotorID})

	// Toggle preview on
	onResp := apiPost(t, page, app.BaseURL+"/api/rotors/preview", map[string]interface{}{
		"id": rotorID, "preview_on": true,
	})
	if onResp.(map[string]interface{})["PreviewOn"] != true {
		t.Fatal("expected preview_on to be true")
	}

	// Toggle preview off
	offResp := apiPost(t, page, app.BaseURL+"/api/rotors/preview", map[string]interface{}{
		"id": rotorID, "preview_on": false,
	})
	if offResp.(map[string]interface{})["PreviewOn"] != false {
		t.Fatal("expected preview_on to be false")
	}
}

// TestCurriculum_TopicReorder tests reordering topics within a theme.
func TestCurriculum_TopicReorder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	classTypes := apiGet(t, page, app.BaseURL+"/api/class-types")
	classTypeID := classTypes.([]interface{})[0].(map[string]interface{})["ID"].(string)

	rotorResp := apiPost(t, page, app.BaseURL+"/api/rotors", map[string]interface{}{
		"class_type_id": classTypeID, "name": "Reorder Rotor",
	})
	rotorID := rotorResp.(map[string]interface{})["ID"].(string)

	themeResp := apiPost(t, page, app.BaseURL+"/api/rotors/themes", map[string]interface{}{
		"rotor_id": rotorID, "name": "Reorder Theme", "position": 0,
	})
	themeID := themeResp.(map[string]interface{})["ID"].(string)

	t1 := apiPost(t, page, app.BaseURL+"/api/rotors/topics", map[string]interface{}{
		"rotor_theme_id": themeID, "name": "Topic A", "position": 0,
	})
	t1ID := t1.(map[string]interface{})["ID"].(string)

	t2 := apiPost(t, page, app.BaseURL+"/api/rotors/topics", map[string]interface{}{
		"rotor_theme_id": themeID, "name": "Topic B", "position": 1,
	})
	t2ID := t2.(map[string]interface{})["ID"].(string)

	t3 := apiPost(t, page, app.BaseURL+"/api/rotors/topics", map[string]interface{}{
		"rotor_theme_id": themeID, "name": "Topic C", "position": 2,
	})
	t3ID := t3.(map[string]interface{})["ID"].(string)

	// Reorder: C, A, B
	reorderStatus, err := page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s/api/rotors/topics/reorder', {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify({rotor_theme_id: '%s', topic_ids: ['%s','%s','%s']})
		});
		return r.status;
	}`, app.BaseURL, themeID, t3ID, t1ID, t2ID))
	if err != nil {
		t.Fatal(err)
	}
	if toInt(reorderStatus) != 204 {
		t.Fatalf("expected 204, got %v", reorderStatus)
	}

	// Verify new order
	topics := apiGet(t, page, app.BaseURL+"/api/rotors/topics?theme_id="+themeID)
	topicList := topics.([]interface{})
	if topicList[0].(map[string]interface{})["Name"].(string) != "Topic C" {
		t.Fatal("expected Topic C first after reorder")
	}
	if topicList[1].(map[string]interface{})["Name"].(string) != "Topic A" {
		t.Fatal("expected Topic A second after reorder")
	}
	if topicList[2].(map[string]interface{})["Name"].(string) != "Topic B" {
		t.Fatal("expected Topic B third after reorder")
	}
}

// --- Helper functions ---

// apiGet performs a GET request via page.Evaluate and returns parsed JSON.
func apiGet(t *testing.T, page playwright.Page, url string) interface{} {
	t.Helper()
	result, err := page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s');
		if (!r.ok) throw new Error('GET failed: ' + r.status);
		return r.json();
	}`, url))
	if err != nil {
		t.Fatalf("apiGet %s: %v", url, err)
	}
	return result
}

// apiPost performs a POST request via page.Evaluate and returns parsed JSON.
func apiPost(t *testing.T, page playwright.Page, url string, body map[string]interface{}) interface{} {
	t.Helper()
	bodyJSON, _ := json.Marshal(body)
	result, err := page.Evaluate(fmt.Sprintf(`async () => {
		const r = await fetch('%s', {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: '%s'
		});
		if (!r.ok && r.status !== 201) throw new Error('POST failed: ' + r.status);
		return r.json();
	}`, url, string(bodyJSON)))
	if err != nil {
		t.Fatalf("apiPost %s: %v", url, err)
	}
	return result
}

// toInt converts a page.Evaluate result to int (handles float64 and int).
func toInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}
