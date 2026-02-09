package email

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/resend/resend-go/v2"
)

// ResendSender sends emails via the Resend API.
type ResendSender struct {
	client *resend.Client
	from   string
}

// NewResendSender creates a new ResendSender with the given API key and default from address.
// PRE: apiKey is a valid Resend API key; from is a valid sender address
// POST: Returns a ready-to-use sender
func NewResendSender(apiKey, from string) *ResendSender {
	return &ResendSender{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

// Send sends a single email via Resend.
// PRE: req has at least one recipient and a subject
// POST: Email is queued for delivery; returns the Resend message ID
func (s *ResendSender) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	from := req.From
	if from == "" {
		from = s.from
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      req.To,
		Subject: req.Subject,
		Html:    req.HTML,
	}
	if req.ReplyTo != "" {
		params.ReplyTo = req.ReplyTo
	}

	sent, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		slog.Error("resend_send_failed", "error", err, "to", req.To, "subject", req.Subject)
		return SendResult{}, fmt.Errorf("resend send failed: %w", err)
	}

	slog.Info("resend_sent", "message_id", sent.Id, "to", req.To, "subject", req.Subject)
	return SendResult{
		MessageID: sent.Id,
		SentAt:    time.Now(),
	}, nil
}

// SendBatch sends multiple emails via Resend's batch API (up to 100 per call).
// PRE: len(reqs) > 0
// POST: All emails are queued; returns results in the same order as requests
func (s *ResendSender) SendBatch(ctx context.Context, reqs []SendRequest) ([]SendResult, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	// Resend batch API supports up to 100 emails per call
	const batchSize = 100
	var allResults []SendResult

	for i := 0; i < len(reqs); i += batchSize {
		end := i + batchSize
		if end > len(reqs) {
			end = len(reqs)
		}
		chunk := reqs[i:end]

		var batchParams []*resend.SendEmailRequest
		for _, req := range chunk {
			from := req.From
			if from == "" {
				from = s.from
			}
			p := &resend.SendEmailRequest{
				From:    from,
				To:      req.To,
				Subject: req.Subject,
				Html:    req.HTML,
			}
			if req.ReplyTo != "" {
				p.ReplyTo = req.ReplyTo
			}
			batchParams = append(batchParams, p)
		}

		resp, err := s.client.Batch.SendWithContext(ctx, batchParams)
		if err != nil {
			slog.Error("resend_batch_failed", "error", err, "batch_size", len(chunk))
			return allResults, fmt.Errorf("resend batch send failed: %w", err)
		}

		for _, item := range resp.Data {
			allResults = append(allResults, SendResult{
				MessageID: item.Id,
				SentAt:    time.Now(),
			})
		}

		slog.Info("resend_batch_sent", "count", len(chunk), "total_sent", len(allResults))
	}

	return allResults, nil
}
