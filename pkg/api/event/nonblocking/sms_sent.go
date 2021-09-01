package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	SMSSent event.Type = "sms.sent"
)

type SMSSentEventPayload struct {
	Sender    string      `json:"sender"`
	Recipient string      `json:"recipient"`
	Type      MessageType `json:"type"`
}

func (e *SMSSentEventPayload) NonBlockingEventType() event.Type {
	return SMSSent
}

func (e *SMSSentEventPayload) UserID() string {
	return ""
}

func (e *SMSSentEventPayload) IsAdminAPI() bool {
	return false
}

func (e *SMSSentEventPayload) FillContext(ctx *event.Context) {
}

func (e *SMSSentEventPayload) ForWebHook() bool {
	return false
}

func (e *SMSSentEventPayload) ForAudit() bool {
	return true
}

var _ event.NonBlockingPayload = &SMSSentEventPayload{}
