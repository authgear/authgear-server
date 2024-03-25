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
	return event.TriggeredByPortal
}

func (e *ProjectDomainVerifiedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectDomainVerifiedEventPayload) ForHook() bool {
	return true
}

func (e *ProjectDomainVerifiedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectDomainVerifiedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *ProjectDomainVerifiedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &ProjectDomainVerifiedEventPayload{}
