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
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		now := timeNow()
		anHourAgo := now.Add(-1 * time.Hour)
		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()
		authinfo := skydb.AuthInfo{
			ID:             "tester-1",
			HashedPassword: []byte("password"),
			Roles: []string{
				"Test",
				"Programmer",
			},
			LastSeenAt: &anHourAgo,
		}
		user := skydb.Record{
			ID: skydb.RecordID{
				Type: "user",
				Key:  "tester-1",
			},
			CreatedAt: anHourAgo,
			CreatorID: "tester-1",
			UpdatedAt: anHourAgo,
			UpdaterID: "tester-1",
			Data: skydb.Data{
				"username":      "tester1",
				"email":         "tester1@example.com",
				"last_login_at": anHourAgo,
			},
		}
		conn.CreateAuth(&authinfo)
		db.Save(&user)

		tokenStore := &authtokentest.SingleTokenStore{}
		handler := &MeHandler{
			TokenStore: tokenStore,
		}

		Convey("Get me with user info", func(c C) {
			r := handlertest.NewSingleRouteRouter(handler, func(p *router.Payload) {
				p.Data["access_token"] = "token-1"
				p.AuthInfo = &authinfo
				p.DBConn = conn
				p.Database = db
				p.User = &user
			})

			resp := r.POST("")
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(
				resp.Body.Bytes(),
				ShouldEqualJSON,
				fmt.Sprintf(`
					{
						"result": {
							"access_token": "%s",
							"user_id": "tester-1",
							"profile": {
								"_type": "record",
								"_id": "user/tester-1",
								"_recordType": "user",
								"_recordID": "tester-1",
								"_access": null,
								"_created_at": "2006-01-02T14:04:05Z",
								"_created_by": "tester-1",
								"_updated_at": "2006-01-02T14:04:05Z",
								"_updated_by": "tester-1",
								"email": "tester1@example.com",
								"username": "tester1",
								"last_login_at": {
									"$type": "date",
									"$date": "2006-01-02T14:04:05Z"
								}
							},
							"roles": ["Test", "Programmer"],
							"last_login_at": "2006-01-02T14:04:05Z",
							"last_seen_at": "2006-01-02T14:04:05Z"
						}
					}`,
					tokenStore.Token.AccessToken,
				),
			)

			updateInfo := skydb.AuthInfo{}
			conn.GetAuth("tester-1", &updateInfo)
			So(updateInfo.LastSeenAt, ShouldResemble, &now)
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
