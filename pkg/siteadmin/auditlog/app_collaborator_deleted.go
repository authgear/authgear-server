package auditlog

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const AppCollaboratorDeleted event.Type = "site_admin.app.collaborator.deleted"

type AppCollaboratorDeletedPayload struct {
	AppID                 string `json:"app_id"`
	CollaboratorID        string `json:"collaborator_id"`
	CollaboratorUserID    string `json:"user_id"`
	CollaboratorUserEmail string `json:"user_email"`
}

func (e *AppCollaboratorDeletedPayload) NonBlockingEventType() event.Type {
	return AppCollaboratorDeleted
}

func (e *AppCollaboratorDeletedPayload) UserID() string {
	return ""
}

func (e *AppCollaboratorDeletedPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredBySiteAdmin
}

func (e *AppCollaboratorDeletedPayload) FillContext(_ *event.Context) {}

func (e *AppCollaboratorDeletedPayload) ForHook() bool {
	return false
}

func (e *AppCollaboratorDeletedPayload) ForAudit() bool {
	return true
}

func (e *AppCollaboratorDeletedPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AppCollaboratorDeletedPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AppCollaboratorDeletedPayload{}
