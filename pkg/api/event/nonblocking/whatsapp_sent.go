package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	WhatsappSent event.Type = "whatsapp.sent"
)

type WhatsappSentEventPayload struct {
	Recipient           string `json:"recipient"`
	Type                string `json:"type"`
	IsNotCountedInUsage bool   `json:"is_not_counted_in_usage"`
}

func (e *WhatsappSentEventPayload) NonBlockingEventType() event.Type {
	return WhatsappSent
}

func (e *WhatsappSentEventPayload) UserID() string {
	return ""
}

func (e *WhatsappSentEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *WhatsappSentEventPayload) FillContext(ctx *event.Context) {
}

func (e *WhatsappSentEventPayload) ForHook() bool {
	return false
}

func (e *WhatsappSentEventPayload) ForAudit() bool {
	return true
}

func (e *WhatsappSentEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *WhatsappSentEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &WhatsappSentEventPayload{}
