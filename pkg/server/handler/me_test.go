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
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/authtoken/authtokentest"
	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skydbtest"

	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMeHandler(t *testing.T) {
	Convey("MeHandler", t, func() {
		conn := skydbtest.NewMapConn()
		lastHour := time.Now().UTC().Add(0 - time.Hour)
		userinfo := skydb.UserInfo{
			ID:             "tester-1",
			Email:          "tester1@example.com",
			Username:       "tester1",
			HashedPassword: []byte("password"),
			Roles: []string{
				"Test",
				"Programmer",
			},
			LastLoginAt: &lastHour,
			LastSeenAt:  &lastHour,
		}
		conn.CreateUser(&userinfo)

		tokenStore := &authtokentest.SingleTokenStore{}
		handler := &MeHandler{
			TokenStore: tokenStore,
		}

		Convey("Get me with user info", func() {
			r := handlertest.NewSingleRouteRouter(handler, func(p *router.Payload) {
				p.Data["access_token"] = "token-1"
				p.UserInfo = &userinfo
				p.DBConn = conn
			})

			resp := r.POST("")
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
        "result": {
          "access_token": "%s",
          "user_id": "tester-1",
          "email": "tester1@example.com",
          "username": "tester1",
          "roles": ["Test", "Programmer"],
          "last_login_at": "%v",
          "last_seen_at": "%v"
        }
      }`,
				tokenStore.Token.AccessToken,
				lastHour.Format(time.RFC3339Nano),
				lastHour.Format(time.RFC3339Nano),
			))
			updateInfo := skydb.UserInfo{}
			conn.GetUser("tester-1", &updateInfo)
			So(updateInfo.LastSeenAt, ShouldNotEqual, lastHour)
		})

		Convey("Get me without user info", func() {
			r := handlertest.NewSingleRouteRouter(handler, func(p *router.Payload) {})
			resp := r.POST("")
			So(resp.Code, ShouldEqual, http.StatusUnauthorized)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
        "error": {
          "name": "NotAuthenticated",
          "code": 101,
          "message": "Authentication is needed to get current user"
        }
      }`)
		})
	})
}
