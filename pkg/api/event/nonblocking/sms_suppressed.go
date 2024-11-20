package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	SMSSuppressed event.Type = "sms.suppressed"
)

type SMSSuppressedEventPayload struct {
	Description string `json:"description"`
}

func (e *SMSSuppressedEventPayload) NonBlockingEventType() event.Type {
	return SMSSuppressed
}

func (e *SMSSuppressedEventPayload) UserID() string {
	return ""
}

func (e *SMSSuppressedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *SMSSuppressedEventPayload) FillContext(ctx *event.Context) {
}

func (e *SMSSuppressedEventPayload) ForHook() bool {
	return false
}

func (e *SMSSuppressedEventPayload) ForAudit() bool {
	return true
}

func (e *SMSSuppressedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *SMSSuppressedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &SMSSuppressedEventPayload{}
