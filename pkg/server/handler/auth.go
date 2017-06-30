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
	"fmt"

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
	Username         string                 `mapstructure:"username"`
	Email            string                 `mapstructure:"email"`
	Password         string                 `mapstructure:"password"`
	Provider         string                 `mapstructure:"provider"`
	ProviderAuthData map[string]interface{} `mapstructure:"provider_auth_data"`
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

// SignupHandler creates an AuthInfo with the supplied information.
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
	AccessModel      skydb.AccessModel  `inject:"AccessModel"`
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
	p := &signupPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	store := h.TokenStore

	info := skydb.AuthInfo{}
	var authdata skydb.AuthData

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
		principalID, authdata, err := authProvider.Login(payload.Context, p.ProviderAuthData)
		if err != nil {
			response.Err = skyerr.NewError(skyerr.InvalidCredentials, "unable to login with the given credentials")
			return
		}
		log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider)

		// Create new user info and set updated auth data
		info = skydb.NewProviderInfoAuthInfo(principalID, authdata)
	} else {
		info = skydb.NewAuthInfo(p.Password)
		authdata = make(skydb.AuthData)
		authdata["username"] = p.Username
		authdata["email"] = p.Email
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
		payload.DBConn, payload.Database, h.AssetStore, h.HookRegistry, payload.Context,
	}
	if response.Err = createContext.execute(&info, &authdata); response.Err != nil {
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

	response.Result = NewAuthResponse(info, &authdata, token.AccessToken)
}

