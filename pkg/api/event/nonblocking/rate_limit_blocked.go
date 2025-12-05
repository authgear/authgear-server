package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	RateLimitBlocked event.Type = "rate_limit.blocked"
)

type RateLimitBlockedEventPayload struct {
	RateLimit string `json:"rate_limit"`
	Bucket    string `json:"bucket"`
}

func (e *RateLimitBlockedEventPayload) NonBlockingEventType() event.Type {
	return RateLimitBlocked
}

func (e *RateLimitBlockedEventPayload) UserID() string {
	return ""
}

func (e *RateLimitBlockedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *RateLimitBlockedEventPayload) FillContext(ctx *event.Context) {}

func (e *RateLimitBlockedEventPayload) ForHook() bool {
	return false
}

func (e *RateLimitBlockedEventPayload) ForAudit() bool {
	return true
}

func (e *RateLimitBlockedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *RateLimitBlockedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &RateLimitBlockedEventPayload{}
