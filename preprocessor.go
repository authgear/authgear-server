package main

import (
	"errors"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/asset"
	"github.com/oursky/skygear/authtoken"
	"github.com/oursky/skygear/hook"
	"github.com/oursky/skygear/oddb"
	"github.com/oursky/skygear/provider"
	"github.com/oursky/skygear/push"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
)

type apiKeyValidatonPreprocessor struct {
	Key     string
	AppName string
}

func (p apiKeyValidatonPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	apiKey := payload.APIKey()
	if apiKey != p.Key {
		log.Debugf("Invalid APIKEY: %v", apiKey)
		response.Err = skyerr.NewFmt(skyerr.CannotVerifyAPIKey, "Cannot verify api key: %v", apiKey)
		return http.StatusUnauthorized
	}

	payload.AppName = p.AppName

	return http.StatusOK
}

type connPreprocessor struct {
	AppName  string
	DBOpener func(string, string, string) (oddb.Conn, error)
	DBImpl   string
	Option   string
}

func (p connPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	log.Debugf("Opening DBConn: {%v %v %v}", p.DBImpl, p.AppName, p.Option)

	conn, err := p.DBOpener(p.DBImpl, p.AppName, p.Option)
	if err != nil {
		response.Err = err
		return http.StatusServiceUnavailable
	}
	payload.DBConn = conn

	log.Debugf("Get DB OK")

	return http.StatusOK
}

type hookRegistryPreprocessor struct {
	Registry *hook.Registry
}

func (p hookRegistryPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	payload.HookRegistry = p.Registry
	return http.StatusOK
}

type providerRegistryPreprocessor struct {
	Registry *provider.Registry
}

func (p providerRegistryPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	payload.ProviderRegistry = p.Registry
	return http.StatusOK
}

type tokenStorePreprocessor struct {
	authtoken.Store
}

func (p tokenStorePreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	payload.TokenStore = p.Store
	return http.StatusOK
}

// UserAuthenticator provides preprocess method to authenicate a user
// with access token or non-login user without api key.
type userAuthenticator struct {
	// These two fields are for non-login user
	APIKey  string
	AppName string
}

func (author *userAuthenticator) Preprocess(payload *router.Payload, response *router.Response) int {
	tokenString := payload.AccessToken()
	if tokenString == "" {
		apiKey := payload.APIKey()
		if apiKey != author.APIKey {
			if author.APIKey != "" && apiKey == "" {
				// if a non-empty api key is set and we received empty
				// api key and access token, then client request
				// has no authentication information
				response.Err = skyerr.NewFmt(skyerr.AuthenticationInfoIncorrectErr, "Both api key and access token are empty")
			} else {
				response.Err = skyerr.NewFmt(skyerr.CannotVerifyAPIKey, "Cannot verify api key: `%v`", apiKey)
			}
			return http.StatusUnauthorized
		}

		payload.AppName = author.AppName
	} else {
		store := payload.TokenStore
		token := authtoken.Token{}

		if err := store.Get(tokenString, &token); err != nil {
			if _, ok := err.(*authtoken.NotFoundError); ok {
				log.WithFields(log.Fields{
					"token": tokenString,
					"err":   err,
				}).Infoln("Token not found")

				response.Err = skyerr.ErrAuthFailure
			} else {
				response.Err = err
			}
			return http.StatusUnauthorized
		}

		payload.AppName = token.AppName
		payload.UserInfoID = token.UserInfoID
	}

	return http.StatusOK
}

func injectUserIfPresent(payload *router.Payload, response *router.Response) int {
	if payload.UserInfoID == "" {
		log.Debugln("injectUser: empty UserInfoID, skipping")
		return http.StatusOK
	}

	conn := payload.DBConn
	userinfo := oddb.UserInfo{}
	if err := conn.GetUser(payload.UserInfoID, &userinfo); err != nil {
		log.Errorf("Cannot find UserInfo.ID = %#v\n", payload.UserInfoID)
		response.Err = err
		return http.StatusInternalServerError
	}

	payload.UserInfo = &userinfo

	return http.StatusOK
}

type assetStorePreprocessor struct {
	Store asset.Store
}

func (p assetStorePreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	payload.AssetStore = p.Store
	return http.StatusOK
}

func injectDatabase(payload *router.Payload, response *router.Response) int {
	conn := payload.DBConn

	databaseID, _ := payload.Data["database_id"].(string)
	switch databaseID {
	case "_public":
		payload.Database = conn.PublicDB()
	case "_private":
		if payload.UserInfo != nil {
			payload.Database = conn.PrivateDB(payload.UserInfo.ID)
		} else {
			response.Err = errors.New("Authentication is needed for private DB access")
			return http.StatusUnauthorized
		}
	}

	return http.StatusOK
}

func requireUserForWrite(payload *router.Payload, response *router.Response) int {
	if payload.UserInfo == nil {
		response.Err = skyerr.ErrWriteDenied
		return http.StatusUnauthorized
	}

	return http.StatusOK
}

type notificationPreprocessor struct {
	NotificationSender push.Sender
}

func (p notificationPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	if p.NotificationSender == nil {
		response.Err = errors.New("Unable to send push notification because APNS is not configured or there was a problem configuring the APNS.\n")
		return http.StatusInternalServerError
	}

	payload.NotificationSender = p.NotificationSender
	return http.StatusOK
}
