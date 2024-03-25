package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectBillingSubscriptionUpdated event.Type = "project.billing.subscription.updated"
)

type ProjectBillingSubscriptionUpdatedEventPayload struct {
	SubscriptionID string `json:"subscription_id"`
	PlanName       string `json:"plan_name"`
}

func (e *ProjectBillingSubscriptionUpdatedEventPayload) NonBlockingEventType() event.Type {
	return ProjectBillingSubscriptionUpdated
}

func (e *ProjectBillingSubscriptionUpdatedEventPayload) UserID() string {
	return ""
}

func (e *ProjectBillingSubscriptionUpdatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByPortal
}

func (e *ProjectBillingSubscriptionUpdatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectBillingSubscriptionUpdatedEventPayload) ForHook() bool {
	return true
}

func (e *ProjectBillingSubscriptionUpdatedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectBillingSubscriptionUpdatedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *ProjectBillingSubscriptionUpdatedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &ProjectBillingSubscriptionUpdatedEventPayload{}
