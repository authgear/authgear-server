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
	"context"

	"github.com/mitchellh/mapstructure"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/plugin/provider"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

var errUserDuplicated = skyerr.NewError(skyerr.Duplicated, "user duplicated")

type signupPayload struct {
	AuthDataData     map[string]interface{} `mapstructure:"auth_data"`
	AuthRecordKeys   [][]string
	AuthData         skydb.AuthData
	Password         string                 `mapstructure:"password"`
	Provider         string                 `mapstructure:"provider"`
	ProviderAuthData map[string]interface{} `mapstructure:"provider_auth_data"`
	Profile          skydb.Data             `mapstructure:"profile"`
}

func (payload *signupPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	payload.AuthData = skydb.NewAuthData(payload.AuthDataData, payload.AuthRecordKeys)
	return payload.Validate()
}

func (payload *signupPayload) Validate() skyerr.Error {
	if payload.IsAnonymous() {
		//no validation logic for anonymous sign up
	} else if payload.Provider == "" {
		identified := payload.AuthData.IsValid()
		if !identified {
			return skyerr.NewInvalidArgument("invalid auth data", []string{"auth_data"})
		}

		if duplicatedKeys := payload.duplicatedKeysInAuthDataAndProfile(); len(duplicatedKeys) > 0 {
			return skyerr.NewInvalidArgument("duplicated keys found in auth data in profile", duplicatedKeys)
		}

		if payload.Password == "" {
			return skyerr.NewInvalidArgument("empty password", []string{"password"})
		}
	}

	return nil
}

func (payload *signupPayload) IsAnonymous() bool {
	return payload.AuthData.IsEmpty() && payload.Password == "" && payload.Provider == ""
}

func (payload *signupPayload) duplicatedKeysInAuthDataAndProfile() []string {
	keys := []string{}

	for k := range payload.AuthData.GetData() {
		if _, found := payload.Profile[k]; found {
			keys = append(keys, k)
		}
	}

	return keys
}

// SignupHandler creates an AuthInfo with the supplied information.
//
// SignupHandler receives three parameters:
//
// * auth_data (json object, optional)
// * password (string, optional)
//
// If auth_data is not supplied, an anonymous user is created and
// have user_id auto-generated. SignupHandler writes an error to
// response.Result if the supplied username or email collides with an existing
// username.
//
// Any entry with null value in auth_data would be purged. If all entries are
// having null value, this would be treated as anonymous sign up.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/ <<EOF
//  {
//      "action": "auth:signup",
//      "auth_data": {
//        "username": "rickmak",
//        "email": "rick.mak@gmail.com",
//      },
//      "password": "123456"
//  }
//  EOF
type SignupHandler struct {
	TokenStore       authtoken.Store    `inject:"TokenStore"`
	ProviderRegistry *provider.Registry `inject:"ProviderRegistry"`
	HookRegistry     *hook.Registry     `inject:"HookRegistry"`
	AssetStore       asset.Store        `inject:"AssetStore"`
	AccessModel      skydb.AccessModel  `inject:"AccessModel"`
	AuthRecordKeys   [][]string         `inject:"AuthRecordKeys"`
	AccessKey        router.Processor   `preprocessor:"accesskey"`
	DBConn           router.Processor   `preprocessor:"dbconn"`
	InjectPublicDB   router.Processor   `preprocessor:"inject_public_db"`
	PluginReady      router.Processor   `preprocessor:"plugin_ready"`
	preprocessors    []router.Processor
}

func (h *SignupHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
		h.InjectPublicDB,
		h.PluginReady,
	}
}

func (h *SignupHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SignupHandler) Handle(payload *router.Payload, response *router.Response) {
	p := &signupPayload{
		AuthRecordKeys: h.AuthRecordKeys,
	}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	store := h.TokenStore

	info := skydb.AuthInfo{}
	authdata := skydb.AuthData{}

	if p.IsAnonymous() {
		info = skydb.NewAnonymousAuthInfo()
	} else if p.Provider != "" {
		// Get AuthProvider and authenticates the user
		log.Debugf(`Client requested auth provider: "%v".`, p.Provider)
		authProvider, err := h.ProviderRegistry.GetAuthProvider(p.Provider)
		if err != nil {
			response.Err = skyerr.NewInvalidArgument(err.Error(), []string{"provider"})
			return
		}
		principalID, providerAuthData, err := authProvider.Login(payload.Context, p.ProviderAuthData)
		if err != nil {
			response.Err = skyerr.NewError(skyerr.InvalidCredentials, "unable to login with the given credentials")
			return
		}
		log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider)

		// Create new user info and set updated auth data
		info = skydb.NewProviderInfoAuthInfo(principalID, providerAuthData)
	} else {
		info = skydb.NewAuthInfo(p.Password)
		authdata = p.AuthData
	}

	// Populate the default roles to user
	if h.AccessModel == skydb.RoleBasedAccess {
		defaultRoles, err := payload.DBConn.GetDefaultRoles()
		if err != nil {
			response.Err = skyerr.NewError(skyerr.InternalQueryInvalid, "unable to query default roles")
			return
		}
		info.Roles = defaultRoles
	}

	// Populate the activity time to user
	now := timeNowUTC()
	info.LastLoginAt = &now
	info.LastSeenAt = &now

	createContext := createUserWithRecordContext{
		payload.DBConn, payload.Database, h.AssetStore, h.HookRegistry, h.AuthRecordKeys, payload.Context,
	}

	user, skyErr := createContext.execute(&info, authdata, p.Profile)
	if skyErr != nil {
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
	}.NewAuthResponse(info, *user, token.AccessToken)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	response.Result = authResponse
}

