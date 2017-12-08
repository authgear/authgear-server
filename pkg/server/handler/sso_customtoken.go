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
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type ssoCustomTokenClaims struct {
	Profile skydb.Data `json:"skyprofile"`
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

	if parsedToken, err := jwt.ParseWithClaims(
		payload.TokenString,
		&payload.Claims,
		payload.keyFunc,
	); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	} else {
		payload.Token = parsedToken
	}

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
SSOCustomTokenLoginHandler authenticate user with password

The user can be either identified by username or password.

curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
	"action": "sso:custom_token:login",
	"token": "eyXXXXXXX",
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

	if err := payload.DBConn.GetCustomTokenInfo(principalID, &customTokenInfo); err != nil {
		if err != skydb.ErrUserNotFound {
			return skyerr.MakeError(err)
		}
		createNewUser = true
		*authinfo = skydb.NewAnonymousAuthInfo()
	}

	if !createNewUser {
		if err := payload.DBConn.GetAuth(customTokenInfo.UserID, authinfo); err != nil {
			if err != skydb.ErrUserNotFound {
				return skyerr.MakeError(err)
			}
			createNewUser = true
		}
	}

	if createNewUser {
		createContext := createUserWithRecordContext{
			payload.DBConn,
			payload.Database,
			h.AssetStore,
			h.HookRegistry,
			h.AuthRecordKeys,
			payload.Context,
		}

		*authinfo = skydb.NewAnonymousAuthInfo()
		if createdUser, err := createContext.execute(
			authinfo,
			skydb.AuthData{},
			p.Profile,
		); err != nil {
			return skyerr.MakeError(err)
		} else {
			*user = *createdUser
		}

		now := timeNow()
		customTokenInfo = skydb.CustomTokenInfo{
			UserID:      user.ID.Key,
			PrincipalID: principalID,
			CreatedAt:   &now,
		}

		if err := payload.DBConn.CreateCustomTokenInfo(&customTokenInfo); err != nil {
			return skyerr.MakeError(err)
		}
	} else {
		if err := payload.DBConn.UpdateAuth(authinfo); err != nil {
			return skyerr.MakeError(err)
		}

		userRecordID := skydb.NewRecordID("user", customTokenInfo.UserID)
		if err := payload.Database.Get(userRecordID, user); err != nil {
			return skyerr.MakeError(err)
		}
	}

	return nil
}
