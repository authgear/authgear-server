package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeSessionCreate Type = "before_session_create"
	AfterSessionCreate  Type = "after_session_create"
)

type SessionCreateReason string

const (
	SessionCreateReasonSignup = "signup"
	SessionCreateReasonLogin  = "login"
)

type SessionCreateEvent struct {
	Reason   SessionCreateReason `json:"reason"`
	User     model.User          `json:"user"`
	Identity model.Identity      `json:"identity"`
}

func (SessionCreateEvent) BeforeEventType() Type {
	return BeforeSessionCreate
}

func (SessionCreateEvent) AfterEventType() Type {
	return AfterSessionCreate
}

func (event SessionCreateEvent) WithMutationsApplied(mutations Mutations) UserAwarePayload {
	user := event.User
	mutations.ApplyToUser(&user)
	return SessionCreateEvent{
		Reason:   event.Reason,
		User:     user,
		Identity: event.Identity,
	}
}

func (event SessionCreateEvent) UserID() string {
	return event.User.ID
}
