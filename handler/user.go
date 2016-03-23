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
	log "github.com/Sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
	"github.com/skygeario/skygear-server/plugin/provider"
	"github.com/skygeario/skygear-server/router"
	"github.com/skygeario/skygear-server/skydb"
	"github.com/skygeario/skygear-server/skyerr"
)

type queryPayload struct {
	Emails []string `mapstructure:"emails"`
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
	preprocessors []router.Processor
}

func (h *UserQueryHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
	}
}

func (h *UserQueryHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *UserQueryHandler) Handle(payload *router.Payload, response *router.Response) {
	qp := &queryPayload{}
	skyErr := qp.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	userinfos, err := payload.DBConn.QueryUser(qp.Emails)
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}

	results := make([]interface{}, len(userinfos))
	for i, userinfo := range userinfos {
		results[i] = map[string]interface{}{
			"id":   userinfo.ID,
			"type": "user",
			"data": struct {
				ID       string `json:"_id"`
				Email    string `json:"email"`
				Username string `json:"username"`
			}{userinfo.ID, userinfo.Email, userinfo.Username},
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
	preprocessors []router.Processor
}

func (h *UserUpdateHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
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

	userinfo := payload.UserInfo
	targetUserinfo := &skydb.UserInfo{}
	payload.DBConn.GetUser(p.ID, targetUserinfo)
	adminRoles, err := payload.DBConn.GetAdminRoles()
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}
	if userinfo.HasAnyRoles(adminRoles) {
		h.updateUserInfo(targetUserinfo, *p)
	} else if userinfo.ID == targetUserinfo.ID {
		if !userinfo.HasAllRoles(p.Roles) {
			response.Err = skyerr.NewError(skyerr.PermissionDenied, "no permission to add new roles")
			return
		}
		h.updateUserInfo(targetUserinfo, *p)
	} else {
		response.Err = skyerr.NewError(skyerr.PermissionDenied, "no permission to modify other users")
		return
	}

	if err := payload.DBConn.UpdateUser(targetUserinfo); err != nil {
		response.Err = skyerr.NewUnknownErr(err)
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

func (h *UserUpdateHandler) updateUserInfo(userinfo *skydb.UserInfo, p userUpdatePayload) skyerr.Error {
	if p.Email != "" {
		userinfo.Email = p.Email
	}
	if p.Username != "" {
		userinfo.Username = p.Username
	}
	if p.Roles != nil {
		userinfo.Roles = p.Roles
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
	preprocessors    []router.Processor
}

func (h *UserLinkHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
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

	info := skydb.UserInfo{}

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

	err = payload.DBConn.GetUserByPrincipalID(principalID, &info)

	if err != nil && err != skydb.ErrUserNotFound {
		// TODO: more error handling here if necessary
		response.Err = skyerr.NewResourceFetchFailureErr("user", p.Username)
		return
	} else if err == nil && info.ID != payload.UserInfo.ID {
		info.RemoveProvidedAuthData(principalID)
		if err := payload.DBConn.UpdateUser(&info); err != nil {
			response.Err = skyerr.NewUnknownErr(err)
			return
		}
	}

	payload.UserInfo.SetProvidedAuthData(principalID, authData)

	if err := payload.DBConn.UpdateUser(payload.UserInfo); err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}
}
