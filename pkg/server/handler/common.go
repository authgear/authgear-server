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

package handler

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

var (
	timeNowUTC = func() time.Time { return time.Now().UTC() }
	uuidNew    = uuid.New
	timeNow    = timeNowUTC
)

// AuthResponse is the unify way of returing a AuthInfo to SDK
type AuthResponse struct {
	UserID      string     `json:"user_id,omitempty"`
	Username    string     `json:"username,omitempty"`
	Email       string     `json:"email,omitempty"`
	Roles       []string   `json:"roles,omitempty"`
	AccessToken string     `json:"access_token,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`
}

func NewAuthResponse(info skydb.AuthInfo, accessToken string) AuthResponse {
	return AuthResponse{
		UserID:      info.ID,
		Username:    info.Username,
		Email:       info.Email,
		Roles:       info.Roles,
		AccessToken: accessToken,
		LastLoginAt: info.LastLoginAt,
		LastSeenAt:  info.LastSeenAt,
	}
}
