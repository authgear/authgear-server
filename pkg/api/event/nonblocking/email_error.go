package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	EmailError event.Type = "email.error"
)

type EmailErrorEventPayload struct {
	Description string `json:"description"`
}

func (e *EmailErrorEventPayload) NonBlockingEventType() event.Type {
	return EmailError
}

func (e *EmailErrorEventPayload) UserID() string {
	return ""
}

func (e *EmailErrorEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *EmailErrorEventPayload) FillContext(ctx *event.Context) {
}

func (e *EmailErrorEventPayload) ForHook() bool {
	return false
}

func (e *EmailErrorEventPayload) ForAudit() bool {
	return true
}

func (e *EmailErrorEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *EmailErrorEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &EmailErrorEventPayload{}
