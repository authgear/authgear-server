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

package preprocessor

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/skygeario/skygear-server/pkg/server/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func checkRequestAccessKey(payload *router.Payload, clientKey string, masterKey string) skyerr.Error {
	apiKey := payload.APIKey()
	if masterKey != "" && apiKey == masterKey {
		payload.AccessKey = router.MasterAccessKey
	} else if clientKey != "" && apiKey == clientKey {
		payload.AccessKey = router.ClientAccessKey
	} else if apiKey == "" {
		payload.AccessKey = router.NoAccessKey
	} else {
		return skyerr.NewErrorf(skyerr.AccessKeyNotAccepted, "Cannot verify api key: `%v`", apiKey)
	}
	payload.Context = context.WithValue(payload.Context, router.AccessKeyTypeContextKey, payload.AccessKey)
	return nil
}

// AccessKeyValidationPreprocessor provides preprocess method to check the
// API key of the request.
type AccessKeyValidationPreprocessor struct {
	ClientKey string
	MasterKey string
	AppName   string
}

func (p AccessKeyValidationPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	if err := checkRequestAccessKey(payload, p.ClientKey, p.MasterKey); err != nil {
		response.Err = err
		return http.StatusUnauthorized
	}

	if payload.AccessKey == router.NoAccessKey {
		response.Err = skyerr.NewErrorf(skyerr.NotAuthenticated, "Api key is empty")
		return http.StatusUnauthorized
	}

	payload.AppName = p.AppName
	return http.StatusOK
}

// UserAuthenticator provides preprocess method to authenicate a user
// with access token or non-login user without api key.
type UserAuthenticator struct {
	ClientKey  string
	MasterKey  string
	AppName    string
	TokenStore authtoken.Store
}

func (p *UserAuthenticator) Preprocess(payload *router.Payload, response *router.Response) int {
	if err := checkRequestAccessKey(payload, p.ClientKey, p.MasterKey); err != nil {
		response.Err = err
		return http.StatusUnauthorized
	}

	// If payload contains an access token, check whether if the access
	// token is valid. API Key is not required if there is valid access token.
	if tokenString := payload.AccessTokenString(); tokenString != "" {
		store := p.TokenStore
		token := authtoken.Token{}

		if err := store.Get(tokenString, &token); err != nil {
			if _, ok := err.(*authtoken.NotFoundError); ok {
				log.WithFields(logrus.Fields{
					"token": tokenString,
					"err":   err,
				}).Infoln("Token not found")

				response.Err = skyerr.NewError(skyerr.AccessTokenNotAccepted, "token does not exist or it has expired")
			} else {
				response.Err = skyerr.NewError(skyerr.UnexpectedError, err.Error())
			}
			return http.StatusUnauthorized
		}

		payload.AppName = token.AppName
		payload.UserInfoID = token.UserInfoID
		payload.Context = context.WithValue(payload.Context, router.UserIDContextKey, token.UserInfoID)
		payload.AccessToken = token
		return http.StatusOK
	}

	if payload.AccessKey == router.NoAccessKey {
		response.Err = skyerr.NewErrorf(skyerr.NotAuthenticated, "Both api key and access token are empty")
		return http.StatusUnauthorized
	}

	// For master access key, it is possible to impersonate any user of
	// the caller's choosing.
	if payload.HasMasterKey() {
		if userID, ok := payload.Data["_user_id"].(string); ok {
			payload.UserInfoID = userID
			payload.Context = context.WithValue(payload.Context, router.UserIDContextKey, userID)
		}
	}

	payload.AppName = p.AppName
	return http.StatusOK
}
