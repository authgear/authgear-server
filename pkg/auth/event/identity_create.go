package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeIdentityCreate Type = "before_identity_create"
	AfterIdentityCreate  Type = "after_identity_create"
)

type IdentityCreateEvent struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
}

func (IdentityCreateEvent) BeforeEventType() Type {
	return BeforeIdentityCreate
}

func (IdentityCreateEvent) AfterEventType() Type {
	return AfterIdentityCreate
}

func (event IdentityCreateEvent) WithMutationsApplied(mutations Mutations) UserAwarePayload {
	user := event.User
	mutations.ApplyToUser(&user)
	return IdentityCreateEvent{
		User:     user,
		Identity: event.Identity,
	}
}

func (event IdentityCreateEvent) UserID() string {
	return event.User.ID
}
