package orchestrators

import (
	"context"
	"errors"
	"testing"
	"time"

	emailAdapter "workshop/internal/adapters/email"
	emailDomain "workshop/internal/domain/email"
)

// --- Mock email store ---

type mockEmailStore struct {
	emails     map[string]emailDomain.Email
	recipients map[string][]emailDomain.Recipient
}

func newMockEmailStore() *mockEmailStore {
	return &mockEmailStore{
		emails:     make(map[string]emailDomain.Email),
		recipients: make(map[string][]emailDomain.Recipient),
	}
}

// GetByID retrieves a mock email by ID.
// PRE: id is non-empty
// POST: Returns mock email or error
func (m *mockEmailStore) GetByID(_ context.Context, id string) (emailDomain.Email, error) {
	e, ok := m.emails[id]
	if !ok {
		return emailDomain.Email{}, errors.New("not found")
	}
	return e, nil
}

// Save persists a mock email.
// PRE: e has a valid ID
// POST: Email stored in map
func (m *mockEmailStore) Save(_ context.Context, e emailDomain.Email) error {
	m.emails[e.ID] = e
	return nil
}

// SaveRecipients saves mock recipients.
// PRE: emailID is non-empty
// POST: Recipients stored in map
func (m *mockEmailStore) SaveRecipients(_ context.Context, emailID string, recipients []emailDomain.Recipient) error {
	m.recipients[emailID] = recipients
	return nil
}

// GetRecipients retrieves mock recipients.
// PRE: emailID is non-empty
// POST: Returns stored recipients
func (m *mockEmailStore) GetRecipients(_ context.Context, emailID string) ([]emailDomain.Recipient, error) {
	return m.recipients[emailID], nil
}

// --- Mock member lookup ---

type mockMemberLookup struct {
	members map[string]struct{ name, email string }
}

func newMockMemberLookup() *mockMemberLookup {
	return &mockMemberLookup{
		members: map[string]struct{ name, email string }{
			"member-1": {name: "Marcus Almeida", email: "marcus@email.com"},
			"member-2": {name: "Yuki Nakai", email: "yuki@email.com"},
			"member-3": {name: "Roger Gracie", email: "roger@email.com"},
		},
	}
}

// GetEmailByMemberID returns mock member name and email.
// PRE: memberID is non-empty
// POST: Returns name and email or error
func (m *mockMemberLookup) GetEmailByMemberID(_ context.Context, memberID string) (string, string, error) {
	member, ok := m.members[memberID]
	if !ok {
		return "", "", errors.New("member not found")
	}
	return member.name, member.email, nil
}

// --- Mock email sender ---

type mockEmailSender struct {
	sent   int
	failAt int // fail on the Nth send (-1 = never fail)
}

func newMockEmailSender() *mockEmailSender {
	return &mockEmailSender{failAt: -1}
}

// Send simulates sending an email.
// PRE: req is valid
// POST: Increments sent counter
func (m *mockEmailSender) Send(_ context.Context, req emailAdapter.SendRequest) (emailAdapter.SendResult, error) {
	m.sent++
	if m.failAt >= 0 && m.sent >= m.failAt {
		return emailAdapter.SendResult{}, errors.New("send failed")
	}
	return emailAdapter.SendResult{MessageID: "mock-msg-id", SentAt: emailFixedTime}, nil
}

// SendBatch simulates batch sending.
// PRE: reqs is non-empty
// POST: Returns results for each request
func (m *mockEmailSender) SendBatch(_ context.Context, reqs []emailAdapter.SendRequest) ([]emailAdapter.SendResult, error) {
	if m.failAt >= 0 {
		return nil, errors.New("batch send failed")
	}
	var results []emailAdapter.SendResult
	for range reqs {
		m.sent++
		results = append(results, emailAdapter.SendResult{MessageID: "mock-batch-id", SentAt: emailFixedTime})
	}
	return results, nil
}

var emailFixedTime = time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
var idCounter int

func testGenerateID() string {
	idCounter++
	return "test-email-" + string(rune('0'+idCounter))
}

func testNow() time.Time {
	return emailFixedTime
}

// --- Compose Tests ---

