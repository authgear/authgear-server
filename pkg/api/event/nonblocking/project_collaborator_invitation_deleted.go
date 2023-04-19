package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectCollaboratorInvitationDeleted event.Type = "project.collaborator.invitation.deleted"
)

type ProjectCollaboratorInvitationDeletedEventPayload struct {
	InviteeEmail string `json:"invitee_email"`
	InvitedBy    string `json:"invited_by"`
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) NonBlockingEventType() event.Type {
	return ProjectCollaboratorInvitationDeleted
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) UserID() string {
	return ""
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) ForHook() bool {
	return false
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) ReindexUserNeeded() bool {
	return false
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) IsUserDeleted() bool {
	return false
}

var _ event.NonBlockingPayload = &ProjectCollaboratorInvitationDeletedEventPayload{}
