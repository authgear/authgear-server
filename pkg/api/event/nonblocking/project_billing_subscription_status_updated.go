package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectBillingSubscriptionStatusUpdated event.Type = "project.billing.subscription.status.updated"
)

type ProjectBillingSubscriptionStatusUpdatedEventPayload struct {
	SubscriptionID string `json:"subscription_id"`
	Cancelled      bool   `json:"cancelled"`
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) NonBlockingEventType() event.Type {
	return ProjectBillingSubscriptionStatusUpdated
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) UserID() string {
	return ""
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByPortal
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) ForHook() bool {
	return true
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &ProjectBillingSubscriptionStatusUpdatedEventPayload{}
