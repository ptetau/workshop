package email

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// NoopSender is a no-op email sender for development and testing.
// It logs sends but does not actually deliver emails.
type NoopSender struct{}

// NewNoopSender creates a new NoopSender.
func NewNoopSender() *NoopSender {
	return &NoopSender{}
}

// Send logs the email but does not deliver it.
// PRE: req is a valid SendRequest
// POST: Returns a noop result without actual delivery
func (s *NoopSender) Send(_ context.Context, req SendRequest) (SendResult, error) {
	slog.Info("noop_email_send", "to", req.To, "subject", req.Subject)
	return SendResult{
		MessageID: fmt.Sprintf("noop-%d", time.Now().UnixNano()),
		SentAt:    time.Now(),
	}, nil
}

// SendBatch logs the batch but does not deliver.
// PRE: reqs is a slice of SendRequests
// POST: Returns noop results for each request without actual delivery
func (s *NoopSender) SendBatch(_ context.Context, reqs []SendRequest) ([]SendResult, error) {
	var results []SendResult
	for i, req := range reqs {
		slog.Info("noop_email_batch", "index", i, "to", req.To, "subject", req.Subject)
		results = append(results, SendResult{
			MessageID: fmt.Sprintf("noop-batch-%d-%d", time.Now().UnixNano(), i),
			SentAt:    time.Now(),
		})
	}
	return results, nil
}
