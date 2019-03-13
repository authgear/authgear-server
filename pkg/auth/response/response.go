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

	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

// AuthResponse is the unify way of returing a AuthInfo with AuthData to SDK
type AuthResponse struct {
	UserID      string                  `json:"user_id,omitempty"`
	Profile     userprofile.UserProfile `json:"metadata,omitempty"`
	Roles       []string                `json:"roles,omitempty"`
	AccessToken string                  `json:"access_token,omitempty"`
	LastLoginAt *time.Time              `json:"last_login_at,omitempty"`
	LastSeenAt  *time.Time              `json:"last_seen_at,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	CreatedBy   string                  `json:"created_by"`
	UpdatedAt   time.Time               `json:"updated_at"`
	UpdatedBy   string                  `json:"updated_by"`
	Verified    bool                    `json:"verified"`
	VerifyInfo  map[string]bool         `json:"verify_info"`
}

func NewAuthResponse(authInfo authinfo.AuthInfo, userProfile userprofile.UserProfile, accessToken string) AuthResponse {
	var lastLoginAt *time.Time

	return AuthResponse{
		UserID:      authInfo.ID,
		Profile:     userProfile,
		Roles:       authInfo.Roles,
		AccessToken: accessToken,
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