type loginPayload struct {
	AuthDataData     map[string]interface{} `mapstructure:"auth_data"`
	AuthRecordKeys   [][]string
	AuthData         skydb.AuthData
	Password         string                 `mapstructure:"password"`
	Provider         string                 `mapstructure:"provider"`
	ProviderAuthData map[string]interface{} `mapstructure:"provider_auth_data"`
}

func (payload *loginPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	payload.AuthData = skydb.NewAuthData(payload.AuthDataData, payload.AuthRecordKeys)
	return payload.Validate()
}

func (payload *loginPayload) Validate() skyerr.Error {
	if payload.Provider == "" {
		identified := payload.AuthData.IsValid()
		if !identified {
			return skyerr.NewInvalidArgument("invalid auth data", []string{"auth_data"})
		}
	}

	return nil
}

/*
LoginHandler authenticate user with password

The user can be either identified by username or password.

curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "auth:login",
    "username": "rickmak",
    "email": "rick.mak@gmail.com",
    "password": "123456"
}
EOF
*/
type LoginHandler struct {
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

func (h *LoginHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
		h.InjectPublicDB,
		h.PluginReady,
	}
}

func (h *LoginHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *LoginHandler) Handle(payload *router.Payload, response *router.Response) {
	p := &loginPayload{
		AuthRecordKeys: h.AuthRecordKeys,
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

	var handleLoginFunc func(*router.Payload, *loginPayload, *skydb.AuthInfo, *skydb.Record) skyerr.Error
	if p.Provider != "" {
		handleLoginFunc = h.handleLoginWithProvider
	} else {
		handleLoginFunc = h.handleLoginWithAuthData
	}

	if skyErr = handleLoginFunc(payload, p, &info, &user); skyErr != nil {
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
	}.NewAuthResponse(info, user, token.AccessToken)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	// Populate the activity time to user
	now := timeNow()
	info.LastLoginAt = &now
	info.LastSeenAt = &now
	if err := payload.DBConn.UpdateAuth(&info); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}
	response.Result = authResponse
}

func (h *LoginHandler) handleLoginWithProvider(payload *router.Payload, p *loginPayload, authinfo *skydb.AuthInfo, user *skydb.Record) skyerr.Error {
	principalID, providerAuthData, skyErr := h.authPrincipal(payload.Context, p)
	if skyErr != nil {
		return skyErr
	}

	if err := payload.DBConn.GetAuthByPrincipalID(principalID, authinfo); err != nil {
		// Create user if and only if no user found with the same principal
		if err != skydb.ErrUserNotFound {
			// TODO: more error handling here if necessary
			return skyerr.NewResourceFetchFailureErr("auth_data", p.AuthData)
		}

		*authinfo = skydb.NewProviderInfoAuthInfo(principalID, providerAuthData)

		createContext := createUserWithRecordContext{
			payload.DBConn, payload.Database, h.AssetStore, h.HookRegistry, h.AuthRecordKeys, payload.Context,
		}

		createdUser, err := createContext.execute(authinfo, skydb.AuthData{}, skydb.Data{})
		if err != nil {
			return skyerr.MakeError(err)
		}

		*user = *createdUser
	} else {
		authinfo.SetProviderInfoData(principalID, providerAuthData)
		if err := payload.DBConn.UpdateAuth(authinfo); err != nil {
			return skyerr.MakeError(err)
		}

		err := payload.Database.Get(skydb.NewRecordID("user", authinfo.ID), user)
		if err != nil {
			return skyerr.MakeError(err)
		}
	}

	return nil
}

