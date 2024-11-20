package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	WhatsappError event.Type = "whatsapp.error"
)

type WhatsappErrorEventPayload struct {
	Description string `json:"description"`
}

func (e *WhatsappErrorEventPayload) NonBlockingEventType() event.Type {
	return WhatsappError
}

func (e *WhatsappErrorEventPayload) UserID() string {
	return ""
}

func (e *WhatsappErrorEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *WhatsappErrorEventPayload) FillContext(ctx *event.Context) {
}

func (e *WhatsappErrorEventPayload) ForHook() bool {
	return false
}

func (e *WhatsappErrorEventPayload) ForAudit() bool {
	return true
}

func (e *WhatsappErrorEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *WhatsappErrorEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &WhatsappErrorEventPayload{}
