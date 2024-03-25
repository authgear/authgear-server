package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectAppCreated event.Type = "project.app.created"
)

type ProjectAppCreatedEventPayload struct {
	AppConfig AppConfig `json:"app_config,omitempty"`
}

func (*ProjectAppCreatedEventPayload) NonBlockingEventType() event.Type {
	return ProjectAppCreated
}

func (*ProjectAppCreatedEventPayload) UserID() string {
	return ""
}

func (*ProjectAppCreatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByPortal
}

func (*ProjectAppCreatedEventPayload) FillContext(ctx *event.Context) {
}

func (*ProjectAppCreatedEventPayload) ForHook() bool {
	return true
}

func (*ProjectAppCreatedEventPayload) ForAudit() bool {
	return true
}

func (*ProjectAppCreatedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (*ProjectAppCreatedEventPayload) DeletedUserIDs() []string {
	return nil
}
