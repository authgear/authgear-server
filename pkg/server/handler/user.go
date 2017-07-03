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
	"github.com/skygeario/skygear-server/pkg/server/plugin/provider"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type queryPayload struct {
	Emails    []string `mapstructure:"emails"`
	Usernames []string `mapstructure:"usernames"`
}

func (payload *queryPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *queryPayload) Validate() skyerr.Error {
	return nil
}

type UserQueryHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *UserQueryHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.PluginReady,
	}
}

func (h *UserQueryHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *UserQueryHandler) Handle(payload *router.Payload, response *router.Response) {
	adminRoles, err := payload.DBConn.GetAdminRoles()
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	if !payload.HasMasterKey() {
		authinfo := payload.AuthInfo
		if authinfo == nil {
			response.Err = skyerr.NewError(skyerr.NotAuthenticated, "Authentication is needed to query user")
			return
		} else if !authinfo.HasAnyRoles(adminRoles) {
			response.Err = skyerr.NewError(skyerr.PermissionDenied, "No permission to query user")
			return
		}
	}

	qp := &queryPayload{}
	skyErr := qp.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	authinfos, err := payload.DBConn.QueryUser(qp.Emails, qp.Usernames)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	results := make([]interface{}, len(authinfos))
	for i, authinfo := range authinfos {
		results[i] = map[string]interface{}{
			"id":   authinfo.ID,
			"type": "user",
			"data": struct {
				ID       string   `json:"_id"`
				Email    string   `json:"email"`
				Username string   `json:"username"`
				Roles    []string `json:"roles,omitempty"`
			}{authinfo.ID, authinfo.Email, authinfo.Username, authinfo.Roles},
		}
	}
	response.Result = results
}

type userUpdatePayload struct {
	ID       string   `mapstructure:"_id"`
	Username string   `mapstructure:"username"`
	Email    string   `mapstructure:"email"`
	Roles    []string `mapstructure:"roles"`
}

func (payload *userUpdatePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *userUpdatePayload) Validate() skyerr.Error {
	roleMap := map[string]bool{}
	for _, role := range payload.Roles {
		existed, ok := roleMap[role]
		if existed {
			return skyerr.NewInvalidArgument("duplicated roles in payload", []string{"roles"})
		}
		if !ok {
			roleMap[role] = true
		}
	}
	if payload.ID == "" {
		return skyerr.NewInvalidArgument("missing required fields", []string{"_id"})
	}
	return nil
}

