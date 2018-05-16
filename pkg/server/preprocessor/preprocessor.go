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
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/recordutil"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

var timeNow = func() time.Time { return time.Now().UTC() }

// InjectAuth preprocessor checks the auth_id in the request and get the auth
// object from the database. It can be configured to
type InjectAuth struct {
	PwExpiryDays int
	Required     bool
}

func isTokenStillValid(token router.AccessToken, authInfo skydb.AuthInfo) bool {
	if authInfo.TokenValidSince == nil {
		return true
	}
	tokenValidSince := *authInfo.TokenValidSince

	// Not all types of access token support this field. The token is
	// still considered if it does not have an issue time.
	if token.IssuedAt().IsZero() {
		return true
	}

	// Due to precision, the issue time of the token can be before
	// AuthInfo.TokenValidSince. We consider the token still valid
	// if the token is issued within 1 second before tokenValidSince.
	return token.IssuedAt().After(tokenValidSince.Add(-1 * time.Second))
}

func (p InjectAuth) Preprocess(payload *router.Payload, response *router.Response) (status int) {
	// If the payload does not already have auth_id, and if the request
	// is authenticated with master key, assume the user is _god.
	if payload.AuthInfoID == "" && payload.HasMasterKey() {
		payload.AuthInfoID = "_god"
		payload.Context = context.WithValue(payload.Context, router.UserIDContextKey, "_god")
	}

	authinfo := skydb.AuthInfo{}
	var lastError skyerr.Error

	defer func() {
		// If auth info is required but missing, return an error instead of
		// allowing the request to continue.
		// If there is an existing error, use that. If there is none, return
		// a generic error.
		if p.Required && payload.AuthInfo == nil {
			if lastError == nil {
				lastError = skyerr.NewError(
					skyerr.NotAuthenticated,
					"Authentication is required for this action, please login.",
				)
			}
			response.Err = lastError
			if status == 0 {
				status = http.StatusUnauthorized
			}
			return
		}

		if status == http.StatusInternalServerError {
			response.Err = lastError
			return
		}

		// If auth info is not requird, any problems with the user is silently
		// ignored.
		status = http.StatusOK
	}()

	// Query database to get auth info
	// If an error occurred at this stage, Internal Server Error is returned.
	if payload.AuthInfoID == "" {
		return
	}

	lastError, status = p.fetchOrCreateAuth(payload, &authinfo)
	if status != 0 {
		return
	}

	lastError, status = p.checkDisabledStatus(payload, &authinfo)
	if status != 0 {
		return
	}

	// If an access token exists checks if the access token has an IssuedAt
	// time that is later than the user's TokenValidSince time. This
	// allows user to invalidate previously issued access token.
	if payload.AccessToken != nil && !isTokenStillValid(payload.AccessToken, authinfo) {
		lastError = skyerr.NewError(skyerr.AccessTokenNotAccepted, "token does not exist or it has expired")
		status = http.StatusUnauthorized
		return
	}

	// Check if password is expired according to policy
	if authinfo.IsPasswordExpired(p.PwExpiryDays) {
		lastError = audit.MakePasswordError(audit.PasswordExpired, "password expired", nil)
		status = http.StatusUnauthorized
		return
	}

	payload.AuthInfo = &authinfo
	return
}

func (p InjectAuth) fetchOrCreateAuth(payload *router.Payload, authInfo *skydb.AuthInfo) (skyerr.Error, int) {
	logger := logging.CreateLogger(payload.Context, "preprocessor")
	var err error
	err = payload.DBConn.GetAuth(payload.AuthInfoID, authInfo)
	if err == skydb.ErrUserNotFound && payload.HasMasterKey() {
		*authInfo = skydb.AuthInfo{
			ID: payload.AuthInfoID,
		}
		err = payload.DBConn.CreateAuth(authInfo)
		if err == skydb.ErrUserDuplicated {
			// user already exists, error can be ignored
			err = nil
		}
	}

	if err != nil {
		logger.Errorf("Cannot find AuthInfo.ID = %#v\n", payload.AuthInfoID)
		return skyerr.NewError(skyerr.UnexpectedAuthInfoNotFound, err.Error()), http.StatusInternalServerError
	}
	return nil, 0
}

func (p InjectAuth) checkDisabledStatus(payload *router.Payload, authInfo *skydb.AuthInfo) (skyerr.Error, int) {
	logger := logging.CreateLogger(payload.Context, "preprocessor")
	// Check if user is disabled
	if authInfo.IsDisabled() {
		logger.Info("User is disabled")
		info := map[string]interface{}{}
		if authInfo.DisabledExpiry != nil {
			info["expiry"] = authInfo.DisabledExpiry.Format(time.RFC3339)
		}
		if authInfo.DisabledMessage != "" {
			info["message"] = authInfo.DisabledMessage
		}
		return skyerr.NewErrorWithInfo(skyerr.UserDisabled, "user is disabled", info), http.StatusForbidden
	}
	authInfo.RefreshDisabledStatus()
	return nil, 0
}

