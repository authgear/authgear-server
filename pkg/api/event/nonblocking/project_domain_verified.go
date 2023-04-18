package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectDomainVerified event.Type = "project.domain.verified"
)

type ProjectDomainVerifiedEventPayload struct {
	Domain   string `json:"domain"`
	DomainID string `json:"domain_id"`
}

func (e *ProjectDomainVerifiedEventPayload) NonBlockingEventType() event.Type {
	return ProjectDomainVerified
}

func (e *ProjectDomainVerifiedEventPayload) UserID() string {
	return ""
}

func (e *ProjectDomainVerifiedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *ProjectDomainVerifiedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectDomainVerifiedEventPayload) ForHook() bool {
	return false
}

func (e *ProjectDomainVerifiedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectDomainVerifiedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *ProjectDomainVerifiedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &ProjectDomainVerifiedEventPayload{}
