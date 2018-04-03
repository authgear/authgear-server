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
	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/plugin/provider"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type ssoCustomTokenClaims struct {
	RawProfile	map[string]interface{} `json:"skyprofile"`
	Profile 	skydb.Data 
	jwt.StandardClaims
}

type ssoCustomTokenLoginPayload struct {
	keyFunc     jwt.Keyfunc
	TokenString string               `mapstructure:"token"`
	Token       *jwt.Token           `mapstructure:"-"`
	Claims      ssoCustomTokenClaims `mapstructure:"-"`
}

func (payload *ssoCustomTokenLoginPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	parsedToken, err := jwt.ParseWithClaims(
		payload.TokenString,
		&payload.Claims,
		payload.keyFunc,
	)
	if err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	if err := (*skyconv.MapData)(&payload.Claims.Profile).FromMap(payload.Claims.RawProfile); err != nil {
		return skyerr.NewError(skyerr.InvalidArgument, err.Error())
	}

	payload.Token = parsedToken
	return payload.Validate()
}

func (payload *ssoCustomTokenLoginPayload) Validate() skyerr.Error {
	claims := payload.Claims
	if claims.Subject == "" {
		return skyerr.NewError(
			skyerr.InvalidCredentials,
			"invalid token: subject (sub) not specified",
		)
	}

	if claims.ExpiresAt == 0 {
		return skyerr.NewError(
			skyerr.InvalidCredentials,
			"invalid token: expires at (exp) not specified",
		)
	}

	if claims.IssuedAt == 0 {
		return skyerr.NewError(
			skyerr.InvalidCredentials,
			"invalid token: issued at (iat) not specified",
		)
	}

	if claims.Valid() != nil {
		return skyerr.NewError(
			skyerr.InvalidCredentials,
			"invalid token: token is not valid at this time",
		)
	}

	return nil
}

/*
SSOCustomTokenLoginHandler authenticates the user with a custom token

An external server is responsible for generating the custom token which
contains a Principal ID and a signature. It is required that the token
has issued-at and expired-at claims.

The custom token is signed by a shared secret and encoded in JWT format.

The claims of the custom token is as follows:

    {
      "sub": "id1234567800",
      "iat": 1513316033,
      "exp": 1828676033,
      "skyprofile": {
        "name": "John Doe"
      }
    }

When signing the above claims with the custom token secret `ssosecret` using
HS256 as algorithm, the following JWT token is produced:

	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJpZDEyMzQ1Njc4MDAiLCJpYXQiOjE1MTMzMTYwMzMsImV4cCI6MTgyODY3NjAzMywic2t5cHJvZmlsZSI6eyJuYW1lIjoiSm9obiBEb2UifX0.JRAwXPF4CDWCpMCvemCBPrUAQAXPV9qVWeAYo1vBAqQ

This token can be used to log in to Skygear Server. If there is no user
associated with the Principal ID (the subject/sub claim), a new user is
created.


curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
	"action": "sso:custom_token:login",
	"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJpZDEyMzQ1Njc4MDAiLCJpYXQiOjE1MTMzMTYwMzMsImV4cCI6MTgyODY3NjAzMywic2t5cHJvZmlsZSI6eyJuYW1lIjoiSm9obiBEb2UifX0.JRAwXPF4CDWCpMCvemCBPrUAQAXPV9qVWeAYo1vBAqQ"
}
EOF
*/
type SSOCustomTokenLoginHandler struct {
	CustomTokenSecret string

	TokenStore       authtoken.Store    `inject:"TokenStore"`
	ProviderRegistry *provider.Registry `inject:"ProviderRegistry"`
	HookRegistry     *hook.Registry     `inject:"HookRegistry"`
	AssetStore       asset.Store        `inject:"AssetStore"`
	AuthRecordKeys   [][]string         `inject:"AuthRecordKeys"`
	AccessKey        router.Processor   `preprocessor:"accesskey"`
	DBConn           router.Processor   `preprocessor:"dbconn"`
	InjectPublicDB   router.Processor   `preprocessor:"inject_public_db"`
	PluginReady      router.Processor   `preprocessor:"plugin_ready"`
	preprocessors    []router.Processor
}

func (h *SSOCustomTokenLoginHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
		h.InjectPublicDB,
		h.PluginReady,
	}
}

func (h *SSOCustomTokenLoginHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SSOCustomTokenLoginHandler) Handle(payload *router.Payload, response *router.Response) {
	if h.CustomTokenSecret == "" {
		response.Err = skyerr.NewError(
			skyerr.NotConfigured,
			"login with custom token requires CUSTOM_TOKEN_SECRET config",
		)
		return
	}

	p := &ssoCustomTokenLoginPayload{
		keyFunc: func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, skyerr.NewInvalidArgument("invalid token", []string{"token"})
			}
			return []byte(h.CustomTokenSecret), nil
		},
	}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	if h.TokenStore == nil {
		panic("token store is nil")
	}
	store := h.TokenStore

	info := skydb.AuthInfo{}
	user := skydb.Record{}

	if skyErr = h.handleLogin(payload, p, &info, &user); skyErr != nil {
		response.Err = skyErr
		return
	}

	if err := checkUserIsNotDisabled(&info); err != nil {
		response.Err = err
		return
	}

	// generate access-token
	token, err := store.NewToken(payload.AppName, info.ID)
	if err != nil {
		panic(err)
	}

	if err = store.Put(&token); err != nil {
		panic(err)
	}

	authResponse, err := AuthResponseFactory{
		AssetStore: h.AssetStore,
		Conn:       payload.DBConn,
	}.NewAuthResponse(info, user, token.AccessToken, payload.HasMasterKey())
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	// Populate the activity time to user
	now := timeNow()
	info.LastSeenAt = &now
	if err := payload.DBConn.UpdateAuth(&info); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	// update user record last login time
	user.UpdatedAt = now
	user.UpdaterID = info.ID
	user.Data[UserRecordLastLoginAtKey] = now
	if err := payload.Database.Save(&user); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	response.Result = authResponse
}

func (h *SSOCustomTokenLoginHandler) handleLogin(payload *router.Payload, p *ssoCustomTokenLoginPayload, authinfo *skydb.AuthInfo, user *skydb.Record) skyerr.Error {
	principalID := p.Claims.Subject
	var customTokenInfo skydb.CustomTokenInfo
	createNewUser := false
	createNewCustomToken := false

	if err := payload.DBConn.GetCustomTokenInfo(principalID, &customTokenInfo); err != nil {
		if err != skydb.ErrUserNotFound {
			return skyerr.MakeError(err)
		}

		// Custom token info does not exist. We always create a new user in
		// this case.
		createNewUser = true
		createNewCustomToken = true
		*authinfo = skydb.NewAnonymousAuthInfo()
	}

	if !createNewUser {
		if err := payload.DBConn.GetAuth(customTokenInfo.UserID, authinfo); err != nil {
			if err != skydb.ErrUserNotFound {
				return skyerr.MakeError(err)
			}

			// There is a custom token but the user does not exist.
			// Creating the new user anyway, using the ID in the custom token.
			createNewUser = true
			*authinfo = skydb.AuthInfo{
				ID: customTokenInfo.UserID,
			}
		}
	}

	userRecordContext := &authUserRecordContext{
		DBConn:         payload.DBConn,
		Database:       payload.Database,
		AssetStore:     h.AssetStore,
		HookRegistry:   h.HookRegistry,
		AuthRecordKeys: h.AuthRecordKeys,
		Context:        payload.Context,
	}

	// Create a new AuthInfo if we are creating a new user, otherwise
	// update the AuthInfo.
	if createNewUser {
		userRecordContext.BeforeSaveFunc = func(conn skydb.Conn, info *skydb.AuthInfo) error {
			if err := conn.CreateAuth(info); err != nil {
				if err == skydb.ErrUserDuplicated {
					return errUserDuplicated
				}

				return skyerr.MakeError(err)
			}
			return nil
		}
	} else {
		userRecordContext.BeforeSaveFunc = func(conn skydb.Conn, info *skydb.AuthInfo) error {
			if err := conn.UpdateAuth(info); err != nil {
				return skyerr.MakeError(err)
			}
			return nil
		}
	}

	// Create a new CustomTokenInfo if it doesn't exist.
	if createNewCustomToken {
		userRecordContext.BeforeCommitFunc = func(conn skydb.Conn, user *skydb.Record) error {
			now := timeNow()
			if err := conn.CreateCustomTokenInfo(&skydb.CustomTokenInfo{
				UserID:      user.ID.Key,
				PrincipalID: principalID,
				CreatedAt:   &now,
			}); err != nil {
				return skyerr.MakeError(err)
			}
			return nil
		}
	}

	modifiedUser, err := userRecordContext.execute(
		authinfo,
		skydb.AuthData{},
		p.Claims.Profile,
	)
	if err != nil {
		return skyerr.MakeError(err)
	}

	*user = *modifiedUser
	return nil
}
