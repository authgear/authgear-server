package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeSessionDelete Type = "before_session_delete"
	AfterSessionDelete  Type = "after_session_delete"
)

const SessionDeleteEventVersion int32 = 1

type SessionDeleteReason string

const (
	SessionDeleteReasonLogout = "logout"
)

type SessionDeleteEvent struct {
	Reason   SessionDeleteReason `json:"reason"`
	User     *model.User         `json:"user"`
	Identity *model.Identity     `json:"identity"`
}
