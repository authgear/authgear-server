package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	EmailSuppressed event.Type = "email.suppressed"
)

type EmailSuppressedEventPayload struct {
	Description string `json:"description"`
}

func (e *EmailSuppressedEventPayload) NonBlockingEventType() event.Type {
	return EmailSuppressed
}

func (e *EmailSuppressedEventPayload) UserID() string {
	return ""
}

func (e *EmailSuppressedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *EmailSuppressedEventPayload) FillContext(ctx *event.Context) {
}

func (e *EmailSuppressedEventPayload) ForHook() bool {
	return false
}

func (e *EmailSuppressedEventPayload) ForAudit() bool {
	return true
}

func (e *EmailSuppressedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *EmailSuppressedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &EmailSuppressedEventPayload{}
