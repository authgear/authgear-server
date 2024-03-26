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
	return event.TriggeredByPortal
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) ForHook() bool {
	return true
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *ProjectCollaboratorInvitationDeletedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &ProjectCollaboratorInvitationDeletedEventPayload{}