func (h *LoginHandler) handleLoginWithAuthData(payload *router.Payload, p *loginPayload, authinfo *skydb.AuthInfo, user *skydb.Record) skyerr.Error {
	authdata := p.AuthData

	if ok := authdata.IsValid(); !ok {
		return skyerr.NewInvalidArgument("Unexpected key found", []string{"authdata"})
	}

	fetcher := newUserAuthFetcher(payload.Database, payload.DBConn)
	fetchedAuthInfo, fetchedUser, err := fetcher.FetchAuth(authdata)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			return skyerr.NewError(skyerr.ResourceNotFound, "user not found")
		}

		// TODO: more error handling here if necessary
		return skyerr.NewResourceFetchFailureErr("auth_data", p.AuthData)
	}

	*authinfo = fetchedAuthInfo
	*user = fetchedUser

	if !authinfo.IsSamePassword(p.Password) {
		return skyerr.NewError(skyerr.InvalidCredentials, "auth_data or password incorrect")
	}

	return nil
}

func (h *LoginHandler) authPrincipal(ctx context.Context, p *loginPayload) (string, map[string]interface{}, skyerr.Error) {
	log.Debugf(`Client requested auth provider: "%v".`, p.Provider)
	authProvider, err := h.ProviderRegistry.GetAuthProvider(p.Provider)
	if err != nil {
		skyErr := skyerr.NewInvalidArgument(err.Error(), []string{"provider"})
		return "", nil, skyErr
	}
	principalID, authData, err := authProvider.Login(ctx, p.ProviderAuthData)
	if err != nil {
		skyErr := skyerr.NewError(skyerr.InvalidCredentials, "invalid authentication information")
		return "", nil, skyErr
	}
	log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider)
	return principalID, authData, nil
}

// LogoutHandler receives an access token and invalidates it
type LogoutHandler struct {
	TokenStore    authtoken.Store  `inject:"TokenStore"`
	Authenticator router.Processor `preprocessor:"authenticator"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *LogoutHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.PluginReady,
	}
}

func (h *LogoutHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *LogoutHandler) Handle(payload *router.Payload, response *router.Response) {
	store := h.TokenStore
	accessToken := payload.AccessTokenString()

	var err error

	if err = store.Delete(accessToken); err != nil {
		if _, notfound := err.(*authtoken.NotFoundError); notfound {
			err = nil
		}
	}
	if err != nil {
		response.Err = skyerr.MakeError(err)
	} else {
		response.Result = struct {
			Status string `json:"status,omitempty"`
		}{
			"OK",
		}
	}
}

// Define the playload that change password handler will process
type passwordPayload struct {
	OldPassword string `mapstructure:"old_password"`
	NewPassword string `mapstructure:"password"`
	Invalidate  bool   `mapstructure:"invalidate"`
}

func (payload *passwordPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *passwordPayload) Validate() skyerr.Error {
	return nil
}

// PasswordHandler change the current user password
//
// PasswordHandler receives three parameters:
//
// * old_password (string, required)
// * password (string, required)
//
// If user is not logged in, an 404 not found will return.
//
//  Current implementation
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/ <<EOF
//  {
//      "action": "auth:password",
//      "old_password": "rick.mak@gmail.com",
//      "password": "123456"
//  }
//  EOF
// Response
// return existing access toektn if not invalidate
//
// TODO:
// Input accept `user_id` and `invalidate`.
// If `user_id` is supplied, will check authorization policy and see if existing
// accept `invalidate` and invaldate all existing access token.
// Return authInfoID with new AccessToken if the invalidate is true
type PasswordHandler struct {
	TokenStore    authtoken.Store  `inject:"TokenStore"`
	AssetStore    asset.Store      `inject:"AssetStore"`
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectAuth    router.Processor `preprocessor:"inject_auth"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	RequireAuth   router.Processor `preprocessor:"require_auth"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *PasswordHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectAuth,
		h.InjectUser,
		h.RequireAuth,
		h.PluginReady,
	}
}

func (h *PasswordHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *PasswordHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("changing password")
	p := &passwordPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	info := payload.AuthInfo
	if !info.IsSamePassword(p.OldPassword) {
		log.Debug("Incorrect old password")
		response.Err = skyerr.NewError(skyerr.InvalidCredentials, "Incorrect old password")
		return
	}
	info.SetPassword(p.NewPassword)
	if err := payload.DBConn.UpdateAuth(info); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	if p.Invalidate {
		log.Warningf("Invalidate is not yet implement")
		// TODO: invalidate all existing token and generate a new one for response
	}
	// Generate new access-token. Because InjectAuthIfPresent preprocessor
	// will expire existing access-token.
	store := h.TokenStore
	token, err := store.NewToken(payload.AppName, info.ID)
	if err != nil {
		panic(err)
	}
	if err = store.Put(&token); err != nil {
		panic(err)
	}

	user := payload.User
	if user == nil {
		user = &skydb.Record{}
	}

	authResponse, err := AuthResponseFactory{
		AssetStore: h.AssetStore,
		Conn:       payload.DBConn,
	}.NewAuthResponse(*info, *user, token.AccessToken)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	response.Result = authResponse
}
