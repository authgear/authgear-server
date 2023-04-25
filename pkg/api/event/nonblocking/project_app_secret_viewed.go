package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectAppSecretViewed event.Type = "project.app.secret.viewed"
)

type ProjectAppSecretViewedEventPayload struct {
	Secrets []string `json:"secrets"`
}

func (e *ProjectAppSecretViewedEventPayload) NonBlockingEventType() event.Type {
	return ProjectAppSecretViewed
}

func (e *ProjectAppSecretViewedEventPayload) UserID() string {
	return ""
}

func (e *ProjectAppSecretViewedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByPortal
}

func (e *ProjectAppSecretViewedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectAppSecretViewedEventPayload) ForHook() bool {
	return false
}

func (e *ProjectAppSecretViewedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectAppSecretViewedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *ProjectAppSecretViewedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &ProjectAppSecretViewedEventPayload{}
