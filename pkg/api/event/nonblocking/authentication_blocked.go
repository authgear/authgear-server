package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AuthenticationBlocked event.Type = "authentication.blocked"
)

type AuthenticationBlockedEventPayload struct {
	User  *model.User         `json:"user"`
	Error *apierrors.APIError `json:"error"`
}

func (e *AuthenticationBlockedEventPayload) NonBlockingEventType() event.Type {
	return AuthenticationBlocked
}

func (e *AuthenticationBlockedEventPayload) UserID() string {
	if e.User != nil {
		return e.User.ID
	}
	return ""
}

func (e *AuthenticationBlockedEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *AuthenticationBlockedEventPayload) FillContext(ctx *event.Context) {
	userID := e.UserID()
	ctx.UserID = &userID
}

func (e *AuthenticationBlockedEventPayload) ForHook() bool {
	return false
}

func (e *AuthenticationBlockedEventPayload) ForAudit() bool {
	return true
}

func (e *AuthenticationBlockedEventPayload) RequireReindexUserIDs() []string {
	return nil
}

func (e *AuthenticationBlockedEventPayload) DeletedUserIDs() []string {
	return nil
}

var _ event.NonBlockingPayload = &AuthenticationBlockedEventPayload{}
