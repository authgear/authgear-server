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
	"github.com/mitchellh/mapstructure"
	"golang.org/x/net/context"

	"github.com/skygeario/skygear-server/asset"
	"github.com/skygeario/skygear-server/authtoken"
	"github.com/skygeario/skygear-server/plugin/hook"
	"github.com/skygeario/skygear-server/plugin/provider"
	"github.com/skygeario/skygear-server/router"
	"github.com/skygeario/skygear-server/skydb"
	"github.com/skygeario/skygear-server/skyerr"
)

var errUserDuplicated = skyerr.NewError(skyerr.Duplicated, "user duplicated")

type authResponse struct {
	UserID      string `json:"user_id,omitempty"`
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

type signupPayload struct {
	Username string                 `mapstructure:"username"`
	Email    string                 `mapstructure:"email"`
	Password string                 `mapstructure:"password"`
	Provider string                 `mapstructure:"provider"`
	AuthData map[string]interface{} `mapstructure:"auth_data"`
}

func (payload *signupPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *signupPayload) Validate() skyerr.Error {

	if payload.IsAnonymous() {
		//no validation logic for anonymous sign up
	} else if payload.Provider == "" {
		identified := payload.Username != "" || payload.Email != ""
		if !identified {
			return skyerr.NewInvalidArgument("empty username and empty email", []string{"username", "email"})
		}

		if payload.Password == "" {
			return skyerr.NewInvalidArgument("empty password", []string{"password"})
		}
	}

	return nil
}

func (payload *signupPayload) IsAnonymous() bool {
	return payload.Email == "" && payload.Password == "" && payload.Username == "" && payload.Provider == ""
}

// SignupHandler creates an UserInfo with the supplied information.
//
// SignupHandler receives three parameters:
//
// * username (string, unique, optional)
// * email  (string, unqiue, optional)
// * password (string, optional)
//
// If both username and email is not supplied, an anonymous user is created and
// have user_id auto-generated. SignupHandler writes an error to
// response.Result if the supplied username or email collides with an existing
// username.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/ <<EOF
//  {
//      "action": "auth:signup",
//      "username": "rickmak",
//      "email": "rick.mak@gmail.com",
//      "password": "123456"
//  }
//  EOF
type SignupHandler struct {
	TokenStore       authtoken.Store    `inject:"TokenStore"`
	ProviderRegistry *provider.Registry `inject:"ProviderRegistry"`
	HookRegistry     *hook.Registry     `inject:"HookRegistry"`
	AssetStore       asset.Store        `inject:"AssetStore"`
	AccessKey        router.Processor   `preprocessor:"accesskey"`
	DBConn           router.Processor   `preprocessor:"dbconn"`
	InjectPublicDB   router.Processor   `preprocessor:"inject_public_db"`
	PluginReady      router.Processor   `preprocessor:"plugin"`
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
	p := &signupPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	store := h.TokenStore

	info := skydb.UserInfo{}
	if p.IsAnonymous() {
		info = skydb.NewAnonymousUserInfo()
	} else if p.Provider != "" {
		// Get AuthProvider and authenticates the user
		log.Debugf(`Client requested auth provider: "%v".`, p.Provider)
		authProvider, err := h.ProviderRegistry.GetAuthProvider(p.Provider)
		if err != nil {
			response.Err = skyerr.NewInvalidArgument(err.Error(), []string{"provider"})
			return
		}
		principalID, authData, err := authProvider.Login(p.AuthData)
		if err != nil {
			response.Err = skyerr.NewError(skyerr.InvalidCredentials, "unable to login with the given credentials")
			return
		}
		log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider)

		// Create new user info and set updated auth data
		info = skydb.NewProvidedAuthUserInfo(principalID, authData)
	} else {
		info = skydb.NewUserInfo(p.Username, p.Email, p.Password)
	}

	createContext := createUserWithRecordContext{
		payload.DBConn, payload.Database, h.AssetStore, h.HookRegistry, payload.Context,
	}
	if response.Err = createContext.execute(&info); response.Err != nil {
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

	response.Result = authResponse{
		UserID:      info.ID,
		Username:    info.Username,
		Email:       info.Email,
		AccessToken: token.AccessToken,
	}
}

type loginPayload struct {
	Username string                 `mapstructure:"username"`
	Email    string                 `mapstructure:"email"`
	Password string                 `mapstructure:"password"`
	Provider string                 `mapstructure:"provider"`
	AuthData map[string]interface{} `mapstructure:"auth_data"`
}

func (payload *loginPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *loginPayload) Validate() skyerr.Error {
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
	AccessKey        router.Processor   `preprocessor:"accesskey"`
	DBConn           router.Processor   `preprocessor:"dbconn"`
	InjectPublicDB   router.Processor   `preprocessor:"inject_public_db"`
	PluginReady      router.Processor   `preprocessor:"plugin"`
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
	p := &loginPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	if h.TokenStore == nil {
		panic("token store is nil")
	}
	store := h.TokenStore

	info := skydb.UserInfo{}

	if p.Provider != "" {
		// Get AuthProvider and authenticates the user
		log.Debugf(`Client requested auth provider: "%v".`, p.Provider)
		authProvider, err := h.ProviderRegistry.GetAuthProvider(p.Provider)
		if err != nil {
			response.Err = skyerr.NewInvalidArgument(err.Error(), []string{"provider"})
			return
		}
		principalID, authData, err := authProvider.Login(p.AuthData)
		if err != nil {
			response.Err = skyerr.NewError(skyerr.InvalidCredentials, "invalid authentication information")
			return
		}
		log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider)

		if err := payload.DBConn.GetUserByPrincipalID(principalID, &info); err != nil {
			// Create user if and only if no user found with the same principal
			if err != skydb.ErrUserNotFound {
				// TODO: more error handling here if necessary
				response.Err = skyerr.NewResourceFetchFailureErr("user", p.Username)
				return
			}

			info = skydb.NewProvidedAuthUserInfo(principalID, authData)
			createContext := createUserWithRecordContext{
				payload.DBConn, payload.Database, h.AssetStore, h.HookRegistry, payload.Context,
			}
			if response.Err = createContext.execute(&info); response.Err != nil {
				return
			}
		} else {
			info.SetProvidedAuthData(principalID, authData)
			if err := payload.DBConn.UpdateUser(&info); err != nil {
				response.Err = skyerr.MakeError(err)
				return
			}
		}
	} else {
		if err := payload.DBConn.GetUserByUsernameEmail(p.Username, p.Email, &info); err != nil {
			if err == skydb.ErrUserNotFound {
				response.Err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			} else {
				// TODO: more error handling here if necessary
				response.Err = skyerr.NewResourceFetchFailureErr("user", p.Username)
			}
			return
		}

		if !info.IsSamePassword(p.Password) {
			response.Err = skyerr.NewError(skyerr.InvalidCredentials, "username or password incorrect")
			return
		}
	}

	// generate access-token
	token, err := store.NewToken(payload.AppName, info.ID)
	if err != nil {
		panic(err)
	}

	if err = store.Put(&token); err != nil {
		panic(err)
	}

	response.Result = authResponse{
		UserID:      info.ID,
		Username:    info.Username,
		Email:       info.Email,
		AccessToken: token.AccessToken,
	}
}

