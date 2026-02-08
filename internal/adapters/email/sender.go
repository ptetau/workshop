package email

import (
	"context"
	"time"
)

// SendRequest contains the data needed to send an email via an external provider.
type SendRequest struct {
	To      []string // Recipient email addresses
	From    string   // Sender address (e.g. "Workshop Jiu Jitsu <noreply@workshopjiujitsu.co.nz>")
	Subject string
	HTML    string // HTML body
	ReplyTo string // Reply-to address
}

// SendResult contains the response from the email provider.
type SendResult struct {
	MessageID string    // Provider's message ID for tracking
	SentAt    time.Time // When the send was accepted
}

// Sender is the interface for sending emails via an external provider.
type Sender interface {
	Send(ctx context.Context, req SendRequest) (SendResult, error)
	SendBatch(ctx context.Context, reqs []SendRequest) ([]SendResult, error)
}
