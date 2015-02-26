package oddb

import (
	"github.com/twinj/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthInfo represents the dictionary of authMethod => authData.
//
// For example, a UserInfo connected with a Facebook account might
// look like this:
//
//   {
//     "facebook": {
//       "accessToken": "someAccessToken",
//       "expiredAt": "2015-02-26T20:05:48",
//       "facebookID": "46709394"
//     }
//   }
type AuthInfo map[string]interface{}

// UserInfo contains a user's information for authentication purpose
type UserInfo struct {
	ID             string   `json:"id"`
	Email          string   `json:"email"`
	HashedPassword []byte   `json:"password"`
	Auth           AuthInfo `json:"auth,omitempty"` // auth data for alternative methods
}

// NewUserInfo returns a new UserInfo with specified email and
// password with generated UUID4 ID
func NewUserInfo(email string, password string) UserInfo {
	info := UserInfo{
		ID:    uuid.NewV4().String(),
		Email: email,
	}
	info.SetPassword(password)

	return info
}

// SetPassword sets the HashedPassword with the password specified
func (info *UserInfo) SetPassword(password string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("userinfo: Failed to hash password")
	}

	info.HashedPassword = hashedPassword
}

// IsSamePassword determines whether the specified password is the same
// password as where the HashedPassword is generated from
func (info UserInfo) IsSamePassword(password string) bool {
	return bcrypt.CompareHashAndPassword(info.HashedPassword, []byte(password)) == nil
}
