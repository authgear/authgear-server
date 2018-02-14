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

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/plugin/provider"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// Define the playload for sso plugin to login user with provider
type loginProviderPayload struct {
	Provider        string                 `mapstructure:"provider"`
	PrincipalID     string                 `mapstructure:"principal_id"`
	TokenResponse   map[string]interface{} `mapstructure:"token_response"`
	ProviderProfile map[string]interface{} `mapstructure:"provider_profile"`
}

func (payload *loginProviderPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *loginProviderPayload) Validate() skyerr.Error {
	if payload.Provider == "" {
		return skyerr.NewInvalidArgument("empty provider", []string{"provider"})
	}

	if payload.PrincipalID == "" {
		return skyerr.NewInvalidArgument("empty principal id", []string{"principal_id"})
	}

	return nil
}

// LoginProviderHandler login user with provider information
//
// LoginProviderHandler receives parameters:
//
// * provider (string, required)
// * principal_id (string, required)
// * token_response (json object, optional)
// * provider_profile (json object, optional)
//
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
// 		"action": "sso:oauth:login",
// 		"provider": "facebook",
// 		"principal_id": "104174434987489953648",
// 		"token_response": {
//			"access_token": "..."
//		},
// 		"provider_profile": {},
// }
// EOF
// Response
// if login exist
// 		return user and token
// else
// 		return skyerr.InvalidCredentials
//

type LoginProviderHandler struct {
	TokenStore       authtoken.Store    `inject:"TokenStore"`
	ProviderRegistry *provider.Registry `inject:"ProviderRegistry"`
	HookRegistry     *hook.Registry     `inject:"HookRegistry"`
	AssetStore       asset.Store        `inject:"AssetStore"`
	AuthRecordKeys   [][]string         `inject:"AuthRecordKeys"`
	AccessKey        router.Processor   `preprocessor:"accesskey"`
	DBConn           router.Processor   `preprocessor:"dbconn"`
	InjectPublicDB   router.Processor   `preprocessor:"inject_public_db"`
	PluginReady      router.Processor   `preprocessor:"plugin_ready"`
	RequireMasterKey router.Processor   `preprocessor:"require_master_key"`
	preprocessors    []router.Processor
}

func (h *LoginProviderHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
		h.InjectPublicDB,
		h.PluginReady,
		h.RequireMasterKey,
	}
}

func (h *LoginProviderHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *LoginProviderHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("Login provider")
	p := &loginProviderPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	store := h.TokenStore
	oauth := skydb.OAuthInfo{}
	info := skydb.AuthInfo{}
	user := skydb.Record{}
	now := timeNow()

	if err := payload.DBConn.GetOAuthInfo(p.Provider, p.PrincipalID, &oauth); err != nil {
		response.Err = skyerr.NewError(skyerr.InvalidCredentials, "no connected user")
		return
	}

	oauth.TokenResponse = p.TokenResponse
	oauth.ProviderProfile = p.ProviderProfile
	oauth.UpdatedAt = &now
	if err := payload.DBConn.UpdateOAuthInfo(&oauth); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	if err := payload.Database.Get(skydb.NewRecordID("user", oauth.UserID), &user); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	if err := payload.DBConn.GetAuth(oauth.UserID, &info); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	if err := checkUserIsNotDisabled(&info); err != nil {
		response.Err = err
		return
	}

	// generate access-token
	token, err := store.NewToken(payload.AppName, oauth.UserID)
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
	return
}

// Define the playload for sso plugin to signup user with provider
type signupProviderPayload struct {
	Provider        string                 `mapstructure:"provider"`
	PrincipalID     string                 `mapstructure:"principal_id"`
	Profile         skydb.Data             `mapstructure:"profile"`
	TokenResponse   map[string]interface{} `mapstructure:"token_response"`
	ProviderProfile map[string]interface{} `mapstructure:"provider_profile"`
}

func (payload *signupProviderPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *signupProviderPayload) Validate() skyerr.Error {
	if payload.Provider == "" {
		return skyerr.NewInvalidArgument("empty provider", []string{"provider"})
	}

	if payload.PrincipalID == "" {
		return skyerr.NewInvalidArgument("empty principal id", []string{"principal_id"})
	}

	return nil
}

