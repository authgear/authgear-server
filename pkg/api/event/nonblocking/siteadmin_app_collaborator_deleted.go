package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	SiteAdminAppCollaboratorDeleted event.Type = "site_admin.app.collaborator.deleted"
)

type SiteAdminAppCollaboratorDeletedEventPayload struct {
	AppID                 string `json:"app_id"`
	CollaboratorID        string `json:"collaborator_id"`
	CollaboratorUserID    string `json:"user_id"`
	CollaboratorUserEmail string `json:"user_email"`
}

func (e *SiteAdminAppCollaboratorDeletedEventPayload) NonBlockingEventType() event.Type {
	return SiteAdminAppCollaboratorDeleted
}

func (e *SiteAdminAppCollaboratorDeletedEventPayload) UserID() string {
	return ""
}

func (e *SiteAdminAppCollaboratorDeletedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredBySiteAdmin
}

func (e *SiteAdminAppCollaboratorDeletedEventPayload) FillContext(_ *event.Context) {}

func (e *SiteAdminAppCollaboratorDeletedEventPayload) ForHook() bool {
	return false
}

func (e *SiteAdminAppCollaboratorDeletedEventPayload) ForAudit() bool {
	return true
}

func (e *SiteAdminAppCollaboratorDeletedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *SiteAdminAppCollaboratorDeletedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &SiteAdminAppCollaboratorDeletedEventPayload{}
