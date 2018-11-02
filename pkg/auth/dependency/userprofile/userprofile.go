package userprofile

import (
	"time"
)

var (
	timeNow = func() time.Time { return time.Now().UTC() }
)

type Store interface {
	CreateUserProfile(userID string, userProfile map[string]interface{}) error
	GetUserProfile(userID string, userProfile *map[string]interface{}) error
}
