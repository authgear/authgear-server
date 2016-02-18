package handler

import (
	log "github.com/Sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
	"github.com/oursky/skygear/plugin/provider"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
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
	Email string   `mapstructure:"email"`
	Roles []string `mapstructure:"roles"`
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
	userinfo.Email = p.Email
	adminRoles, err := payload.DBConn.GetAdminRoles()
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}
	skyErr = h.updateRoles(userinfo, adminRoles, p.Roles)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	if err := payload.DBConn.UpdateUser(userinfo); err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}
	response.Result = struct {
		ID       string   `json:"_id"`
		Email    string   `json:"email"`
		Username string   `json:"username"`
		Roles    []string `json:"roles,omitempty"`
	}{
		userinfo.ID,
		userinfo.Email,
		userinfo.Username,
		userinfo.Roles,
	}
}

func (h *UserUpdateHandler) updateRoles(userinfo *skydb.UserInfo, admins []string, roles []string) skyerr.Error {
	if userinfo.HasAnyRoles(admins) {
		userinfo.Roles = roles
		return nil
	}
	if userinfo.HasAllRoles(roles) {
		userinfo.Roles = roles
		return nil
	}
	return skyerr.NewError(skyerr.PermissionDenied, "no permission to add new roles")
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
	authProvider := h.ProviderRegistry.GetAuthProvider(p.Provider)
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
