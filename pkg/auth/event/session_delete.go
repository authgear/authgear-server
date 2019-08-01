package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeSessionDelete Type = "before_session_delete"
	AfterSessionDelete  Type = "after_session_delete"
)

type SessionDeleteReason string

const (
	SessionDeleteReasonLogout = "logout"
)

type SessionDeleteEvent struct {
	Reason   SessionDeleteReason `json:"reason"`
	User     model.User          `json:"user"`
	Identity model.Identity      `json:"identity"`
}

func (SessionDeleteEvent) BeforeEventType() Type {
	return BeforeSessionDelete
}

func (SessionDeleteEvent) AfterEventType() Type {
	return AfterSessionDelete
}

func (event SessionDeleteEvent) WithMutationsApplied(mutations Mutations) UserAwarePayload {
	user := event.User
	mutations.ApplyToUser(&user)
	return SessionDeleteEvent{
		Reason:   event.Reason,
		User:     user,
		Identity: event.Identity,
	}
}

func (event SessionDeleteEvent) UserID() string {
	return event.User.ID
}
