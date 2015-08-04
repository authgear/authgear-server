package oddb

import (
	"github.com/oursky/ourd/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthInfo represents the dictionary of authenticated principal ID => authData.
//
// For example, a UserInfo connected with a Facebook account might
// look like this:
//
//   {
//     "com.facebook:46709394": {
//       "accessToken": "someAccessToken",
//       "expiredAt": "2015-02-26T20:05:48",
//       "facebookID": "46709394"
//     }
//   }
//
// It is assumed that the Facebook AuthProvider has "com.facebook" as
// provider name and "46709394" as the authenticated Facebook account ID.
type AuthInfo map[string]map[string]interface{}

// UserInfo contains a user's information for authentication purpose
type UserInfo struct {
	ID             string   `json:"_id"`
	Email          string   `json:"email,omitempty"`
	HashedPassword []byte   `json:"password,omitempty"`
	Auth           AuthInfo `json:"auth,omitempty"` // auth data for alternative methods
}

// NewUserInfo returns a new UserInfo with specified email and
// password with generated UUID4 ID
func NewUserInfo(id, email, password string) UserInfo {
	if id == "" {
		id = uuid.New()
	}

	info := UserInfo{
		ID:    id,
		Email: email,
	}
	info.SetPassword(password)

	return info
}

// NewAnonymousUserInfo returns an anonymous UserInfo, which has
// no Email and Password.
func NewAnonymousUserInfo() UserInfo {
	return UserInfo{
		ID: uuid.New(),
	}
}

// NewProvidedAuthUserInfo returns an UserInfo provided by a AuthProvider,
// which has no Email and Password.
func NewProvidedAuthUserInfo(principalID string, authData map[string]interface{}) UserInfo {
	return UserInfo{
		ID: uuid.New(),
		Auth: AuthInfo(map[string]map[string]interface{}{
			principalID: authData,
		}),
	}
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

// SetProvidedAuthData sets the auth data to the specified principal.
func (info *UserInfo) SetProvidedAuthData(principalID string, authData map[string]interface{}) {
	if info.Auth == nil {
		info.Auth = make(map[string]map[string]interface{})
	}
	info.Auth[principalID] = authData
}

// GetProvidedAuthData gets the auth data for the specified principal.
func (info *UserInfo) GetProvidedAuthData(principalID string) map[string]interface{} {
	if info.Auth == nil {
		return nil
	}
	value, _ := info.Auth[principalID]
	return value
}

// RemoveProvidedAuthData remove the auth data for the specified principal.
func (info *UserInfo) RemoveProvidedAuthData(principalID string) {
	if info.Auth != nil {
		delete(info.Auth, principalID)
	}
}
