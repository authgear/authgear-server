package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	UserSync Type = "user_sync"
)

const UserSyncEventVersion int32 = 1

type UserSyncEvent struct {
	User *model.User `json:"user"`
}

func (UserSyncEvent) Version() int32 {
	return 1
}

func (UserSyncEvent) EventType() Type {
	return UserSync
}
