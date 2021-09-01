package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	EmailSent event.Type = "email.sent"
)

type EmailSentEventPayload struct {
	Sender    string      `json:"sender"`
	Recipient string      `json:"recipient"`
	Type      MessageType `json:"type"`
}

func (e *EmailSentEventPayload) NonBlockingEventType() event.Type {
	return EmailSent
}

func (e *EmailSentEventPayload) UserID() string {
	return ""
}

func (e *EmailSentEventPayload) IsAdminAPI() bool {
	return false
}

func (e *EmailSentEventPayload) FillContext(ctx *event.Context) {
}

func (e *EmailSentEventPayload) ForWebHook() bool {
	return false
}

func (e *EmailSentEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &EmailSentEventPayload{}
