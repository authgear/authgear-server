package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeIdentityDelete Type = "before_identity_delete"
	AfterIdentityDelete  Type = "after_identity_delete"
)

type IdentityDeleteEvent struct {
	User     *model.User     `json:"user"`
	Identity *model.Identity `json:"identity"`
}

func (IdentityDeleteEvent) Version() int32 {
	return 1
}
