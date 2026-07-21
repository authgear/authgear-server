package auditlog

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const AppPlanUpdated event.Type = "site_admin.app.plan.updated"

type AppPlanUpdatedPayload struct {
	AppID   string `json:"app_id"`
	OldPlan string `json:"old_plan"`
	NewPlan string `json:"new_plan"`
}

func (e *AppPlanUpdatedPayload) NonBlockingEventType() event.Type {
	return AppPlanUpdated
}

func (e *AppPlanUpdatedPayload) UserID() string {
	return ""
}

func (e *AppPlanUpdatedPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredBySiteAdmin
}

func (e *AppPlanUpdatedPayload) FillContext(_ *event.Context) {}

func (e *AppPlanUpdatedPayload) ForHook() bool {
	return false
}

func (e *AppPlanUpdatedPayload) ForAudit() bool {
	return true
}

func (e *AppPlanUpdatedPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AppPlanUpdatedPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AppPlanUpdatedPayload{}
