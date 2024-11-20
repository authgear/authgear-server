package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	SMSError event.Type = "sms.error"
)

type SMSErrorEventPayload struct {
	Description string `json:"description"`
}

func (e *SMSErrorEventPayload) NonBlockingEventType() event.Type {
	return SMSError
}

func (e *SMSErrorEventPayload) UserID() string {
	return ""
}

func (e *SMSErrorEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *SMSErrorEventPayload) FillContext(ctx *event.Context) {
}

func (e *SMSErrorEventPayload) ForHook() bool {
	return false
}

func (e *SMSErrorEventPayload) ForAudit() bool {
	return true
}

func (e *SMSErrorEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *SMSErrorEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &SMSErrorEventPayload{}