// SignupProviderHandler create new user with provider information
//
// LoginProviderHandler receives parameters:
//
// * provider (string, required)
// * principal_id (string, required)
// * token_response (json object, optional)
// * provider_profile (json object, optional)
// * profile (json object, optional)
//
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
// 		"action": "sso:oauth:signup",
// 		"provider": "facebook",
// 		"principal_id": "104174434987489953648",
// 		"token_response": {
// 			"access_token": "access_token"
// 		},
// 		"provider_profile": {
//			"id": "104174434987489953648",
// 			"email": "chima@skygeario.com"
// 		},
// 		"profile": {"email": "chima@skygeario.com"}
// }
// EOF
// Response
// if no connected user
// 		return user and token
// else
// 		return skyerr.InvalidArgument

type SignupProviderHandler struct {
	TokenStore       authtoken.Store  `inject:"TokenStore"`
	HookRegistry     *hook.Registry   `inject:"HookRegistry"`
	AssetStore       asset.Store      `inject:"AssetStore"`
	AuthRecordKeys   [][]string       `inject:"AuthRecordKeys"`
	AccessKey        router.Processor `preprocessor:"accesskey"`
	DBConn           router.Processor `preprocessor:"dbconn"`
	InjectPublicDB   router.Processor `preprocessor:"inject_public_db"`
	PluginReady      router.Processor `preprocessor:"plugin_ready"`
	RequireMasterKey router.Processor `preprocessor:"require_master_key"`
	preprocessors    []router.Processor
}

func (h *SignupProviderHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
		h.InjectPublicDB,
		h.PluginReady,
		h.RequireMasterKey,
	}
}

func (h *SignupProviderHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SignupProviderHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("Signup provider")
	p := &signupProviderPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	var (
		oauth skydb.OAuthInfo
	)
	store := h.TokenStore
	info := skydb.AuthInfo{}
	user := skydb.Record{}
	now := timeNow()

	if err := payload.DBConn.GetOAuthInfo(p.Provider, p.PrincipalID, &oauth); err != nil {
		if err != skydb.ErrUserNotFound {
			// TODO: more error handling here if necessary
			response.Err = skyerr.NewResourceFetchFailureErr("provider", p.Provider)
			return
		}

		// oauth record not found
		// create new user with anonymous authInfo
		info = skydb.NewAnonymousAuthInfo()
		createContext := createUserWithRecordContext{
			payload.DBConn, payload.Database, h.AssetStore, h.HookRegistry, h.AuthRecordKeys, payload.Context,
		}
		createdUser, err := createContext.execute(&info, skydb.AuthData{}, p.Profile)
		if err != nil {
			response.Err = skyerr.MakeError(err)
			return
		}

		oauth = skydb.OAuthInfo{
			UserID:          info.ID,
			Provider:        p.Provider,
			PrincipalID:     p.PrincipalID,
			TokenResponse:   p.TokenResponse,
			ProviderProfile: p.ProviderProfile,
			CreatedAt:       &now,
			UpdatedAt:       &now,
		}

		if err := payload.DBConn.CreateOAuthInfo(&oauth); err != nil {
			response.Err = skyerr.MakeError(err)
			return
		}

		user = *createdUser
	} else {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "user already connected")
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
	return
}

// Define the playload for sso plugin to connect user with provider
type linkProviderPayload struct {
	Provider        string                 `mapstructure:"provider"`
	PrincipalID     string                 `mapstructure:"principal_id"`
	TokenResponse   map[string]interface{} `mapstructure:"token_response"`
	ProviderProfile map[string]interface{} `mapstructure:"provider_profile"`
	UserID          string                 `mapstructure:"user_id"`
}

func (payload *linkProviderPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *linkProviderPayload) Validate() skyerr.Error {
	if payload.Provider == "" {
		return skyerr.NewInvalidArgument("empty provider", []string{"provider"})
	}

	if payload.PrincipalID == "" {
		return skyerr.NewInvalidArgument("empty principal id", []string{"principal_id"})
	}

	if payload.UserID == "" {
		return skyerr.NewInvalidArgument("empty user id", []string{"user_id"})
	}

	return nil
}

// LinkProviderHandler connect user with provider information
//
// LinkProviderHandler receives parameters:
//
// * provider (string, required)
// * principal_id (string, required)
// * user_id (string, required)
// * token_response (json object, optional)
// * provider_profile (json object, optional)
//
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
// 		"action": "sso:oauth:link",
// 		"provider": "facebook",
// 		"principal_id": "104174434987489953648",
// 		"user_id": "c0959b6b-15ea-4e21-8afb-9c8308ad79db",
// 		"token_response": {
// 			"access_token": "access_token"
// 		},
// 		"provider_profile": {
//			"id": "104174434987489953648",
// 			"email": "chima@skygeario.com"
// 		}
// }
// EOF
// Response
// {
//     "result": "OK"
// }
type LinkProviderHandler struct {
	HookRegistry     *hook.Registry   `inject:"HookRegistry"`
	AssetStore       asset.Store      `inject:"AssetStore"`
	AuthRecordKeys   [][]string       `inject:"AuthRecordKeys"`
	AccessKey        router.Processor `preprocessor:"accesskey"`
	DBConn           router.Processor `preprocessor:"dbconn"`
	InjectPublicDB   router.Processor `preprocessor:"inject_public_db"`
	PluginReady      router.Processor `preprocessor:"plugin_ready"`
	RequireMasterKey router.Processor `preprocessor:"require_master_key"`
	preprocessors    []router.Processor
}

