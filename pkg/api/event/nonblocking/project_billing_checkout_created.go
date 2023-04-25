package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectBillingCheckoutCreated event.Type = "project.billing.checkout.created"
)

type ProjectBillingCheckoutCreatedEventPayload struct {
	SubscriptionCheckoutID string `json:"subscription_checkout_id"`
	PlanName               string `json:"plan_name"`
}

func (e *ProjectBillingCheckoutCreatedEventPayload) NonBlockingEventType() event.Type {
	return ProjectBillingCheckoutCreated
}

func (e *ProjectBillingCheckoutCreatedEventPayload) UserID() string {
	return ""
}

func (e *ProjectBillingCheckoutCreatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByPortal
}

func (e *ProjectBillingCheckoutCreatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectBillingCheckoutCreatedEventPayload) ForHook() bool {
	return false
}

func (e *ProjectBillingCheckoutCreatedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectBillingCheckoutCreatedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *ProjectBillingCheckoutCreatedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &ProjectBillingCheckoutCreatedEventPayload{}
