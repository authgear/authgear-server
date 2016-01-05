package handler

import (
	log "github.com/Sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
	"github.com/oursky/skygear/provider"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
)

type queryPayload struct {
	Emails []string `json:"emails"`
}

type updatePayload struct {
	Email string `json:"email"`
}

type UserQueryHandler struct {
}

func (h *UserQueryHandler) Handle(payload *router.Payload, response *router.Response) {
	qp := queryPayload{}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &qp,
		TagName: "json",
	})
	if err != nil {
		panic(err)
	}

	if err := mapDecoder.Decode(payload.Data); err != nil {
		response.Err = skyerr.NewError(skyerr.BadRequest, err.Error())
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

type UserUpdateHandler struct {
}

func (h *UserUpdateHandler) Handle(payload *router.Payload, response *router.Response) {
	p := updatePayload{}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &p,
		TagName: "json",
	})
	if err != nil {
		panic(err)
	}

	if err := mapDecoder.Decode(payload.Data); err != nil {
		response.Err = skyerr.NewError(skyerr.BadRequest, err.Error())
		return
	}

	payload.UserInfo.Email = p.Email

	if err := payload.DBConn.UpdateUser(payload.UserInfo); err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}
}

// UserLinkHandler lets user associate third-party accounts with the
// user, with third-party authentication handled by plugin.
type UserLinkHandler struct {
	ProviderRegistry *provider.Registry `inject:"ProviderRegistry"`
}

func (h *UserLinkHandler) Handle(payload *router.Payload, response *router.Response) {
	p := loginPayload{
		AppName: payload.AppName,
		Meta:    payload.Meta,
		Data:    payload.Data,
	}

	if p.Provider() == "" {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, "empty provider")
		return
	}

	info := skydb.UserInfo{}

	// Get AuthProvider and authenticates the user
	log.Debugf(`Client requested auth provider: "%v".`, p.Provider())
	authProvider := h.ProviderRegistry.GetAuthProvider(p.Provider())
	principalID, authData, err := authProvider.Login(p.AuthData())
	if err != nil {
		response.Err = skyerr.NewError(skyerr.InvalidCredentials, "unable to login with the given credentials")
		return
	}
	log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider())

	err = payload.DBConn.GetUserByPrincipalID(principalID, &info)

	if err != nil && err != skydb.ErrUserNotFound {
		// TODO: more error handling here if necessary
		response.Err = skyerr.NewResourceFetchFailureErr("user", p.Username())
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
