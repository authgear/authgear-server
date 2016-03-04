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
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/oursky/skygear/authtoken"
	"github.com/oursky/skygear/authtoken/authtokentest"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
)

func TestAccessKeyValidationPreprocessor(t *testing.T) {
	Convey("test access key validation preprocessor", t, func() {
		pp := AccessKeyValidationPreprocessor{
			ClientKey: "client-key",
			MasterKey: "master-key",
			AppName:   "app-name",
		}

		payload := &router.Payload{
			Data: map[string]interface{}{},
			Meta: map[string]interface{}{},
		}
		resp := &router.Response{}

		Convey("test client key", func() {
			payload.Data["api_key"] = "client-key"
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusOK)
			So(payload.AccessKey, ShouldEqual, router.ClientAccessKey)
			So(payload.AppName, ShouldEqual, "app-name")
			So(resp.Err, ShouldBeNil)
		})

		Convey("test master key", func() {
			payload.Data["api_key"] = "master-key"
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusOK)
			So(payload.AccessKey, ShouldEqual, router.MasterAccessKey)
			So(payload.AppName, ShouldEqual, "app-name")
			So(resp.Err, ShouldBeNil)
		})

		Convey("test wrong key", func() {
			payload.Data["api_key"] = "wrong-key"
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusUnauthorized)
			So(payload.AccessKey, ShouldEqual, router.NoAccessKey)
			So(resp.Err, ShouldNotBeNil)
			So(resp.Err.Code(), ShouldEqual, skyerr.AccessKeyNotAccepted)
		})

		Convey("test no key", func() {
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusUnauthorized)
			So(payload.AccessKey, ShouldEqual, router.NoAccessKey)
			So(resp.Err, ShouldNotBeNil)
			So(resp.Err.Code(), ShouldEqual, skyerr.NotAuthenticated)
		})
	})
}

func TestUserAuthenticator(t *testing.T) {
	Convey("test access user authenticator for api key", t, func() {
		pp := UserAuthenticator{
			ClientKey:  "client-key",
			MasterKey:  "master-key",
			AppName:    "app-name",
			TokenStore: &authtokentest.SingleTokenStore{},
		}

		payload := &router.Payload{
			Data: map[string]interface{}{},
			Meta: map[string]interface{}{},
		}
		resp := &router.Response{}

		Convey("test client key", func() {
			payload.Data["api_key"] = "client-key"
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusOK)
			So(payload.AccessKey, ShouldEqual, router.ClientAccessKey)
			So(payload.AppName, ShouldEqual, "app-name")
			So(resp.Err, ShouldBeNil)
		})

		Convey("test master key", func() {
			payload.Data["api_key"] = "master-key"
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusOK)
			So(payload.AccessKey, ShouldEqual, router.MasterAccessKey)
			So(payload.AppName, ShouldEqual, "app-name")
			So(resp.Err, ShouldBeNil)
		})

		Convey("test wrong key", func() {
			payload.Data["api_key"] = "wrong-key"
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusUnauthorized)
			So(payload.AccessKey, ShouldEqual, router.NoAccessKey)
			So(resp.Err, ShouldNotBeNil)
			So(resp.Err.Code(), ShouldEqual, skyerr.AccessKeyNotAccepted)
		})

		Convey("test no key", func() {
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusUnauthorized)
			So(payload.AccessKey, ShouldEqual, router.NoAccessKey)
			So(resp.Err, ShouldNotBeNil)
			So(resp.Err.Code(), ShouldEqual, skyerr.NotAuthenticated)
		})
	})

	Convey("test access user authenticator with master key", t, func() {
		pp := UserAuthenticator{
			ClientKey:  "client-key",
			MasterKey:  "master-key",
			AppName:    "app-name",
			TokenStore: &authtokentest.SingleTokenStore{},
		}

		payload := &router.Payload{
			Data: map[string]interface{}{},
			Meta: map[string]interface{}{},
		}
		resp := &router.Response{}

		Convey("impersonate a user", func() {
			payload.Data["api_key"] = "master-key"
			payload.Data["_user_id"] = "user-id"
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusOK)
			So(payload.AccessKey, ShouldEqual, router.MasterAccessKey)
			So(payload.AppName, ShouldEqual, "app-name")
			So(payload.UserInfoID, ShouldEqual, "user-id")
			So(resp.Err, ShouldBeNil)
		})

		Convey("impersonate a user without master key", func() {
			payload.Data["api_key"] = "client-key"
			payload.Data["_user_id"] = "user-id"
			So(payload.UserInfoID, ShouldNotEqual, "user-id")
		})
	})

	Convey("test access user authenticator for access token", t, func() {
		pp := UserAuthenticator{
			ClientKey:  "client-key",
			MasterKey:  "master-key",
			AppName:    "app-name",
			TokenStore: &authtokentest.SingleTokenStore{},
		}

		payload := &router.Payload{
			Data: map[string]interface{}{},
			Meta: map[string]interface{}{},
		}
		resp := &router.Response{}

		Convey("test valid token", func() {
			token := authtoken.New("app-name", "user-id", time.Time{})
			pp.TokenStore.Put(&token)
			payload.Data["access_token"] = token.AccessToken
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusOK)
			So(payload.AppName, ShouldEqual, "app-name")
			So(payload.UserInfoID, ShouldEqual, "user-id")
			So(resp.Err, ShouldBeNil)
		})

		Convey("test expired token", func() {
			token := authtoken.New("app-name", "user-id", time.Now())
			// do not put it in the test token store to simulate expired token
			payload.Data["access_token"] = token.AccessToken
			So(pp.Preprocess(payload, resp), ShouldEqual, http.StatusUnauthorized)
			So(resp.Err, ShouldNotBeNil)
			So(resp.Err.Code(), ShouldEqual, skyerr.AccessTokenNotAccepted)
		})
	})
}
