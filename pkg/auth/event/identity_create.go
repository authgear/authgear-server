package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeIdentityCreate Type = "before_identity_create"
	AfterIdentityCreate  Type = "after_identity_create"
)

const IdentityCreateEventVersion int32 = 1

type IdentityCreateEvent struct {
	User     *model.User     `json:"user"`
	Identity *model.Identity `json:"identity"`
}
