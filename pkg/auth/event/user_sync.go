package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	UserSync Type = "user_sync"
)

type UserSyncEvent struct {
	User model.User `json:"user"`
}

func (UserSyncEvent) EventType() Type {
	return UserSync
}

func (event UserSyncEvent) WithMutationsApplied(mutations Mutations) UserAwarePayload {
	user := event.User
	mutations.ApplyToUser(&user)
	return UserSyncEvent{
		User: user,
	}
}

func (event UserSyncEvent) UserID() string {
	return event.User.ID
}
