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
	return event.TriggeredByTypeUser
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) ForHook() bool {
	return false
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *ProjectBillingSubscriptionStatusUpdatedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &ProjectBillingSubscriptionStatusUpdatedEventPayload{}
