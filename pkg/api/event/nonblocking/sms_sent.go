package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	SMSSent event.Type = "sms.sent"
)

type SMSSentEventPayload struct {
	Sender              string `json:"sender"`
	Recipient           string `json:"recipient"`
	Type                string `json:"type"`
	IsNotCountedInUsage bool   `json:"is_not_counted_in_usage"`
}

func (e *SMSSentEventPayload) NonBlockingEventType() event.Type {
	return SMSSent
}

func (e *SMSSentEventPayload) UserID() string {
	return ""
}

func (e *SMSSentEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *SMSSentEventPayload) FillContext(ctx *event.Context) {
}

func (e *SMSSentEventPayload) ForHook() bool {
	return false
}

func (e *SMSSentEventPayload) ForAudit() bool {
	return true
}

func (e *SMSSentEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *SMSSentEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &SMSSentEventPayload{}
