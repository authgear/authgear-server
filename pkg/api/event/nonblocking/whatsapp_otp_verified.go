package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	WhatsappOTPVerified event.Type = "whatsapp.otp.verified"
)

// Deprecated. Do not use this event.
type WhatsappOTPVerifiedEventPayload struct {
	Phone string `json:"phone"`
}

func (e *WhatsappOTPVerifiedEventPayload) NonBlockingEventType() event.Type {
	return WhatsappOTPVerified
}

func (e *WhatsappOTPVerifiedEventPayload) UserID() string {
	return ""
}

func (e *WhatsappOTPVerifiedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *WhatsappOTPVerifiedEventPayload) FillContext(ctx *event.Context) {
}

func (e *WhatsappOTPVerifiedEventPayload) ForHook() bool {
	return false
}

func (e *WhatsappOTPVerifiedEventPayload) ForAudit() bool {
	return true
}

func (e *WhatsappOTPVerifiedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *WhatsappOTPVerifiedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &WhatsappOTPVerifiedEventPayload{}