// TestComposeEmail_NewDraft tests creating a new email draft with recipients.
func TestComposeEmail_NewDraft(t *testing.T) {
	store := newMockEmailStore()
	lookup := newMockMemberLookup()

	input := ComposeEmailInput{
		Subject:   "Schedule Change",
		Body:      "Monday class moves to Tuesday.",
		SenderID:  "admin-1",
		MemberIDs: []string{"member-1", "member-2"},
	}
	deps := ComposeEmailDeps{
		EmailStore:   store,
		MemberLookup: lookup,
		GenerateID:   func() string { return "draft-1" },
		Now:          testNow,
	}

	em, err := ExecuteComposeEmail(context.Background(), input, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if em.ID != "draft-1" {
		t.Errorf("ID = %q, want %q", em.ID, "draft-1")
	}
	if em.Status != emailDomain.StatusDraft {
		t.Errorf("status = %q, want %q", em.Status, emailDomain.StatusDraft)
	}
	if em.Subject != "Schedule Change" {
		t.Errorf("subject = %q, want %q", em.Subject, "Schedule Change")
	}

	// Check recipients were saved
	recs := store.recipients["draft-1"]
	if len(recs) != 2 {
		t.Fatalf("recipients = %d, want 2", len(recs))
	}
	if recs[0].MemberName != "Marcus Almeida" {
		t.Errorf("recipient[0].MemberName = %q, want %q", recs[0].MemberName, "Marcus Almeida")
	}
	if recs[1].MemberEmail != "yuki@email.com" {
		t.Errorf("recipient[1].MemberEmail = %q, want %q", recs[1].MemberEmail, "yuki@email.com")
	}
}

// TestComposeEmail_UpdateDraft tests updating an existing draft.
func TestComposeEmail_UpdateDraft(t *testing.T) {
	store := newMockEmailStore()
	lookup := newMockMemberLookup()

	// Seed an existing draft
	store.emails["draft-1"] = emailDomain.Email{
		ID:        "draft-1",
		Subject:   "Old Subject",
		Body:      "Old body",
		SenderID:  "admin-1",
		Status:    emailDomain.StatusDraft,
		CreatedAt: fixedTime.Add(-time.Hour),
	}

	input := ComposeEmailInput{
		EmailID:   "draft-1",
		Subject:   "New Subject",
		Body:      "New body",
		SenderID:  "admin-1",
		MemberIDs: []string{"member-3"},
	}
	deps := ComposeEmailDeps{
		EmailStore:   store,
		MemberLookup: lookup,
		GenerateID:   func() string { return "should-not-use" },
		Now:          testNow,
	}

	em, err := ExecuteComposeEmail(context.Background(), input, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if em.ID != "draft-1" {
		t.Errorf("ID = %q, want %q (should reuse existing)", em.ID, "draft-1")
	}
	if em.Subject != "New Subject" {
		t.Errorf("subject = %q, want %q", em.Subject, "New Subject")
	}
}

// TestComposeEmail_CannotUpdateSentEmail tests that sent emails cannot be updated.
func TestComposeEmail_CannotUpdateSentEmail(t *testing.T) {
	store := newMockEmailStore()
	store.emails["sent-1"] = emailDomain.Email{
		ID:        "sent-1",
		Subject:   "Already Sent",
		Body:      "body",
		SenderID:  "admin-1",
		Status:    emailDomain.StatusSent,
		CreatedAt: fixedTime,
	}

	input := ComposeEmailInput{
		EmailID:  "sent-1",
		Subject:  "Update attempt",
		Body:     "new body",
		SenderID: "admin-1",
	}
	deps := ComposeEmailDeps{
		EmailStore:   store,
		MemberLookup: newMockMemberLookup(),
		GenerateID:   func() string { return "x" },
		Now:          testNow,
	}

	_, err := ExecuteComposeEmail(context.Background(), input, deps)
	if err != emailDomain.ErrNotDraft {
		t.Errorf("expected ErrNotDraft, got: %v", err)
	}
}

// TestComposeEmail_MissingSender tests that missing sender is rejected.
func TestComposeEmail_MissingSender(t *testing.T) {
	input := ComposeEmailInput{Subject: "s", Body: "b"}
	deps := ComposeEmailDeps{
		EmailStore:   newMockEmailStore(),
		MemberLookup: newMockMemberLookup(),
		GenerateID:   func() string { return "x" },
		Now:          testNow,
	}

	_, err := ExecuteComposeEmail(context.Background(), input, deps)
	if err == nil {
		t.Error("expected error for missing sender")
	}
}

// --- Send Tests ---

// TestSendEmail_Success tests sending a draft email to recipients.
func TestSendEmail_Success(t *testing.T) {
	store := newMockEmailStore()
	sender := newMockEmailSender()

	// Seed a draft with recipients
	store.emails["draft-1"] = emailDomain.Email{
		ID:        "draft-1",
		Subject:   "Grading Day",
		Body:      "<p>Grading is on Saturday!</p>",
		SenderID:  "admin-1",
		Status:    emailDomain.StatusDraft,
		CreatedAt: fixedTime,
	}
	store.recipients["draft-1"] = []emailDomain.Recipient{
		{EmailID: "draft-1", MemberID: "member-1", MemberName: "Marcus", MemberEmail: "marcus@email.com"},
		{EmailID: "draft-1", MemberID: "member-2", MemberName: "Yuki", MemberEmail: "yuki@email.com"},
	}

	input := SendEmailInput{EmailID: "draft-1", SenderID: "admin-1"}
	deps := SendEmailDeps{
		EmailStore:  store,
		EmailSender: sender,
		Now:         testNow,
		FromAddress: "Workshop <noreply@test.com>",
		ReplyTo:     "info@test.com",
	}

	em, err := ExecuteSendEmail(context.Background(), input, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if em.Status != emailDomain.StatusSent {
		t.Errorf("status = %q, want %q", em.Status, emailDomain.StatusSent)
	}
	if sender.sent != 2 {
		t.Errorf("sent count = %d, want 2", sender.sent)
	}
	if em.ResendMessageID == "" {
		t.Error("expected ResendMessageID to be set")
	}
}

// TestSendEmail_NoRecipients tests that emails without recipients cannot be sent.
func TestSendEmail_NoRecipients(t *testing.T) {
	store := newMockEmailStore()
	store.emails["draft-1"] = emailDomain.Email{
		ID:        "draft-1",
		Subject:   "Test",
		Body:      "body",
		SenderID:  "admin-1",
		Status:    emailDomain.StatusDraft,
		CreatedAt: fixedTime,
	}
	// No recipients saved

	input := SendEmailInput{EmailID: "draft-1"}
	deps := SendEmailDeps{
		EmailStore:  store,
		EmailSender: newMockEmailSender(),
		Now:         testNow,
	}

	_, err := ExecuteSendEmail(context.Background(), input, deps)
	if err != emailDomain.ErrNoRecipients {
		t.Errorf("expected ErrNoRecipients, got: %v", err)
	}
}

// TestSendEmail_AlreadySent tests that already-sent emails cannot be re-sent.
func TestSendEmail_AlreadySent(t *testing.T) {
	store := newMockEmailStore()
	store.emails["sent-1"] = emailDomain.Email{
		ID:        "sent-1",
		Subject:   "Already Sent",
		Body:      "body",
		SenderID:  "admin-1",
		Status:    emailDomain.StatusSent,
		CreatedAt: fixedTime,
	}

	input := SendEmailInput{EmailID: "sent-1"}
	deps := SendEmailDeps{
		EmailStore:  store,
		EmailSender: newMockEmailSender(),
		Now:         testNow,
	}

	_, err := ExecuteSendEmail(context.Background(), input, deps)
	if err != emailDomain.ErrNotDraft {
		t.Errorf("expected ErrNotDraft, got: %v", err)
	}
}

// TestSendEmail_ProviderFailure tests that provider failures mark the email as failed.
func TestSendEmail_ProviderFailure(t *testing.T) {
	store := newMockEmailStore()
	sender := newMockEmailSender()
	sender.failAt = 0 // fail immediately

	store.emails["draft-1"] = emailDomain.Email{
		ID:        "draft-1",
		Subject:   "Test",
		Body:      "body",
		SenderID:  "admin-1",
		Status:    emailDomain.StatusDraft,
		CreatedAt: fixedTime,
	}
	store.recipients["draft-1"] = []emailDomain.Recipient{
		{EmailID: "draft-1", MemberID: "member-1", MemberName: "Marcus", MemberEmail: "marcus@email.com"},
	}

	input := SendEmailInput{EmailID: "draft-1"}
	deps := SendEmailDeps{
		EmailStore:  store,
		EmailSender: sender,
		Now:         testNow,
	}

	em, err := ExecuteSendEmail(context.Background(), input, deps)
	if err == nil {
		t.Error("expected error from provider failure")
	}
	// Email should be marked as failed in store
	saved := store.emails["draft-1"]
	if saved.Status != emailDomain.StatusFailed && em.Status != emailDomain.StatusFailed {
		t.Errorf("expected failed status, got stored=%q returned=%q", saved.Status, em.Status)
	}
}

// TestSendEmail_MissingEmailID tests that missing email ID is rejected.
func TestSendEmail_MissingEmailID(t *testing.T) {
	input := SendEmailInput{}
	deps := SendEmailDeps{
		EmailStore:  newMockEmailStore(),
		EmailSender: newMockEmailSender(),
		Now:         testNow,
	}

	_, err := ExecuteSendEmail(context.Background(), input, deps)
	if err == nil {
		t.Error("expected error for missing email ID")
	}
}
