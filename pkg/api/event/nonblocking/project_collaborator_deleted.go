package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectCollaboratorDeleted event.Type = "project.collaborator.deleted"
)

type ProjectCollaboratorDeletedEventPayload struct {
	CollaboratorID     string `json:"collaborator_id"`
	CollaboratorUserID string `json:"collaborator_user_id"`
	CollaboratorRole   string `json:"collaborator_role"`
}

func (e *ProjectCollaboratorDeletedEventPayload) NonBlockingEventType() event.Type {
	return ProjectCollaboratorDeleted
}

func (e *ProjectCollaboratorDeletedEventPayload) UserID() string {
	return ""
}

func (e *ProjectCollaboratorDeletedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByPortal
}

func (e *ProjectCollaboratorDeletedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectCollaboratorDeletedEventPayload) ForHook() bool {
	return true
}

func (e *ProjectCollaboratorDeletedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectCollaboratorDeletedEventPayload) RequireReindexUserIDs() []string {
	return []string{}
}

func (e *ProjectCollaboratorDeletedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &ProjectCollaboratorDeletedEventPayload{}
