package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UsageAlertTriggered event.Type = "usage.alert.triggered"
)

type UsageAlertPayload struct {
	Name         model.UsageName        `json:"name"`
	Action       model.UsageLimitAction `json:"action"`
	Period       model.UsageLimitPeriod `json:"period"`
	Quota        int                    `json:"quota"`
	CurrentValue int                    `json:"current_value"`
}

type UsageAlertTriggeredEventPayload struct {
	Usage    UsageAlertPayload `json:"usage"`
	HookURLs []string          `json:"-"`
}

func (e *UsageAlertTriggeredEventPayload) NonBlockingEventType() event.Type {
	return UsageAlertTriggered
}

func (e *UsageAlertTriggeredEventPayload) UserID() string {
	return ""
}

func (e *UsageAlertTriggeredEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredBySystem
}

func (e *UsageAlertTriggeredEventPayload) FillContext(ctx *event.Context) {}

func (e *UsageAlertTriggeredEventPayload) ForHook() bool {
	return true
}

func (e *UsageAlertTriggeredEventPayload) ForAudit() bool {
	return true
}

func (e *UsageAlertTriggeredEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *UsageAlertTriggeredEventPayload) DeletedUserIDs() []string {
	return nil
}

func (e *UsageAlertTriggeredEventPayload) ExtraHookURLs() []string {
	return e.HookURLs
}

var _ event.NonBlockingPayload = &UsageAlertTriggeredEventPayload{}
var _ event.ExtraHookURLsProvider = &UsageAlertTriggeredEventPayload{}
