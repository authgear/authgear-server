package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeSessionDelete Type = "before_session_delete"
	AfterSessionDelete  Type = "after_session_delete"
)

type SessionDeleteReason string

const (
	SessionDeleteReasonLogout = "logout"
)

type SessionDeleteEvent struct {
	Reason   SessionDeleteReason `json:"reason"`
	User     *model.User         `json:"user"`
	Identity *model.Identity     `json:"identity"`
}

func (SessionDeleteEvent) Version() int32 {
	return 1
}

func (SessionDeleteEvent) BeforeEventType() Type {
	return BeforeSessionDelete
}

func (SessionDeleteEvent) AfterEventType() Type {
	return AfterSessionDelete
}
