// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package response

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

// User is the unify way of returning a AuthInfo with LoginID to SDK
type User struct {
	UserID      string            `json:"user_id,omitempty"`
	LoginIDs    map[string]string `json:"login_ids,omitempty"`
	Metadata    userprofile.Data  `json:"metadata"`
	LastLoginAt *time.Time        `json:"last_login_at,omitempty"`
	LastSeenAt  *time.Time        `json:"last_seen_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	CreatedBy   string            `json:"created_by"`
	UpdatedAt   time.Time         `json:"updated_at"`
	UpdatedBy   string            `json:"updated_by"`
	Verified    bool              `json:"verified"`
	VerifyInfo  map[string]bool   `json:"verify_info"`
}

type UserFactory struct {
	PasswordAuthProvider password.Provider
}

func (u UserFactory) NewUser(authInfo authinfo.AuthInfo, userProfile userprofile.UserProfile) User {
	var lastLoginAt *time.Time

	var loginIDs map[string]string
	if u.PasswordAuthProvider != nil {
		if principals, err := u.PasswordAuthProvider.GetPrincipalsByUserID(authInfo.ID); err == nil {
			loginIDs = password.PrincipalsToLoginIDs(principals)
		}
	}

	return User{
		UserID:      authInfo.ID,
		LoginIDs:    loginIDs,
		Metadata:    userProfile.Data,
		LastLoginAt: lastLoginAt,
		LastSeenAt:  authInfo.LastSeenAt,
		CreatedAt:   userProfile.CreatedAt,
		CreatedBy:   userProfile.CreatedBy,
		UpdatedAt:   userProfile.UpdatedAt,
		UpdatedBy:   userProfile.UpdatedBy,
		Verified:    authInfo.Verified,
		VerifyInfo:  authInfo.VerifyInfo,
	}
}

type AuthResponse struct {
	User
	AccessToken string `json:"access_token,omitempty"`
}

type AuthResponseFactory struct {
	PasswordAuthProvider password.Provider
}

func (a AuthResponseFactory) NewAuthResponse(authInfo authinfo.AuthInfo, userProfile userprofile.UserProfile, accessToken string) AuthResponse {
	userFactory := UserFactory{
		PasswordAuthProvider: a.PasswordAuthProvider,
	}

	return AuthResponse{
		User:        userFactory.NewUser(authInfo, userProfile),
		AccessToken: accessToken,
	}
}

func NewAuthResponseByUser(user User, accessToken string) AuthResponse {
	return AuthResponse{
		User:        user,
		AccessToken: accessToken,
	}
}