type UserUpdateHandler struct {
	AccessModel   skydb.AccessModel `inject:"AccessModel"`
	Authenticator router.Processor  `preprocessor:"authenticator"`
	DBConn        router.Processor  `preprocessor:"dbconn"`
	InjectUser    router.Processor  `preprocessor:"inject_user"`
	InjectDB      router.Processor  `preprocessor:"inject_db"`
	RequireUser   router.Processor  `preprocessor:"require_user"`
	PluginReady   router.Processor  `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *UserUpdateHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
		h.PluginReady,
	}
}

func (h *UserUpdateHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *UserUpdateHandler) Handle(payload *router.Payload, response *router.Response) {
	p := &userUpdatePayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	if p.Roles != nil && h.AccessModel != skydb.RoleBasedAccess {
		response.Err = skyerr.NewInvalidArgument(
			"Cannot assign user role on AcceesModel is not RoleBaseAccess",
			[]string{"roles"})
		return
	}

	authinfo := payload.AuthInfo
	targetUserinfo := &skydb.AuthInfo{}
	payload.DBConn.GetUser(p.ID, targetUserinfo)
	adminRoles, err := payload.DBConn.GetAdminRoles()
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}
	if authinfo.HasAnyRoles(adminRoles) || payload.HasMasterKey() {
		h.updateAuthInfo(targetUserinfo, *p)
	} else if authinfo.ID == targetUserinfo.ID {
		// Make sure no new roles will be added. But some roles can be removed.
		if !authinfo.HasAllRoles(p.Roles) {
			response.Err = skyerr.NewError(skyerr.PermissionDenied, "no permission to add new roles")
			return
		}
		h.updateAuthInfo(targetUserinfo, *p)
	} else {
		response.Err = skyerr.NewError(skyerr.PermissionDenied, "no permission to modify other users")
		return
	}

	if err := payload.DBConn.UpdateUser(targetUserinfo); err != nil {
		if err == skydb.ErrUserDuplicated {
			response.Err = errUserDuplicated
			return
		}
		response.Err = skyerr.MakeError(err)
		return
	}
	response.Result = struct {
		ID       string   `json:"_id"`
		Email    string   `json:"email"`
		Username string   `json:"username"`
		Roles    []string `json:"roles,omitempty"`
	}{
		targetUserinfo.ID,
		targetUserinfo.Email,
		targetUserinfo.Username,
		targetUserinfo.Roles,
	}
}

func (h *UserUpdateHandler) updateAuthInfo(authinfo *skydb.AuthInfo, p userUpdatePayload) skyerr.Error {
	if p.Email != "" {
		authinfo.Email = p.Email
	}
	if p.Username != "" {
		authinfo.Username = p.Username
	}
	if p.Roles != nil {
		authinfo.Roles = p.Roles
	}
	return nil
}

type userLinkPayload struct {
	Username string                 `mapstructure:"username"`
	Email    string                 `mapstructure:"email"`
	Password string                 `mapstructure:"password"`
	Provider string                 `mapstructure:"provider"`
	AuthData map[string]interface{} `mapstructure:"auth_data"`
}

func (payload *userLinkPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *userLinkPayload) Validate() skyerr.Error {
	if payload.Provider == "" {
		return skyerr.NewInvalidArgument("empty provider", []string{"provider"})
	}

	return nil
}

// UserLinkHandler lets user associate third-party accounts with the
// user, with third-party authentication handled by plugin.
type UserLinkHandler struct {
	ProviderRegistry *provider.Registry `inject:"ProviderRegistry"`
	Authenticator    router.Processor   `preprocessor:"authenticator"`
	DBConn           router.Processor   `preprocessor:"dbconn"`
	InjectUser       router.Processor   `preprocessor:"inject_user"`
	InjectDB         router.Processor   `preprocessor:"inject_db"`
	RequireUser      router.Processor   `preprocessor:"require_user"`
	PluginReady      router.Processor   `preprocessor:"plugin_ready"`
	preprocessors    []router.Processor
}

func (h *UserLinkHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
		h.PluginReady,
	}
}

func (h *UserLinkHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *UserLinkHandler) Handle(payload *router.Payload, response *router.Response) {
	p := &userLinkPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	info := skydb.AuthInfo{}

	// Get AuthProvider and authenticates the user
	log.Debugf(`Client requested auth provider: "%v".`, p.Provider)
	authProvider, err := h.ProviderRegistry.GetAuthProvider(p.Provider)
	if err != nil {
		response.Err = skyerr.NewInvalidArgument(err.Error(), []string{"provider"})
		return
	}
	principalID, authData, err := authProvider.Login(payload.Context, p.AuthData)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.InvalidCredentials, "unable to login with the given credentials")
		return
	}
	log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider)

	err = payload.DBConn.GetUserByPrincipalID(principalID, &info)

	if err != nil && err != skydb.ErrUserNotFound {
		// TODO: more error handling here if necessary
		response.Err = skyerr.NewResourceFetchFailureErr("user", p.Username)
		return
	} else if err == nil && info.ID != payload.AuthInfo.ID {
		info.RemoveProvidedAuthData(principalID)
		if err := payload.DBConn.UpdateUser(&info); err != nil {
			response.Err = skyerr.MakeError(err)
			return
		}
	}

	payload.AuthInfo.SetProvidedAuthData(principalID, authData)

	if err := payload.DBConn.UpdateUser(payload.AuthInfo); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}
}