// LogoutHandler receives an access token and invalidates it
type LogoutHandler struct {
	TokenStore    authtoken.Store  `inject:"TokenStore"`
	Authenticator router.Processor `preprocessor:"authenticator"`
	PluginReady   router.Processor `preprocessor:"plugin"`
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
// Return userInfoID with new AccessToken if the invalidate is true
type PasswordHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	preprocessors []router.Processor
}

func (h *PasswordHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
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

	info := skydb.UserInfo{}
	if err := payload.DBConn.GetUser(payload.UserInfoID, &info); err != nil {
		if err == skydb.ErrUserNotFound {
			response.Err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
		} else {
			// TODO: more error handling here if necessary
			response.Err = skyerr.NewResourceFetchFailureErr("user", payload.UserInfoID)
		}
		return
	}

	if !info.IsSamePassword(p.OldPassword) {
		log.Debug("Incorrect old password")
		response.Err = skyerr.NewError(skyerr.InvalidCredentials, "Incorrect old password")
		return
	}
	info.SetPassword(p.NewPassword)
	if err := payload.DBConn.UpdateUser(&info); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	if p.Invalidate {
		log.Warningf("Invalidate is not yet implement")
		// TODO: invalidate all existing token and generate a new one for response
	}
	response.Result = authResponse{
		UserID:      info.ID,
		AccessToken: payload.AccessTokenString(),
	}
}

// createUserWithRecordContext is a context for creating a new user with
// database record
type createUserWithRecordContext struct {
	DBConn       skydb.Conn
	Database     skydb.Database
	AssetStore   asset.Store
	HookRegistry *hook.Registry
	Context      context.Context
}

func (ctx *createUserWithRecordContext) execute(info *skydb.UserInfo) skyerr.Error {
	db := ctx.Database
	txDB, ok := db.(skydb.TxDatabase)
	if !ok {
		return skyerr.NewError(skyerr.NotSupported, "database impl does not support transaction")
	}

	txErr := withTransaction(txDB, func() error {
		if err := ctx.DBConn.CreateUser(info); err != nil {
			if err == skydb.ErrUserDuplicated {
				return errUserDuplicated
			}
			return skyerr.NewResourceSaveFailureErrWithStringID("user", info.Username)
		}

		userRecord := skydb.Record{
			ID: skydb.NewRecordID(db.UserRecordType(), info.ID),
		}

		recordReq := recordModifyRequest{
			Db:           db,
			Conn:         ctx.DBConn,
			AssetStore:   ctx.AssetStore,
			HookRegistry: ctx.HookRegistry,
			Atomic:       false,
			Context:      ctx.Context,
			UserInfo:     info,
			RecordsToSave: []*skydb.Record{
				&userRecord,
			},
		}

		recordResp := recordModifyResponse{
			ErrMap: map[skydb.RecordID]skyerr.Error{},
		}

		return recordSaveHandler(&recordReq, &recordResp)
	})

	if txErr == nil {
		return nil
	}

	if err, ok := txErr.(skyerr.Error); ok {
		return err
	}

	return skyerr.MakeError(txErr)
}
