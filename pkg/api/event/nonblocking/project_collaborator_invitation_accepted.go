package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	ProjectCollaboratorInvitationAccepted event.Type = "project.collaborator.invitation.accepted"
)

type ProjectCollaboratorInvitationAcceptedEventPayload struct {
	CollaboratorUserID string `json:"collaborator_user_id"`
	CollaboratorRole   string `json:"collaborator_role"`
}

func (e *ProjectCollaboratorInvitationAcceptedEventPayload) NonBlockingEventType() event.Type {
	return ProjectCollaboratorInvitationAccepted
}

func (e *ProjectCollaboratorInvitationAcceptedEventPayload) UserID() string {
	return ""
}

func (e *ProjectCollaboratorInvitationAcceptedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByPortal
}

func (e *ProjectCollaboratorInvitationAcceptedEventPayload) FillContext(ctx *event.Context) {
}

func (e *ProjectCollaboratorInvitationAcceptedEventPayload) ForHook() bool {
	return true
}

func (e *ProjectCollaboratorInvitationAcceptedEventPayload) ForAudit() bool {
	return true
}

func (e *ProjectCollaboratorInvitationAcceptedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *ProjectCollaboratorInvitationAcceptedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &ProjectCollaboratorInvitationAcceptedEventPayload{}
