package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	EmailSent event.Type = "email.sent"
)

type EmailSentEventPayload struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Type      string `json:"type"`
}

func (e *EmailSentEventPayload) NonBlockingEventType() event.Type {
	return EmailSent
}

func (e *EmailSentEventPayload) UserID() string {
	return ""
}

func (e *EmailSentEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *EmailSentEventPayload) FillContext(ctx *event.Context) {
}

func (e *EmailSentEventPayload) ForHook() bool {
	return false
}

func (e *EmailSentEventPayload) ForAudit() bool {
	return true
}

func (e *EmailSentEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *EmailSentEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &EmailSentEventPayload{}
