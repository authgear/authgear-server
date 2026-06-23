package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	SiteAdminAppCollaboratorAdded event.Type = "site_admin.app.collaborator.added"
)

type SiteAdminAppCollaboratorAddedEventPayload struct {
	AppID              string `json:"app_id"`
	CollaboratorID     string `json:"collaborator_id"`
	CollaboratorUserID string `json:"user_id"`
	UserEmail          string `json:"user_email"`
	Role               string `json:"role"`
}

func (e *SiteAdminAppCollaboratorAddedEventPayload) NonBlockingEventType() event.Type {
	return SiteAdminAppCollaboratorAdded
}

func (e *SiteAdminAppCollaboratorAddedEventPayload) UserID() string {
	return ""
}

func (e *SiteAdminAppCollaboratorAddedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredBySiteAdmin
}

func (e *SiteAdminAppCollaboratorAddedEventPayload) FillContext(_ *event.Context) {}

func (e *SiteAdminAppCollaboratorAddedEventPayload) ForHook() bool {
	return false
}

func (e *SiteAdminAppCollaboratorAddedEventPayload) ForAudit() bool {
	return true
}

func (e *SiteAdminAppCollaboratorAddedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *SiteAdminAppCollaboratorAddedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &SiteAdminAppCollaboratorAddedEventPayload{}
