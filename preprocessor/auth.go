package preprocessor

import (
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/authtoken"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
	"golang.org/x/net/context"
)

type AccessKeyValidationPreprocessor struct {
	Key     string
	AppName string
}

func (p AccessKeyValidationPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	apiKey := payload.APIKey()
	if apiKey != p.Key {
		log.Debugf("Invalid APIKEY: %v", apiKey)
		response.Err = skyerr.NewErrorf(skyerr.AccessKeyNotAccepted, "Cannot verify api key: %v", apiKey)
		return http.StatusUnauthorized
	}

	payload.AppName = p.AppName

	return http.StatusOK
}

// UserAuthenticator provides preprocess method to authenicate a user
// with access token or non-login user without api key.
type UserAuthenticator struct {
	// These two fields are for non-login user
	APIKey     string
	AppName    string
	TokenStore authtoken.Store
}

func (author *UserAuthenticator) Preprocess(payload *router.Payload, response *router.Response) int {
	tokenString := payload.AccessToken()
	if tokenString == "" {
		apiKey := payload.APIKey()
		if apiKey != author.APIKey {
			if author.APIKey != "" && apiKey == "" {
				// if a non-empty api key is set and we received empty
				// api key and access token, then client request
				// has no authentication information
				response.Err = skyerr.NewErrorf(skyerr.NotAuthenticated, "Both api key and access token are empty")
			} else {
				response.Err = skyerr.NewErrorf(skyerr.AccessKeyNotAccepted, "Cannot verify api key: `%v`", apiKey)
			}
			return http.StatusUnauthorized
		}

		payload.AppName = author.AppName
	} else {
		store := author.TokenStore
		token := authtoken.Token{}

		if err := store.Get(tokenString, &token); err != nil {
			if _, ok := err.(*authtoken.NotFoundError); ok {
				log.WithFields(log.Fields{
					"token": tokenString,
					"err":   err,
				}).Infoln("Token not found")

				response.Err = skyerr.NewError(skyerr.AccessTokenNotAccepted, "token expired")
			} else {
				response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
			}
			return http.StatusUnauthorized
		}

		payload.AppName = token.AppName
		payload.UserInfoID = token.UserInfoID
		payload.Context = context.WithValue(payload.Context, "UserID", token.UserInfoID)
	}

	return http.StatusOK
}
