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
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"

	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMeHandler(t *testing.T) {
	Convey("MeHandler", t, func() {
		sampleUserInfo := skydb.UserInfo{
			ID:             "tester-1",
			Email:          "tester1@example.com",
			Username:       "tester1",
			HashedPassword: []byte("password"),
			Roles: []string{
				"Test",
				"Programmer",
			},
		}

		Convey("Get me with user info", func() {
			r := handlertest.NewSingleRouteRouter(&MeHandler{}, func(p *router.Payload) {
				p.Data["access_token"] = "token-1"
				p.UserInfo = &sampleUserInfo
			})

			resp := r.POST("")
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
        "result": {
          "access_token": "token-1",
          "user_id": "tester-1",
          "email": "tester1@example.com",
          "username": "tester1",
          "roles": ["Test", "Programmer"]
        }
      }`)
		})

		Convey("Get me without user info", func() {
			r := handlertest.NewSingleRouteRouter(&MeHandler{}, func(p *router.Payload) {})
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
