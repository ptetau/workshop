package browser_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"

	gradingDomain "workshop/internal/domain/grading"
	memberDomain "workshop/internal/domain/member"
)

func TestMembers_BeltIconsVisibleInListAndProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	page := app.newPage(t)
	app.login(t, page)

	ctx := context.Background()
	memberID := uuid.New().String()
	m := memberDomain.Member{
		ID:        memberID,
		Name:      "Belt Icon Member",
		Email:     "belt-icon-member@test.com",
		Program:   "Adults",
		Status:    "active",
		Fee:       100,
		Frequency: "monthly",
	}
	if err := app.Stores.MemberStore.Save(ctx, m); err != nil {
		t.Fatalf("seed member: %v", err)
	}

	rec := gradingDomain.Record{
		ID:         uuid.New().String(),
		MemberID:   memberID,
		Belt:       "blue",
		Stripe:     2,
		PromotedAt: time.Now().Add(-24 * time.Hour),
		Method:     gradingDomain.MethodStandard,
	}
	if err := app.Stores.GradingRecordStore.Save(ctx, rec); err != nil {
		t.Fatalf("seed grading record: %v", err)
	}

	if _, err := page.Goto(app.BaseURL + "/members"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	// Member row should include belt icon.
	row := page.Locator("tr:has-text('Belt Icon Member')")
	if err := row.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("member row not visible: %v", err)
	}
	icon := row.Locator(".belt-icon.belt-blue")
	if err := icon.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("belt icon not visible in list: %v", err)
	}

	// Navigate to profile.
	if err := page.Locator("a:has-text('Belt Icon Member')").Click(); err != nil {
		t.Fatalf("click member: %v", err)
	}
	if err := page.WaitForLoadState(); err != nil {
		t.Fatalf("wait load: %v", err)
	}

	profileIcon := page.Locator(".belt-icon.belt-blue")
	if err := profileIcon.First().WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		t.Fatalf("belt icon not visible on profile: %v", err)
	}
}
