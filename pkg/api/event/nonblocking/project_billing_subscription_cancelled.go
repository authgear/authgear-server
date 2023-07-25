package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectBillingSubscriptionCancelled event.Type = "project.billing.subscription.cancelled"
)

type ProjectBillingSubscriptionCancelledEventPayload struct {
	SubscriptionID string `json:"subscription_id"`
	CustomerID     string `json:"customer_id"`
}

func (e *ProjectBillingSubscriptionCancelledEventPayload) NonBlockingEventType() event.Type {
	return ProjectBillingSubscriptionCancelled
}

func (e *ProjectBillingSubscriptionCancelledEventPayload) UserID() string {
	return ""
}

func (e *ProjectBillingSubscriptionCancelledEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByPortal
}

func (e *ProjectBillingSubscriptionCancelledEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectBillingSubscriptionCancelledEventPayload) ForHook() bool {
	return true
}

func (e *ProjectBillingSubscriptionCancelledEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectBillingSubscriptionCancelledEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *ProjectBillingSubscriptionCancelledEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &ProjectBillingSubscriptionCancelledEventPayload{}
