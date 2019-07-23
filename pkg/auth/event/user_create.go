package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeUserCreate Type = "before_user_create"
	AfterUserCreate  Type = "after_user_create"
)

const UserCreateEventVersion int32 = 1

type UserCreateEvent struct {
	User       *model.User       `json:"user"`
	Identities []*model.Identity `json:"identities"`
}
