package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeSessionCreate Type = "before_session_create"
	AfterSessionCreate  Type = "after_session_create"
)

const SessionCreateEventVersion int32 = 1

type SessionCreateReason string

const (
	SessionCreateReasonSignup = "signup"
	SessionCreateReasonLogin  = "login"
)

type SessionCreateEvent struct {
	Reason   SessionCreateReason `json:"reason"`
	User     *model.User         `json:"user"`
	Identity *model.Identity     `json:"identity"`
}
