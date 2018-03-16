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

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/recordutil"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

var (
	uuidNew = uuid.New
	timeNow = func() time.Time { return time.Now().UTC() }
)

// AuthResponse is the unify way of returing a AuthInfo with AuthData to SDK
type AuthResponse struct {
	UserID      string              `json:"user_id,omitempty"`
	Profile     *skyconv.JSONRecord `json:"profile"`
	Roles       []string            `json:"roles,omitempty"`
	AccessToken string              `json:"access_token,omitempty"`
	LastLoginAt *time.Time          `json:"last_login_at,omitempty"`
	LastSeenAt  *time.Time          `json:"last_seen_at,omitempty"`
}

type AuthResponseFactory struct {
	AssetStore asset.Store
	Conn       skydb.Conn
}

func (f AuthResponseFactory) NewAuthResponse(info skydb.AuthInfo, user skydb.Record, accessToken string, hasMasterKey bool) (AuthResponse, error) {
	var jsonUser *skyconv.JSONRecord
	var lastLoginAt *time.Time

	if user.ID.Type != "" {
		filter, err := recordutil.NewRecordResultFilter(f.Conn, f.AssetStore, &info, hasMasterKey)
		if err != nil {
			return AuthResponse{}, err
		}

		jsonUser = filter.JSONResult(&user)
		if lastLoginAtTime, ok := user.Get(UserRecordLastLoginAtKey).(time.Time); ok {
			lastLoginAt = &lastLoginAtTime
		}
	}

	return AuthResponse{
		UserID:      info.ID,
		Profile:     jsonUser,
		Roles:       info.Roles,
		AccessToken: accessToken,
		LastLoginAt: lastLoginAt,
		LastSeenAt:  info.LastSeenAt,
	}, nil
}
