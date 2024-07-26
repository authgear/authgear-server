package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	BotProtectionVerificationFailed event.Type = "bot_protection.verification.failed"
)

type BotProtectionVerificationFailedEventPayload struct {
	// Event.Context already have Timestamp & IPAddress, no need specify here
	Token        string `json:"token"`
	ProviderType string `json:"provider_type"`
}

func (e *BotProtectionVerificationFailedEventPayload) NonBlockingEventType() event.Type {
	return BotProtectionVerificationFailed
}

func (e *BotProtectionVerificationFailedEventPayload) UserID() string {
	return ""
}

func (e *BotProtectionVerificationFailedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *BotProtectionVerificationFailedEventPayload) FillContext(ctx *event.Context) {}

func (e *BotProtectionVerificationFailedEventPayload) ForHook() bool {
	return false
}

func (e *BotProtectionVerificationFailedEventPayload) ForAudit() bool {
	return true
}

func (e *BotProtectionVerificationFailedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *BotProtectionVerificationFailedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &BotProtectionVerificationFailedEventPayload{}
