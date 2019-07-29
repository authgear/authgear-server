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

func (event UserSyncEvent) ApplyingMutations(mutations Mutations) UserAwarePayload {
	return UserSyncEvent{
		User: mutations.ApplyingToUser(event.User),
	}
}

func (event UserSyncEvent) UserID() string {
	return event.User.ID
}
