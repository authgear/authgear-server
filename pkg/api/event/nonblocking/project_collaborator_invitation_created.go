package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectCollaboratorInvitationCreated event.Type = "project.collaborator.invitation.created"
)

type ProjectCollaboratorInvitationCreatedEventPayload struct {
	InviteeEmail string `json:"invitee_email"`
	InvitedBy    string `json:"invited_by"`
}

func (e *ProjectCollaboratorInvitationCreatedEventPayload) NonBlockingEventType() event.Type {
	return ProjectCollaboratorInvitationCreated
}

func (e *ProjectCollaboratorInvitationCreatedEventPayload) UserID() string {
	return ""
}

func (e *ProjectCollaboratorInvitationCreatedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByPortal
}

func (e *ProjectCollaboratorInvitationCreatedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectCollaboratorInvitationCreatedEventPayload) ForHook() bool {
	return true
}

func (e *ProjectCollaboratorInvitationCreatedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectCollaboratorInvitationCreatedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *ProjectCollaboratorInvitationCreatedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &ProjectCollaboratorInvitationCreatedEventPayload{}
