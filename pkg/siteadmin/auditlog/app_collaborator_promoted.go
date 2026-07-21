package auditlog

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const AppCollaboratorPromoted event.Type = "site_admin.app.collaborator.promoted"

type AppCollaboratorPromotedPayload struct {
	AppID                  string `json:"app_id"`
	NewOwnerCollaboratorID string `json:"new_owner_collaborator_id"`
	NewOwnerUserID         string `json:"new_owner_user_id"`
	NewOwnerUserEmail      string `json:"new_owner_user_email"`
	// DemotedEditor* fields are omitted when the app had no previous owner.
	DemotedEditorCollaboratorID string `json:"demoted_editor_collaborator_id,omitempty"`
	DemotedEditorUserID         string `json:"demoted_editor_user_id,omitempty"`
	DemotedEditorUserEmail      string `json:"demoted_editor_user_email,omitempty"`
}

func (e *AppCollaboratorPromotedPayload) NonBlockingEventType() event.Type {
	return AppCollaboratorPromoted
}

func (e *AppCollaboratorPromotedPayload) UserID() string {
	return ""
}

func (e *AppCollaboratorPromotedPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredBySiteAdmin
}

func (e *AppCollaboratorPromotedPayload) FillContext(_ *event.Context) {}

func (e *AppCollaboratorPromotedPayload) ForHook() bool {
	return false
}

func (e *AppCollaboratorPromotedPayload) ForAudit() bool {
	return true
}

func (e *AppCollaboratorPromotedPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AppCollaboratorPromotedPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AppCollaboratorPromotedPayload{}
