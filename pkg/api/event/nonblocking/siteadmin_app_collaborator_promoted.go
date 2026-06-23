package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	SiteAdminAppCollaboratorPromoted event.Type = "site_admin.app.collaborator.promoted"
)

type SiteAdminAppCollaboratorPromotedEventPayload struct {
	AppID                  string `json:"app_id"`
	NewOwnerCollaboratorID string `json:"new_owner_collaborator_id"`
	NewOwnerUserID         string `json:"new_owner_user_id"`
	NewOwnerUserEmail      string `json:"new_owner_user_email"`
	// DemotedEditor* fields are omitted when the app had no previous owner.
	DemotedEditorCollaboratorID string `json:"demoted_editor_collaborator_id,omitempty"`
	DemotedEditorUserID         string `json:"demoted_editor_user_id,omitempty"`
	DemotedEditorUserEmail      string `json:"demoted_editor_user_email,omitempty"`
}

func (e *SiteAdminAppCollaboratorPromotedEventPayload) NonBlockingEventType() event.Type {
	return SiteAdminAppCollaboratorPromoted
}

func (e *SiteAdminAppCollaboratorPromotedEventPayload) UserID() string {
	return ""
}

func (e *SiteAdminAppCollaboratorPromotedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredBySiteAdmin
}

func (e *SiteAdminAppCollaboratorPromotedEventPayload) FillContext(_ *event.Context) {}

func (e *SiteAdminAppCollaboratorPromotedEventPayload) ForHook() bool {
	return false
}

func (e *SiteAdminAppCollaboratorPromotedEventPayload) ForAudit() bool {
	return true
}

func (e *SiteAdminAppCollaboratorPromotedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *SiteAdminAppCollaboratorPromotedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &SiteAdminAppCollaboratorPromotedEventPayload{}
