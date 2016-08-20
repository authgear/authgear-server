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

package preprocessor

import (
	"net/http"
	"strings"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

var log = logging.LoggerEntry("preprocessor")

type InjectUserIfPresent struct {
}

func isTokenStillValid(token router.AccessToken, userInfo skydb.UserInfo) bool {
	if userInfo.TokenValidSince == nil {
		return true
	}
	tokenValidSince := *userInfo.TokenValidSince

	// Not all types of access token support this field. The token is
	// still considered if it does not have an issue time.
	if token.IssuedAt().IsZero() {
		return true
	}

	// Due to precision, the issue time of the token can be before
	// UserInfo.TokenValidSince. We consider the token still valid
	// if the token is issued within 1 second before tokenValidSince.
	return token.IssuedAt().After(tokenValidSince.Add(-1 * time.Second))
}

func (p InjectUserIfPresent) Preprocess(payload *router.Payload, response *router.Response) int {
	if payload.UserInfoID == "" {
		log.Debugln("injectUser: empty UserInfoID, skipping")
		return http.StatusOK
	}

	conn := payload.DBConn
	userinfo := skydb.UserInfo{}
	if err := conn.GetUser(payload.UserInfoID, &userinfo); err != nil {
		log.Errorf("Cannot find UserInfo.ID = %#v\n", payload.UserInfoID)
		response.Err = skyerr.NewError(skyerr.UnexpectedUserInfoNotFound, err.Error())
		return http.StatusInternalServerError
	}

	// If an access token exists checks if the access token has an IssuedAt
	// time that is later than the user's TokenValidSince time. This
	// allows user to invalidate previously issued access token.
	if payload.AccessToken != nil && !isTokenStillValid(payload.AccessToken, userinfo) {
		response.Err = skyerr.NewError(skyerr.AccessTokenNotAccepted, "token does not exist or it has expired")
		return http.StatusUnauthorized
	}

	payload.UserInfo = &userinfo

	return http.StatusOK
}

type InjectDatabase struct {
}

func (p InjectDatabase) Preprocess(payload *router.Payload, response *router.Response) int {
	conn := payload.DBConn

	databaseID, ok := payload.Data["database_id"].(string)
	if !ok || databaseID == "" {
		databaseID = "_public"
	}

	switch databaseID {
	case "_private":
		if payload.UserInfo != nil {
			payload.Database = conn.PrivateDB(payload.UserInfo.ID)
		} else {
			response.Err = skyerr.NewError(skyerr.NotAuthenticated, "Authentication is needed for private DB access")
			return http.StatusUnauthorized
		}
	case "_public":
		payload.Database = conn.PublicDB()
	case "_union":
		if !payload.HasMasterKey() {
			response.Err = skyerr.NewError(skyerr.NotAuthenticated, "Master key is needed for union DB access")
			return http.StatusUnauthorized
		}
		payload.Database = conn.UnionDB()
	default:
		if strings.HasPrefix(databaseID, "_") {
			response.Err = skyerr.NewInvalidArgument("invalid database ID", []string{"database_id"})
			return http.StatusBadRequest
		} else if payload.HasMasterKey() {
			payload.Database = conn.PrivateDB(databaseID)
		} else if payload.UserInfo != nil && databaseID == payload.UserInfo.ID {
			payload.Database = conn.PrivateDB(databaseID)
		} else {
			response.Err = skyerr.NewError(skyerr.PermissionDenied, "The selected DB cannot be accessed because permission is denied")
			return http.StatusForbidden
		}
	}

	return http.StatusOK
}

type InjectPublicDatabase struct {
}

func (p InjectPublicDatabase) Preprocess(payload *router.Payload, response *router.Response) int {
	conn := payload.DBConn
	payload.Database = conn.PublicDB()
	return http.StatusOK
}

type RequireUserForWrite struct {
}

func (p RequireUserForWrite) Preprocess(payload *router.Payload, response *router.Response) int {
	if payload.UserInfo == nil {
		response.Err = skyerr.NewError(skyerr.PermissionDenied, "write is not allowed")
		return http.StatusUnauthorized
	}

	return http.StatusOK
}
