package orchestrators

import (
	"context"
	"errors"
	"log/slog"
	"time"

	emailAdapter "workshop/internal/adapters/email"
	emailDomain "workshop/internal/domain/email"
)

// EmailStoreForOrchestrator defines the store interface needed by email orchestrators.
type EmailStoreForOrchestrator interface {
	GetByID(ctx context.Context, id string) (emailDomain.Email, error)
	Save(ctx context.Context, e emailDomain.Email) error
	SaveRecipients(ctx context.Context, emailID string, recipients []emailDomain.Recipient) error
	GetRecipients(ctx context.Context, emailID string) ([]emailDomain.Recipient, error)
	GetActiveTemplate(ctx context.Context) (emailDomain.EmailTemplate, error)
}

// MemberLookup defines the interface for looking up member details for recipient resolution.
type MemberLookup interface {
	GetEmailByMemberID(ctx context.Context, memberID string) (name string, email string, err error)
}

// --- Compose Email (Save as Draft) ---

// ComposeEmailInput carries input for composing/saving an email draft.
type ComposeEmailInput struct {
	EmailID   string // Empty for new, set for updating existing draft
	Subject   string
	Body      string
	SenderID  string
	MemberIDs []string // Selected recipient member IDs
}

// ComposeEmailDeps holds dependencies for ComposeEmail.
type ComposeEmailDeps struct {
	EmailStore   EmailStoreForOrchestrator
	MemberLookup MemberLookup
	GenerateID   func() string
	Now          func() time.Time
}

