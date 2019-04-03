package userprofile

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

var (
	timeNow = func() time.Time { return time.Now().UTC() }
)

// Data refers the profile info of a user,
// like username, email, age, phone number
type Data map[string]interface{}

// UserProfile refers user profile data type
type UserProfile struct {
	ID        string
	CreatedAt time.Time
	CreatedBy string
	UpdatedAt time.Time
	UpdatedBy string
	Data
}

func (in UserProfile) MergeLoginIDs(loginIDs map[string]string) (out UserProfile) {
	out = in
	out.Data = make(map[string]interface{})
	for k, v := range in.Data {
		out.Data[k] = v
	}
	for k, v := range loginIDs {
		out.Data[k] = v
	}
	return out
}

type Store interface {
	CreateUserProfile(userID string, authInfo *authinfo.AuthInfo, data Data) (UserProfile, error)
	GetUserProfile(userID string) (UserProfile, error)
	UpdateUserProfile(userID string, authInfo *authinfo.AuthInfo, data Data) (UserProfile, error)
}
