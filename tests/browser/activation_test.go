package browser_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"

	accountDomain "workshop/internal/domain/account"
)

// TestActivation_CreateAccountAndActivate tests the full activation flow:
// admin creates a member account → account is pending → member activates via token → can log in.
// Covers #125, #126, #128.
func TestActivation_CreateAccountAndActivate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	ctx := context.Background()

	// Create a pending_activation account directly via store
	acctID := uuid.New().String()
	acct := accountDomain.Account{
		ID:           acctID,
		Email:        "newmember@test.com",
		PasswordHash: "pending_activation",
		Role:         accountDomain.RoleMember,
		Status:       accountDomain.StatusPendingActivation,
		CreatedAt:    time.Now(),
	}
	if err := app.Stores.AccountStore.Save(ctx, acct); err != nil {
		t.Fatalf("failed to seed pending account: %v", err)
	}

	// Create activation token
	tokenStr := uuid.New().String()
	tok := accountDomain.ActivationToken{
		ID:        uuid.New().String(),
		AccountID: acctID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(72 * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := app.Stores.AccountStore.SaveActivationToken(ctx, tok); err != nil {
		t.Fatalf("failed to save activation token: %v", err)
	}

	// Visit the activation page with the token
	page := app.newPage(t)
	_, err := page.Goto(app.BaseURL + "/activate?token=" + tokenStr)
	if err != nil {
		t.Fatalf("failed to navigate to activation page: %v", err)
	}

	// Should see the password form (no error)
	err = page.Locator("#activateForm").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatal("activation form not shown")
	}

	// Fill password and confirm
	if err := page.Locator("#newPassword").Fill("SecurePass123!"); err != nil {
		t.Fatalf("failed to fill password: %v", err)
	}
	if err := page.Locator("#confirmPassword").Fill("SecurePass123!"); err != nil {
		t.Fatalf("failed to fill confirm password: %v", err)
	}

	// Submit
	if err := page.Locator("#activateBtn").Click(); err != nil {
		t.Fatalf("failed to click activate: %v", err)
	}

	// Should show success message
	err = page.Locator("#activateMsg >> text=Account activated").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("activation success message not shown")
	}

	// Should redirect to login after 2 seconds
	if err := page.WaitForURL(app.BaseURL+"/login", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Error("did not redirect to login after activation")
	}

	// Now log in with the new credentials
	if err := page.Locator("input[name=Email]").Fill("newmember@test.com"); err != nil {
		t.Fatalf("failed to fill email: %v", err)
	}
	if err := page.Locator("input[name=Password]").Fill("SecurePass123!"); err != nil {
		t.Fatalf("failed to fill password: %v", err)
	}
	if err := page.Locator("button[type=submit]").Click(); err != nil {
		t.Fatalf("failed to click login: %v", err)
	}
	if err := page.WaitForURL(app.BaseURL+"/dashboard", playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Error("activated member could not log in")
	}
}

// TestActivation_ExpiredToken tests that expired tokens show the correct error.
// Covers #128.
func TestActivation_ExpiredToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	ctx := context.Background()

	// Create a pending account
	acctID := uuid.New().String()
	acct := accountDomain.Account{
		ID:           acctID,
		Email:        "expired@test.com",
		PasswordHash: "pending_activation",
		Role:         accountDomain.RoleMember,
		Status:       accountDomain.StatusPendingActivation,
		CreatedAt:    time.Now(),
	}
	app.Stores.AccountStore.Save(ctx, acct)

	// Create an expired token (expired 1 hour ago)
	tokenStr := uuid.New().String()
	tok := accountDomain.ActivationToken{
		ID:        uuid.New().String(),
		AccountID: acctID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-73 * time.Hour),
	}
	app.Stores.AccountStore.SaveActivationToken(ctx, tok)

	page := app.newPage(t)
	_, err := page.Goto(app.BaseURL + "/activate?token=" + tokenStr)
	if err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Should see the expired error message
	err = page.Locator("text=expired").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("expired token error message not shown")
	}
}