// ExecuteComposeEmail creates or updates an email draft with recipients.
// PRE: SenderID is non-empty; at least one MemberID if saving recipients
// POST: Email saved as draft with recipient list
func ExecuteComposeEmail(ctx context.Context, input ComposeEmailInput, deps ComposeEmailDeps) (emailDomain.Email, error) {
	if input.SenderID == "" {
		return emailDomain.Email{}, errors.New("sender ID is required")
	}

	now := deps.Now()
	var em emailDomain.Email

	if input.EmailID != "" {
		// Update existing draft
		existing, err := deps.EmailStore.GetByID(ctx, input.EmailID)
		if err != nil {
			return emailDomain.Email{}, err
		}
		if !existing.IsDraft() {
			return emailDomain.Email{}, emailDomain.ErrNotDraft
		}
		em = existing
		em.Subject = input.Subject
		em.Body = input.Body
		em.UpdatedAt = now
	} else {
		// New draft
		em = emailDomain.Email{
			ID:        deps.GenerateID(),
			Subject:   input.Subject,
			Body:      input.Body,
			SenderID:  input.SenderID,
			Status:    emailDomain.StatusDraft,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	if err := deps.EmailStore.Save(ctx, em); err != nil {
		return emailDomain.Email{}, err
	}

	// Resolve and save recipients
	if len(input.MemberIDs) > 0 {
		recipients, err := resolveRecipients(ctx, em.ID, input.MemberIDs, deps.MemberLookup)
		if err != nil {
			return emailDomain.Email{}, err
		}
		if err := deps.EmailStore.SaveRecipients(ctx, em.ID, recipients); err != nil {
			return emailDomain.Email{}, err
		}
	}

	slog.Info("email_event", "event", "email_draft_saved", "email_id", em.ID, "sender_id", em.SenderID, "recipient_count", len(input.MemberIDs))
	return em, nil
}

// --- Send Email ---

// SendEmailInput carries input for sending an email.
type SendEmailInput struct {
	EmailID  string
	SenderID string
}

// SendEmailDeps holds dependencies for SendEmail.
type SendEmailDeps struct {
	EmailStore   EmailStoreForOrchestrator
	EmailSender  emailAdapter.Sender
	MemberLookup MemberLookup
	Now          func() time.Time
	FromAddress  string // Default from address
	ReplyTo      string // Reply-to address
}

// ExecuteSendEmail sends a draft email to all its recipients via the email provider.
// PRE: EmailID exists and is in draft status; has at least one recipient
// POST: Email sent via provider, status updated to sent, in-app messages created
func ExecuteSendEmail(ctx context.Context, input SendEmailInput, deps SendEmailDeps) (emailDomain.Email, error) {
	if input.EmailID == "" {
		return emailDomain.Email{}, errors.New("email ID is required")
	}

	em, err := deps.EmailStore.GetByID(ctx, input.EmailID)
	if err != nil {
		return emailDomain.Email{}, err
	}

	if err := em.MarkQueued(); err != nil {
		return emailDomain.Email{}, err
	}

	// Validate the email content
	if err := em.Validate(); err != nil {
		return emailDomain.Email{}, err
	}

	// Get recipients
	recipients, err := deps.EmailStore.GetRecipients(ctx, input.EmailID)
	if err != nil {
		return emailDomain.Email{}, err
	}
	if len(recipients) == 0 {
		return emailDomain.Email{}, emailDomain.ErrNoRecipients
	}

	// Collect email addresses for sending
	var toAddresses []string
	for _, r := range recipients {
		if r.MemberEmail != "" {
			toAddresses = append(toAddresses, r.MemberEmail)
		}
	}

	if len(toAddresses) == 0 {
		return emailDomain.Email{}, emailDomain.ErrNoRecipients
	}

	// Apply active template if one exists
	htmlBody := em.Body
	tpl, tplErr := deps.EmailStore.GetActiveTemplate(ctx)
	if tplErr == nil {
		htmlBody = tpl.WrapBody(em.Body)
		em.TemplateVersionID = tpl.ID
	}

	// Send via provider — one email per recipient for individual delivery
	var sendReqs []emailAdapter.SendRequest
	for _, addr := range toAddresses {
		sendReqs = append(sendReqs, emailAdapter.SendRequest{
			To:      []string{addr},
			From:    deps.FromAddress,
			Subject: em.Subject,
			HTML:    htmlBody,
			ReplyTo: deps.ReplyTo,
		})
	}

	results, err := deps.EmailSender.SendBatch(ctx, sendReqs)
	if err != nil {
		em.MarkFailed()
		deps.EmailStore.Save(ctx, em)
		return em, err
	}

	// Use the first result's message ID as the primary reference
	resendID := ""
	if len(results) > 0 {
		resendID = results[0].MessageID
	}

	em.MarkSent(deps.Now(), resendID)
	if err := deps.EmailStore.Save(ctx, em); err != nil {
		return emailDomain.Email{}, err
	}

	slog.Info("email_event", "event", "email_sent", "email_id", em.ID, "recipient_count", len(toAddresses), "resend_id", resendID)
	return em, nil
}

// --- Test Send Email ---

// TestSendEmailInput carries input for sending a test email to a single address.
type TestSendEmailInput struct {
	EmailID     string
	TestAddress string // The address to send the test to
}

// TestSendEmailDeps holds dependencies for TestSendEmail.
type TestSendEmailDeps struct {
	EmailStore  EmailStoreForOrchestrator
	EmailSender emailAdapter.Sender
	FromAddress string
	ReplyTo     string
}

// ExecuteTestSendEmail sends a single test email to the specified address.
// The draft is NOT modified — status stays as-is, no recipients are recorded.
// PRE: EmailID exists; TestAddress is a valid email; EmailSender is configured
// POST: Exactly one email is delivered to TestAddress; draft is unchanged
func ExecuteTestSendEmail(ctx context.Context, input TestSendEmailInput, deps TestSendEmailDeps) error {
	if input.EmailID == "" {
		return errors.New("email ID is required")
	}
	if input.TestAddress == "" {
		return errors.New("test address is required")
	}

	em, err := deps.EmailStore.GetByID(ctx, input.EmailID)
	if err != nil {
		return err
	}

	if err := em.Validate(); err != nil {
		return err
	}

	// Apply active template if one exists
	htmlBody := em.Body
	tpl, tplErr := deps.EmailStore.GetActiveTemplate(ctx)
	if tplErr == nil {
		htmlBody = tpl.WrapBody(em.Body)
	}

	// Send single email to the test address
	req := emailAdapter.SendRequest{
		To:      []string{input.TestAddress},
		From:    deps.FromAddress,
		Subject: "[TEST] " + em.Subject,
		HTML:    htmlBody,
		ReplyTo: deps.ReplyTo,
	}

	_, err = deps.EmailSender.Send(ctx, req)
	if err != nil {
		return err
	}

	slog.Info("email_event", "event", "test_email_sent", "email_id", em.ID, "test_address", input.TestAddress)
	return nil
}

// --- Schedule Email ---

// ScheduleEmailInput carries input for scheduling an email.
type ScheduleEmailInput struct {
	EmailID     string
	ScheduledAt time.Time
}

// ScheduleEmailDeps holds dependencies for ScheduleEmail.
type ScheduleEmailDeps struct {
	EmailStore EmailStoreForOrchestrator
	Now        func() time.Time
}

// ExecuteScheduleEmail schedules a draft email for future delivery.
// PRE: EmailID exists and is in draft status; ScheduledAt is in the future
// POST: Email status set to scheduled with ScheduledAt time
func ExecuteScheduleEmail(ctx context.Context, input ScheduleEmailInput, deps ScheduleEmailDeps) (emailDomain.Email, error) {
	if input.EmailID == "" {
		return emailDomain.Email{}, errors.New("email ID is required")
	}

	em, err := deps.EmailStore.GetByID(ctx, input.EmailID)
	if err != nil {
		return emailDomain.Email{}, err
	}

	// Validate recipients exist
	recipients, err := deps.EmailStore.GetRecipients(ctx, input.EmailID)
	if err != nil {
		return emailDomain.Email{}, err
	}
	if len(recipients) == 0 {
		return emailDomain.Email{}, emailDomain.ErrNoRecipients
	}

	if err := em.Schedule(input.ScheduledAt); err != nil {
		return emailDomain.Email{}, err
	}
	em.UpdatedAt = deps.Now()

	if err := deps.EmailStore.Save(ctx, em); err != nil {
		return emailDomain.Email{}, err
	}

	slog.Info("email_event", "event", "email_scheduled", "email_id", em.ID, "scheduled_at", input.ScheduledAt)
	return em, nil
}

// --- Cancel Email ---

// CancelEmailInput carries input for cancelling a scheduled email.
type CancelEmailInput struct {
	EmailID string
}

// CancelEmailDeps holds dependencies for CancelEmail.
type CancelEmailDeps struct {
	EmailStore EmailStoreForOrchestrator
	Now        func() time.Time
}

// ExecuteCancelEmail cancels a scheduled or draft email.
// PRE: EmailID exists and is in scheduled or draft status
// POST: Email status set to cancelled
func ExecuteCancelEmail(ctx context.Context, input CancelEmailInput, deps CancelEmailDeps) (emailDomain.Email, error) {
	if input.EmailID == "" {
		return emailDomain.Email{}, errors.New("email ID is required")
	}

	em, err := deps.EmailStore.GetByID(ctx, input.EmailID)
	if err != nil {
		return emailDomain.Email{}, err
	}

	if err := em.Cancel(); err != nil {
		return emailDomain.Email{}, err
	}
	em.UpdatedAt = deps.Now()

	if err := deps.EmailStore.Save(ctx, em); err != nil {
		return emailDomain.Email{}, err
	}

	slog.Info("email_event", "event", "email_cancelled", "email_id", em.ID)
	return em, nil
}

// --- Reschedule Email ---

// RescheduleEmailInput carries input for rescheduling an email.
type RescheduleEmailInput struct {
	EmailID     string
	ScheduledAt time.Time
}

// RescheduleEmailDeps holds dependencies for RescheduleEmail.
type RescheduleEmailDeps struct {
	EmailStore EmailStoreForOrchestrator
	Now        func() time.Time
}

// ExecuteRescheduleEmail changes the scheduled delivery time of an email.
// PRE: EmailID exists and is in scheduled status; ScheduledAt is in the future
// POST: ScheduledAt is updated
func ExecuteRescheduleEmail(ctx context.Context, input RescheduleEmailInput, deps RescheduleEmailDeps) (emailDomain.Email, error) {
	if input.EmailID == "" {
		return emailDomain.Email{}, errors.New("email ID is required")
	}

	em, err := deps.EmailStore.GetByID(ctx, input.EmailID)
	if err != nil {
		return emailDomain.Email{}, err
	}

	if err := em.Reschedule(input.ScheduledAt); err != nil {
		return emailDomain.Email{}, err
	}
	em.UpdatedAt = deps.Now()

	if err := deps.EmailStore.Save(ctx, em); err != nil {
		return emailDomain.Email{}, err
	}

	slog.Info("email_event", "event", "email_rescheduled", "email_id", em.ID, "scheduled_at", input.ScheduledAt)
	return em, nil
}

// resolveRecipients looks up member names and emails from member IDs.
func resolveRecipients(ctx context.Context, emailID string, memberIDs []string, lookup MemberLookup) ([]emailDomain.Recipient, error) {
	var recipients []emailDomain.Recipient
	for _, mid := range memberIDs {
		name, email, err := lookup.GetEmailByMemberID(ctx, mid)
		if err != nil {
			slog.Warn("email_recipient_lookup_failed", "member_id", mid, "error", err)
			continue
		}
		recipients = append(recipients, emailDomain.Recipient{
			EmailID:     emailID,
			MemberID:    mid,
			MemberName:  name,
			MemberEmail: email,
		})
	}
	return recipients, nil
}
