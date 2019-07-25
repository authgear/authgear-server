package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeIdentityDelete Type = "before_identity_delete"
	AfterIdentityDelete  Type = "after_identity_delete"
)

type IdentityDeleteEvent struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
}

func (IdentityDeleteEvent) Version() int32 {
	return 1
}

func (IdentityDeleteEvent) BeforeEventType() Type {
	return BeforeIdentityDelete
}

func (IdentityDeleteEvent) AfterEventType() Type {
	return AfterIdentityDelete
}

func (event IdentityDeleteEvent) ApplyingMutations(mutations Mutations) UserAwarePayload {
	return IdentityDeleteEvent{
		User:     mutations.ApplyingToUser(event.User),
		Identity: event.Identity,
	}
}

func (event IdentityDeleteEvent) UserID() string {
	return event.User.ID
}
