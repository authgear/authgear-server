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

func (IdentityCreateEvent) Version() int32 {
	return 1
}

func (IdentityCreateEvent) BeforeEventType() Type {
	return BeforeIdentityCreate
}

func (IdentityCreateEvent) AfterEventType() Type {
	return AfterIdentityCreate
}

func (event IdentityCreateEvent) ApplyingMutations(mutations Mutations) UserAwarePayload {
	return IdentityCreateEvent{
		User:     mutations.ApplyingToUser(event.User),
		Identity: event.Identity,
	}
}

func (event IdentityCreateEvent) UserID() string {
	return event.User.ID
}