// InjectUser injects a user record to the payload
//
// An AuthInfo must be injected before this, if it is not found, the preprocessor
// would just skip the injection
//
// If AuthInfo is injected but a user record is not found, the preprocessor would
// create a new user record and inject it to the payload
type InjectUser struct {
	HookRegistry      *hook.Registry `inject:"HookRegistry"`
	AssetStore        asset.Store    `inject:"AssetStore"`
	Required          bool
	CheckVerification bool
}

func (p InjectUser) Preprocess(payload *router.Payload, response *router.Response) int {
	logger := logging.CreateLogger(payload.Context, "preprocessor")
	db := payload.DBConn.PublicDB()

	if payload.User == nil && payload.AuthInfo != nil {
		user := skydb.Record{}
		err := db.Get(skydb.NewRecordID("user", payload.AuthInfo.ID), &user)

		if err == skydb.ErrRecordNotFound {
			user, err = p.createUser(payload)
		}

		if err != nil {
			logger.Error("injectUser: unable to find or create user record", err)
			response.Err = skyerr.NewError(skyerr.UnexpectedUserNotFound, err.Error())
			return http.StatusInternalServerError
		}
		payload.User = &user
	}

	if p.Required && payload.User == nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedUserNotFound, "user not found")
		return http.StatusInternalServerError
	}

	if p.CheckVerification && payload.User != nil {
		if val, ok := payload.User.Data["is_verified"].(bool); !ok || !val {
			response.Err = skyerr.NewError(
				skyerr.VerificationRequired,
				"User is not yet verified",
			)
			return http.StatusForbidden
		}
	}

	return http.StatusOK
}

func (p InjectUser) createUser(payload *router.Payload) (skydb.Record, error) {
	authInfo := payload.AuthInfo
	db := payload.DBConn.PublicDB()
	txDB, ok := db.(skydb.Transactional)
	if !ok {
		return skydb.Record{}, skyerr.NewError(skyerr.NotSupported, "database impl does not support transaction")
	}

	var user *skydb.Record
	txErr := skydb.WithTransaction(txDB, func() error {
		userRecord := skydb.Record{
			ID: skydb.NewRecordID(db.UserRecordType(), authInfo.ID),
		}

		recordReq := recordutil.RecordModifyRequest{
			Db:           db,
			Conn:         payload.DBConn,
			AssetStore:   p.AssetStore,
			HookRegistry: p.HookRegistry,
			Atomic:       true,
			Context:      payload.Context,
			AuthInfo:     authInfo,
			ModifyAt:     timeNow(),
			RecordsToSave: []*skydb.Record{
				&userRecord,
			},
		}

		recordResp := recordutil.RecordModifyResponse{
			ErrMap: map[skydb.RecordID]skyerr.Error{},
		}

		err := recordutil.RecordSaveHandler(&recordReq, &recordResp)
		if err != nil {
			return err
		}

		user = recordResp.SavedRecords[0]
		return nil
	})

	if txErr != nil {
		return skydb.Record{}, txErr
	}

	return *user, nil
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
		if payload.AuthInfo != nil {
			payload.Database = conn.PrivateDB(payload.AuthInfo.ID)
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
		} else if payload.AuthInfo != nil && databaseID == payload.AuthInfo.ID {
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

type RequireAdminOrMasterKey struct {
}

func (p RequireAdminOrMasterKey) Preprocess(payload *router.Payload, response *router.Response) int {
	if payload.HasMasterKey() {
		return http.StatusOK
	}

	if payload.AuthInfo == nil {
		response.Err = skyerr.NewError(
			skyerr.NotAuthenticated,
			"User is required for this action, please login.",
		)
		return http.StatusUnauthorized
	}

	adminRoles, err := payload.DBConn.GetAdminRoles()
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return http.StatusInternalServerError
	}

	if payload.AuthInfo.HasAnyRoles(adminRoles) {
		return http.StatusOK
	}

	response.Err = skyerr.NewError(
		skyerr.PermissionDenied,
		"no permission to perform this action",
	)
	return http.StatusUnauthorized
}

type RequireMasterKey struct {
}

func (p RequireMasterKey) Preprocess(payload *router.Payload, response *router.Response) int {
	if payload.HasMasterKey() == false {
		response.Err = skyerr.NewError(skyerr.PermissionDenied, "no permission to this action")
		return http.StatusUnauthorized
	}

	return http.StatusOK
}

type Null struct {
}

func (p Null) Preprocess(payload *router.Payload, response *router.Response) int {
	return http.StatusOK
}
