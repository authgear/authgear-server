package nonblocking

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserCreatedUserSignup         event.Type = "user.created.user_signup"
	UserCreatedAdminAPICreateUser event.Type = "user.created.admin_api_create_user"
)

type UserCreatedUserSignupEvent struct {
	User       model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
}

func (e *UserCreatedUserSignupEvent) NonBlockingEventType() event.Type {
	return UserCreatedUserSignup
}

func (e *UserCreatedUserSignupEvent) UserID() string {
	return e.User.ID
}

type UserCreatedAdminAPICreateUserEvent struct {
	User       model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
}

func (e *UserCreatedAdminAPICreateUserEvent) NonBlockingEventType() event.Type {
	return UserCreatedAdminAPICreateUser
}

func (e *UserCreatedAdminAPICreateUserEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &UserCreatedUserSignupEvent{}
var _ event.NonBlockingPayload = &UserCreatedAdminAPICreateUserEvent{}
