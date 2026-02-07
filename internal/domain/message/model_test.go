package message_test

import (
	"testing"
	"time"

	"workshop/internal/domain/message"
)

// TestMessage_Validate tests validation of Message.
func TestMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		msg     message.Message
		wantErr bool
	}{
		{
			name:    "valid message",
			msg:     message.Message{ID: "1", SenderID: "admin1", ReceiverID: "m1", Content: "Hello!", CreatedAt: time.Now()},
			wantErr: false,
		},
		{
			name:    "empty sender",
			msg:     message.Message{ID: "2", ReceiverID: "m1", Content: "Hello!", CreatedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "empty receiver",
			msg:     message.Message{ID: "3", SenderID: "admin1", Content: "Hello!", CreatedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "empty content",
			msg:     message.Message{ID: "4", SenderID: "admin1", ReceiverID: "m1", CreatedAt: time.Now()},
			wantErr: true,
		},
		{
			name:    "zero created_at",
			msg:     message.Message{ID: "5", SenderID: "admin1", ReceiverID: "m1", Content: "Hello!"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMessage_ReadStatus tests IsRead and MarkRead on Message.
func TestMessage_ReadStatus(t *testing.T) {
	t.Run("unread message", func(t *testing.T) {
		m := message.Message{ID: "1", SenderID: "a", ReceiverID: "b", Content: "c", CreatedAt: time.Now()}
		if m.IsRead() {
			t.Error("new message should be unread")
		}
	})

	t.Run("mark read", func(t *testing.T) {
		m := message.Message{ID: "1", SenderID: "a", ReceiverID: "b", Content: "c", CreatedAt: time.Now()}
		m.MarkRead()
		if !m.IsRead() {
			t.Error("message should be read after MarkRead")
		}
	})

	t.Run("mark read idempotent", func(t *testing.T) {
		m := message.Message{ID: "1", SenderID: "a", ReceiverID: "b", Content: "c", CreatedAt: time.Now()}
		m.MarkRead()
		first := m.ReadAt
		m.MarkRead()
		if m.ReadAt != first {
			t.Error("MarkRead should be idempotent")
		}
	})
}
