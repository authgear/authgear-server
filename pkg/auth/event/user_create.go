package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeUserCreate Type = "before_user_create"
	AfterUserCreate  Type = "after_user_create"
)

type UserCreateEvent struct {
	User       model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
}

func (UserCreateEvent) BeforeEventType() Type {
	return BeforeUserCreate
}

func (UserCreateEvent) AfterEventType() Type {
	return AfterUserCreate
}

func (event UserCreateEvent) WithMutationsApplied(mutations Mutations) UserAwarePayload {
	user := event.User
	mutations.ApplyToUser(&user)
	return UserCreateEvent{
		User:       user,
		Identities: event.Identities,
	}
}

func (event UserCreateEvent) UserID() string {
	return event.User.ID
}
