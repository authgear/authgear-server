package auditlog

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const AppCollaboratorAdded event.Type = "site_admin.app.collaborator.added"

type AppCollaboratorAddedPayload struct {
	AppID              string `json:"app_id"`
	CollaboratorID     string `json:"collaborator_id"`
	CollaboratorUserID string `json:"user_id"`
	UserEmail          string `json:"user_email"`
	Role               string `json:"role"`
}

func (e *AppCollaboratorAddedPayload) NonBlockingEventType() event.Type {
	return AppCollaboratorAdded
}

func (e *AppCollaboratorAddedPayload) UserID() string {
	return ""
}

func (e *AppCollaboratorAddedPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredBySiteAdmin
}

func (e *AppCollaboratorAddedPayload) FillContext(_ *event.Context) {}

func (e *AppCollaboratorAddedPayload) ForHook() bool {
	return false
}

func (e *AppCollaboratorAddedPayload) ForAudit() bool {
	return true
}

func (e *AppCollaboratorAddedPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AppCollaboratorAddedPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AppCollaboratorAddedPayload{}
