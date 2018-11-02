package userprofile

import (
	"time"
)

var (
	timeNow = func() time.Time { return time.Now().UTC() }
)

// Data refers the profile info of a user,
// like username, email, age, phone number
type Data map[string]interface{}

// UserProfile refers the serialized user profile,
// it is a serialized record object
type UserProfile map[string]interface{}

type Store interface {
	CreateUserProfile(userID string, data Data) (UserProfile, error)
	GetUserProfile(userID string) (UserProfile, error)
}
