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
	// DemotedEditors lists every collaborator demoted by this promotion --
	// normally at most one, but the schema has no constraint preventing an
	// app from having more than one owner-role collaborator before this
	// promotion (e.g. a direct DB edit can create that state), and none of
	// them should be silently dropped from the audit trail.
	DemotedEditors []DemotedEditor `json:"demoted_editors,omitempty"`
}

type DemotedEditor struct {
	CollaboratorID string `json:"collaborator_id"`
	UserID         string `json:"user_id"`
	UserEmail      string `json:"user_email"`
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
