package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectDomainCreated event.Type = "project.domain.created"
)

type ProjectDomainCreatedEventPayload struct {
	Domain   string `json:"domain"`
	DomainID string `json:"domain_id"`
}

func (e *ProjectDomainCreatedEventPayload) NonBlockingEventType() event.Type {
	return ProjectDomainCreated
}

func (e *ProjectDomainCreatedEventPayload) UserID() string {
	return ""
}

func (e *ProjectDomainCreatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *ProjectDomainCreatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectDomainCreatedEventPayload) ForHook() bool {
	return false
}

func (e *ProjectDomainCreatedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectDomainCreatedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *ProjectDomainCreatedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &ProjectDomainCreatedEventPayload{}
