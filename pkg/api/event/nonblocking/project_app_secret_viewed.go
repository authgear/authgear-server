package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectAppSecretViewed event.Type = "project.app.secret.viewed" // nolint:gosec
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
	return true
}

func (e *ProjectAppSecretViewedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectAppSecretViewedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *ProjectAppSecretViewedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &ProjectAppSecretViewedEventPayload{}
