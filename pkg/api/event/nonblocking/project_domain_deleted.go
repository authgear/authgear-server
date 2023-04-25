package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectDomainDeleted event.Type = "project.domain.deleted"
)

type ProjectDomainDeletedEventPayload struct {
	Domain   string `json:"domain"`
	DomainID string `json:"domain_id"`
}

func (e *ProjectDomainDeletedEventPayload) NonBlockingEventType() event.Type {
	return ProjectDomainDeleted
}

func (e *ProjectDomainDeletedEventPayload) UserID() string {
	return ""
}

func (e *ProjectDomainDeletedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByPortal
}

func (e *ProjectDomainDeletedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectDomainDeletedEventPayload) ForHook() bool {
	return false
}

func (e *ProjectDomainDeletedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectDomainDeletedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *ProjectDomainDeletedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &ProjectDomainDeletedEventPayload{}