// TestActivation_ResendToken tests admin resending an activation token.
// Covers #127.
func TestActivation_ResendToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	ctx := context.Background()

	// Create a pending account
	acctID := uuid.New().String()
	acct := accountDomain.Account{
		ID:           acctID,
		Email:        "resend@test.com",
		PasswordHash: "pending_activation",
		Role:         accountDomain.RoleMember,
		Status:       accountDomain.StatusPendingActivation,
		CreatedAt:    time.Now(),
	}
	app.Stores.AccountStore.Save(ctx, acct)

	// Create an old (used) token
	oldToken := uuid.New().String()
	tok := accountDomain.ActivationToken{
		ID:        uuid.New().String(),
		AccountID: acctID,
		Token:     oldToken,
		ExpiresAt: time.Now().Add(72 * time.Hour),
		Used:      true,
		CreatedAt: time.Now(),
	}
	app.Stores.AccountStore.SaveActivationToken(ctx, tok)

	// Admin resends activation via API
	page := app.newPage(t)
	app.login(t, page)

	// Call the resend API directly via JS
	result, err := page.Evaluate(`async () => {
		const resp = await fetch('/api/admin/resend-activation', {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify({AccountID: '` + acctID + `'})
		});
		return {status: resp.status, body: await resp.json()};
	}`)
	if err != nil {
		t.Fatalf("failed to call resend API: %v", err)
	}

	resultMap := result.(map[string]interface{})
	statusVal := resultMap["status"]
	var statusCode int
	switch v := statusVal.(type) {
	case float64:
		statusCode = int(v)
	case int:
		statusCode = v
	default:
		t.Fatalf("unexpected status type %T", statusVal)
	}
	if statusCode != 200 {
		t.Errorf("resend API status = %d, want 200", statusCode)
	}
	body := resultMap["body"].(map[string]interface{})
	if body["status"] != "sent" {
		t.Errorf("resend response status = %v, want 'sent'", body["status"])
	}
	newToken, ok := body["token"].(string)
	if !ok || newToken == "" {
		t.Error("resend did not return a new token")
	}

	// Old token should no longer work (used tokens are rejected)
	page2 := app.newPage(t)
	_, _ = page2.Goto(app.BaseURL + "/activate?token=" + oldToken)
	err = page2.Locator("text=already been used").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("old token should show 'already been used' error")
	}

	// New token should work — shows the form
	page3 := app.newPage(t)
	_, _ = page3.Goto(app.BaseURL + "/activate?token=" + newToken)
	err = page3.Locator("#activateForm").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("new token should show activation form")
	}
}

// TestActivation_PendingAccountCannotLogin tests that pending accounts are blocked from login.
// Covers #125 acceptance criteria.
func TestActivation_PendingAccountCannotLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	app := newTestApp(t)
	ctx := context.Background()

	// Create a pending account with a known password hash
	acctID := uuid.New().String()
	acct := accountDomain.Account{
		ID:           acctID,
		Email:        "pending@test.com",
		PasswordHash: "pending_activation",
		Role:         accountDomain.RoleMember,
		Status:       accountDomain.StatusPendingActivation,
		CreatedAt:    time.Now(),
	}
	app.Stores.AccountStore.Save(ctx, acct)

	page := app.newPage(t)
	_, _ = page.Goto(app.BaseURL + "/login")

	if err := page.Locator("input[name=Email]").Fill("pending@test.com"); err != nil {
		t.Fatalf("failed to fill email: %v", err)
	}
	if err := page.Locator("input[name=Password]").Fill("anypassword12"); err != nil {
		t.Fatalf("failed to fill password: %v", err)
	}
	if err := page.Locator("button[type=submit]").Click(); err != nil {
		t.Fatalf("failed to click login: %v", err)
	}

	// Should see activation pending error
	err := page.Locator("text=pending activation").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Error("pending activation error not shown on login attempt")
	}
}
