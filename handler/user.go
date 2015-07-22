package handler

import (
	"errors"
	log "github.com/Sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

type queryPayload struct {
	Emails []string `json:"emails"`
}

type updatePayload struct {
	Email string `json:"email"`
}

func UserQueryHandler(payload *router.Payload, response *router.Response) {
	qp := queryPayload{}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &qp,
		TagName: "json",
	})
	if err != nil {
		panic(err)
	}

	if err := mapDecoder.Decode(payload.Data); err != nil {
		response.Err = oderr.NewRequestInvalidErr(err)
		return
	}

	userinfos, err := payload.DBConn.QueryUser(qp.Emails)
	if err != nil {
		response.Err = oderr.NewUnknownErr(err)
		return
	}

	results := make([]interface{}, len(userinfos))
	for i, userinfo := range userinfos {
		results[i] = map[string]interface{}{
			"id":   userinfo.ID,
			"type": "user",
			"data": struct {
				ID    string `json:"_id"`
				Email string `json:"email"`
			}{userinfo.ID, userinfo.Email},
		}
	}
	response.Result = results
}

func UserUpdateHandler(payload *router.Payload, response *router.Response) {
	p := updatePayload{}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &p,
		TagName: "json",
	})
	if err != nil {
		panic(err)
	}

	if err := mapDecoder.Decode(payload.Data); err != nil {
		response.Err = oderr.NewRequestInvalidErr(err)
		return
	}

	payload.UserInfo.Email = p.Email

	if err := payload.DBConn.UpdateUser(payload.UserInfo); err != nil {
		response.Err = oderr.NewUnknownErr(err)
		return
	}
}

// UserLinkHandler lets user associate third-party accounts with the
// user, with third-party authentication handled by plugin.
func UserLinkHandler(payload *router.Payload, response *router.Response) {
	p := loginPayload{
		AppName: payload.AppName,
		Meta:    payload.Meta,
		Data:    payload.Data,
	}

	if p.Provider() == "" {
		response.Err = oderr.NewRequestInvalidErr(errors.New("empty provider"))
		return
	}

	info := oddb.UserInfo{}

	// Get AuthProvider and authenticates the user
	log.Debugf(`Client requested auth provider: "%v".`, p.Provider())
	authProvider := payload.ProviderRegistry.GetAuthProvider(p.Provider())
	principalID, authData, err := authProvider.Login(p.AuthData())
	if err != nil {
		response.Err = oderr.ErrAuthFailure
		return
	}
	log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider())

	err = payload.DBConn.GetUserByPrincipalID(principalID, &info)

	if err != nil && err != oddb.ErrUserNotFound {
		// TODO: more error handling here if necessary
		response.Err = oderr.NewResourceFetchFailureErr("user", p.UserID())
		return
	} else if err == nil && info.ID != payload.UserInfo.ID {
		response.Err = oderr.NewRequestInvalidErr(errors.New("already associated with another user"))
		return
	}

	payload.UserInfo.SetProvidedAuthData(principalID, authData)

	if err := payload.DBConn.UpdateUser(payload.UserInfo); err != nil {
		response.Err = oderr.NewUnknownErr(err)
		return
	}
}
