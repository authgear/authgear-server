package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	SiteAdminAppPlanUpdated event.Type = "site_admin.app.plan.updated"
)

type SiteAdminAppPlanUpdatedEventPayload struct {
	AppID   string `json:"app_id"`
	OldPlan string `json:"old_plan"`
	NewPlan string `json:"new_plan"`
}

func (e *SiteAdminAppPlanUpdatedEventPayload) NonBlockingEventType() event.Type {
	return SiteAdminAppPlanUpdated
}

func (e *SiteAdminAppPlanUpdatedEventPayload) UserID() string {
	return ""
}

func (e *SiteAdminAppPlanUpdatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredBySiteAdmin
}

func (e *SiteAdminAppPlanUpdatedEventPayload) FillContext(_ *event.Context) {}

func (e *SiteAdminAppPlanUpdatedEventPayload) ForHook() bool {
	return false
}

func (e *SiteAdminAppPlanUpdatedEventPayload) ForAudit() bool {
	return true
}

func (e *SiteAdminAppPlanUpdatedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *SiteAdminAppPlanUpdatedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &SiteAdminAppPlanUpdatedEventPayload{}
