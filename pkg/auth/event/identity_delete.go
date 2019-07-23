package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeIdentityDelete Type = "before_identity_delete"
	AfterIdentityDelete  Type = "after_identity_delete"
)

const IdentityDeleteEventVersion int32 = 1

type IdentityDeleteEvent struct {
	User     *model.User     `json:"user"`
	Identity *model.Identity `json:"identity"`
}
