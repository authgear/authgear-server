package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	WhatsappSuppressed event.Type = "whatsapp.suppressed"
)

type WhatsappSuppressedEventPayload struct {
	Description string `json:"description"`
}

func (e *WhatsappSuppressedEventPayload) NonBlockingEventType() event.Type {
	return WhatsappSuppressed
}

func (e *WhatsappSuppressedEventPayload) UserID() string {
	return ""
}

func (e *WhatsappSuppressedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *WhatsappSuppressedEventPayload) FillContext(ctx *event.Context) {
}

func (e *WhatsappSuppressedEventPayload) ForHook() bool {
	return false
}

func (e *WhatsappSuppressedEventPayload) ForAudit() bool {
	return true
}

func (e *WhatsappSuppressedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *WhatsappSuppressedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &WhatsappSuppressedEventPayload{}
