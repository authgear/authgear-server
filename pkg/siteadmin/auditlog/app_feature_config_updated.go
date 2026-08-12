package auditlog

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const AppFeatureConfigUpdated event.Type = "site_admin.app.feature_config.updated"

type AppFeatureConfigUpdatedPayload struct {
	AppID                   string `json:"app_id"`
	OldAppFeatureConfigYAML string `json:"old_app_feature_config_yaml"`
	NewAppFeatureConfigYAML string `json:"new_app_feature_config_yaml"`
}

func (e *AppFeatureConfigUpdatedPayload) NonBlockingEventType() event.Type {
	return AppFeatureConfigUpdated
}

func (e *AppFeatureConfigUpdatedPayload) UserID() string {
	return ""
}

func (e *AppFeatureConfigUpdatedPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredBySiteAdmin
}

func (e *AppFeatureConfigUpdatedPayload) FillContext(_ *event.Context) {}

func (e *AppFeatureConfigUpdatedPayload) ForHook() bool {
	return false
}

func (e *AppFeatureConfigUpdatedPayload) ForAudit() bool {
	return true
}

func (e *AppFeatureConfigUpdatedPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AppFeatureConfigUpdatedPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AppFeatureConfigUpdatedPayload{}