func (h *LinkProviderHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
		h.InjectPublicDB,
		h.PluginReady,
		h.RequireMasterKey,
	}
}

func (h *LinkProviderHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *LinkProviderHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("Link provider")
	p := &linkProviderPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	oauth := skydb.OAuthInfo{}
	info := skydb.AuthInfo{}
	userID := p.UserID

	if err := payload.DBConn.GetOAuthInfo(p.Provider, p.PrincipalID, &oauth); err != nil {
		if err != skydb.ErrUserNotFound {
			response.Err = skyerr.NewResourceFetchFailureErr("sso_auth", p.PrincipalID)
			return
		}
	} else {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "provider account already linked with existing user")
		return
	}

	if err := payload.DBConn.GetOAuthInfoByProviderAndUserID(p.Provider, userID, &oauth); err != skydb.ErrUserNotFound {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "user linked to the provider already")
		return
	}

	if err := payload.DBConn.GetAuth(userID, &info); err != nil {
		response.Err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
		return
	}

	// new oauth record for linking provider
	now := timeNow()
	oauth = skydb.OAuthInfo{
		UserID:          info.ID,
		Provider:        p.Provider,
		PrincipalID:     p.PrincipalID,
		TokenResponse:   p.TokenResponse,
		ProviderProfile: p.ProviderProfile,
		CreatedAt:       &now,
		UpdatedAt:       &now,
	}

	if err := payload.DBConn.CreateOAuthInfo(&oauth); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	response.Result = "OK"
	return
}

// Define the playload for sso plugin to disconnect user with provider
type unlinkProviderPayload struct {
	Provider string `mapstructure:"provider"`
	UserID   string `mapstructure:"user_id"`
}

func (payload *unlinkProviderPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *unlinkProviderPayload) Validate() skyerr.Error {
	if payload.Provider == "" {
		return skyerr.NewInvalidArgument("empty provider", []string{"provider"})
	}

	if payload.UserID == "" {
		return skyerr.NewInvalidArgument("empty user id", []string{"user_id"})
	}

	return nil
}

// UnlinkProviderHandler disconnect user with specific provider
//
// UnlinkProviderHandler receives parameters:
//
// * provider (string, required)
// * user_id (string, required)
//
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
// 		"action": "sso:oauth:unlink",
// 		"provider": "facebook",
// 		"user_id": "c0959b6b-15ea-4e21-8afb-9c8308ad79db"
// }
// EOF
// Response
// {
//     "result": "OK"
// }
type UnlinkProviderHandler struct {
	HookRegistry     *hook.Registry   `inject:"HookRegistry"`
	AssetStore       asset.Store      `inject:"AssetStore"`
	AuthRecordKeys   [][]string       `inject:"AuthRecordKeys"`
	AccessKey        router.Processor `preprocessor:"accesskey"`
	DBConn           router.Processor `preprocessor:"dbconn"`
	InjectPublicDB   router.Processor `preprocessor:"inject_public_db"`
	PluginReady      router.Processor `preprocessor:"plugin_ready"`
	RequireMasterKey router.Processor `preprocessor:"require_master_key"`
	preprocessors    []router.Processor
}

func (h *UnlinkProviderHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
		h.InjectPublicDB,
		h.PluginReady,
		h.RequireMasterKey,
	}
}

func (h *UnlinkProviderHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *UnlinkProviderHandler) Handle(payload *router.Payload, response *router.Response) {
	log.Debugf("Unlink provider")
	p := &unlinkProviderPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	oauth := skydb.OAuthInfo{}

	if err := payload.DBConn.GetOAuthInfoByProviderAndUserID(p.Provider, p.UserID, &oauth); err != nil {
		response.Err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
		return
	}

	if err := payload.DBConn.DeleteOAuth(oauth.Provider, oauth.PrincipalID); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	response.Result = "OK"
	return
}