type loginPayload struct {
	Username         string                 `mapstructure:"username"`
	Email            string                 `mapstructure:"email"`
	Password         string                 `mapstructure:"password"`
	Provider         string                 `mapstructure:"provider"`
	ProviderAuthData map[string]interface{} `mapstructure:"provider_auth_data"`
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

	info := skydb.AuthInfo{}
	authdata := skydb.AuthData{
		"username": p.Username,
		"email":    p.Email,
	}

	if p.Provider != "" {
		// Get AuthProvider and authenticates the user
		principalID, providerAuthData, skyErr := h.authPrincipal(payload.Context, p)
		if skyErr != nil {
			response.Err = skyErr
			return
		}
		if err := payload.DBConn.GetAuthByPrincipalID(principalID, &info); err != nil {

			// Create user if and only if no user found with the same principal
			if err != skydb.ErrUserNotFound {
				// TODO: more error handling here if necessary
				response.Err = skyerr.NewResourceFetchFailureErr("user", p.Username)
				return
			}

			info = skydb.NewProviderInfoAuthInfo(principalID, providerAuthData)
			authdata = skydb.AuthData{}
			createContext := createUserWithRecordContext{
				payload.DBConn, payload.Database, h.AssetStore, h.HookRegistry, payload.Context,
			}

			if response.Err = createContext.execute(&info, &authdata); response.Err != nil {
				return
			}
		} else {
			info.SetProviderInfoData(principalID, providerAuthData)
			if err := payload.DBConn.UpdateAuth(&info); err != nil {
				response.Err = skyerr.MakeError(err)
				return
			}
		}
	} else {
		fetcher := newUserAuthFetcher(payload.Database, payload.DBConn, authdata, &info, nil)
		if err := fetcher.FetchAuth(); err != nil {
			if err == skydb.ErrUserNotFound {
				response.Err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			} else {
				// TODO: more error handling here if necessary
				response.Err = skyerr.NewResourceFetchFailureErr("user", p.Username)
			}
			return
		}

		authdata = fetcher.AuthData

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

	authResponse := NewAuthResponse(info, &authdata, token.AccessToken)
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
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *PasswordHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
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

	info := skydb.AuthInfo{}
	if err := payload.DBConn.GetAuth(payload.AuthInfoID, &info); err != nil {
		if err == skydb.ErrUserNotFound {
			response.Err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
		} else {
			// TODO: more error handling here if necessary
			response.Err = skyerr.NewResourceFetchFailureErr("user", payload.AuthInfoID)
		}
		return
	}

	if !info.IsSamePassword(p.OldPassword) {
		log.Debug("Incorrect old password")
		response.Err = skyerr.NewError(skyerr.InvalidCredentials, "Incorrect old password")
		return
	}
	info.SetPassword(p.NewPassword)
	if err := payload.DBConn.UpdateAuth(&info); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	if p.Invalidate {
		log.Warningf("Invalidate is not yet implement")
		// TODO: invalidate all existing token and generate a new one for response
	}
	// Generate new access-token. Because InjectUserIfPresent preprocessor
	// will expire existing access-token.
	store := h.TokenStore
	token, err := store.NewToken(payload.AppName, info.ID)
	if err != nil {
		panic(err)
	}
	if err = store.Put(&token); err != nil {
		panic(err)
	}

	response.Result = AuthResponse{
		UserID:      info.ID,
		AccessToken: token.AccessToken,
	}
}

type UserAuthFetcher struct {
	DBConn   skydb.Conn
	Database skydb.Database

	// AuthData is the input of the Fetcher
	// while a user may contain more than one set of AuthData
	// it can be updated after the user is fetched
	AuthData skydb.AuthData

	// Result
	AuthInfo *skydb.AuthInfo
	User     *skydb.Record
}

func newUserAuthFetcher(db skydb.Database, conn skydb.Conn, authData skydb.AuthData, authInfo *skydb.AuthInfo, user *skydb.Record) UserAuthFetcher {
	if authInfo == nil {
		authInfo = &skydb.AuthInfo{}
	}

	if user == nil {
		user = &skydb.Record{}
	}

	return UserAuthFetcher{
		DBConn:   conn,
		Database: db,
		AuthData: authData,
		AuthInfo: authInfo,
		User:     user,
	}
}

func (f *UserAuthFetcher) FetchAuth() error {
	if err := f.FetchUser(); err != nil {
		return err
	}

	// update auth data
	f.AuthData["username"] = f.User.Data["username"]
	f.AuthData["email"] = f.User.Data["email"]

	return f.DBConn.GetAuth(f.User.ID.Key, f.AuthInfo)
}

func (f *UserAuthFetcher) FetchUser() error {
	var err error
	if err = validateAuthData(f.AuthData); err != nil {
		return err
	}

	query := f.buildAuthDataQuery()

	var results *skydb.Rows
	results, err = f.Database.Query(&query)
	if err != nil {
		return err
	}
	defer results.Close()

	records := []skydb.Record{}
	for results.Scan() {
		record := results.Record()
		records = append(records, record)
	}

	if err = results.Err(); err != nil {
		return err
	}

	recordsQueried := len(records)
	if recordsQueried == 0 {
		return skydb.ErrUserNotFound
	} else if recordsQueried > 1 {
		panic(fmt.Errorf("want 1 records queried, got %v", recordsQueried))
	}

	f.User = &records[0]

	return nil
}

func (f *UserAuthFetcher) buildAuthDataQuery() skydb.Query {
	username, _ := f.AuthData["username"].(string)
	email, _ := f.AuthData["email"].(string)

	predicates := []skydb.Predicate{}

	makeILikePredicate := func(keyPath string, value string) skydb.Predicate {
		return skydb.Predicate{
			Operator: skydb.ILike,
			Children: []interface{}{
				skydb.Expression{Type: skydb.KeyPath, Value: keyPath},
				skydb.Expression{Type: skydb.Literal, Value: value},
			},
		}
	}

	if username != "" {
		predicates = append(predicates, makeILikePredicate("username", username))
	}

	if email != "" {
		predicates = append(predicates, makeILikePredicate("email", email))
	}

	var predicate skydb.Predicate
	if len(predicates) == 1 {
		predicate = predicates[0]
	} else {
		predicate = skydb.Predicate{
			Operator: skydb.And,
			Children: []interface{}{predicates[0], predicates[1]},
		}
	}

	query := skydb.Query{
		Type:      "user",
		Predicate: predicate,
		Limit:     new(uint64),
	}
	*query.Limit = 1
	return query
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

func (ctx *createUserWithRecordContext) execute(info *skydb.AuthInfo, authData *skydb.AuthData) skyerr.Error {
	db := ctx.Database
	txDB, ok := db.(skydb.Transactional)
	if !ok {
		return skyerr.NewError(skyerr.NotSupported, "database impl does not support transaction")
	}

	txErr := withTransaction(txDB, func() error {
		// Check if AuthData duplicated only when it is provided
		// AuthData may be absent, e.g. login with provider
		if validateAuthData(*authData) == nil {
			fetcher := newUserAuthFetcher(db, ctx.DBConn, *authData, nil, nil)
			err := fetcher.FetchAuth()
			if err == nil {
				return errUserDuplicated
			}

			if err != skydb.ErrUserNotFound {
				return skyerr.MakeError(err)
			}
		}

		if err := ctx.DBConn.CreateAuth(info); err != nil {
			if err == skydb.ErrUserDuplicated {
				return errUserDuplicated
			}

			username, _ := (*authData)["username"].(string)
			return skyerr.NewResourceSaveFailureErrWithStringID("user", username)
		}

		userRecord := skydb.Record{
			ID:   skydb.NewRecordID(db.UserRecordType(), info.ID),
			Data: skydb.Data(*authData),
		}

		recordReq := recordModifyRequest{
			Db:           db,
			Conn:         ctx.DBConn,
			AssetStore:   ctx.AssetStore,
			HookRegistry: ctx.HookRegistry,
			Atomic:       false,
			Context:      ctx.Context,
			AuthInfo:     info,
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

// TODO: validate according to environment variable
func validateAuthData(authData skydb.AuthData) error {
	username, _ := authData["username"].(string)
	email, _ := authData["email"].(string)

	if username == "" && email == "" {
		return fmt.Errorf("Either username and email must be set")
	}

	return nil
}
