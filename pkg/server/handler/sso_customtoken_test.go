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
	//"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/authtoken/authtokentest"
	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skydbtest"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
)

func TestSSOCustomTokenLoginHandler(t *testing.T) {
	Convey("SSOCustomTokenLoginHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		tokenStore := authtokentest.SingleTokenStore{}
		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()
		txdb := skydbtest.NewMockTxDatabase(db)

		r := handlertest.NewSingleRouteRouter(&SSOCustomTokenLoginHandler{
			CustomTokenSecret: "ssosecret",
			TokenStore:        &tokenStore,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = txdb
			p.AccessKey = router.ClientAccessKey
		})

		Convey("create user account with custom token", func(c C) {
			tokenString, err := jwt.NewWithClaims(
				jwt.SigningMethodHS256,
				ssoCustomTokenClaims{
					StandardClaims: jwt.StandardClaims{
						IssuedAt:  time.Now().Unix(),
						ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
						Subject:   "otherid1",
					},
					RawProfile: skydb.Data{
						"name": "John Doe",
						"birthday": map[string]interface{}{
							"$date": "2006-01-02T15:04:05Z",
							"$type": "date",
						},
					},
				},
			).SignedString([]byte("ssosecret"))
			So(err, ShouldBeNil)

			resp := r.POST(fmt.Sprintf(`
				{
					"token": "%s"
				}`, tokenString))

			c.Printf("Response: %s", string(resp.Body.Bytes()))
			So(resp.Code, ShouldEqual, 200)

			token := tokenStore.Token
			So(token, ShouldNotBeNil)
			So(token.AccessToken, ShouldNotBeBlank)

			newCustomTokenInfo := skydb.CustomTokenInfo{}
			conn.GetCustomTokenInfo("otherid1", &newCustomTokenInfo)
			So(newCustomTokenInfo.UserID, ShouldNotBeBlank)

			newAuthInfo := skydb.AuthInfo{}
			conn.GetAuth(newCustomTokenInfo.UserID, &newAuthInfo)
			So(newCustomTokenInfo.UserID, ShouldNotBeBlank)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`
				{
					"result": {
						"access_token": "%v",
						"profile": {
							"_access": null,
							"_created_at": "2006-01-02T15:04:05Z",
							"_created_by": "%v",
							"_id": "user/%v",
							"_ownerID": "%v",
							"_type": "record",
							"_updated_at": "2006-01-02T15:04:05Z",
							"_updated_by": "%v",
							"name": "John Doe",
							"birthday": {
								"$date": "2006-01-02T15:04:05Z",
								"$type": "date"
							}
						},
						"user_id": "%v"
					}
				}`,
				token.AccessToken,
				newAuthInfo.ID,
				newAuthInfo.ID,
				newAuthInfo.ID,
				newAuthInfo.ID,
				newAuthInfo.ID,
			))
		})

		Convey("update user account with custom token", func(c C) {
			tokenString, err := jwt.NewWithClaims(
				jwt.SigningMethodHS256,
				ssoCustomTokenClaims{
					StandardClaims: jwt.StandardClaims{
						IssuedAt:  time.Now().Unix(),
						ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
						Subject:   "otherid1",
					},
					RawProfile: skydb.Data{
						"name": "John Doe",
					},
				},
			).SignedString([]byte("ssosecret"))
			So(err, ShouldBeNil)

			now := timeNow()

			authInfo := skydb.NewAnonymousAuthInfo()
			conn.CreateAuth(&authInfo)
			conn.CreateCustomTokenInfo(&skydb.CustomTokenInfo{
				PrincipalID: "otherid1",
				UserID:      authInfo.ID,
				CreatedAt:   &now,
			})
			db.Save(&skydb.Record{
				ID: skydb.NewRecordID("user", authInfo.ID),
				Data: map[string]interface{}{
					"name": "Jane Doe",
				},
			})

			resp := r.POST(fmt.Sprintf(`{"token": "%s"}`, tokenString))

			c.Printf("Response: %s", string(resp.Body.Bytes()))
			So(resp.Code, ShouldEqual, 200)

			fetchedRecord := skydb.Record{}
			db.Get(skydb.NewRecordID("user", authInfo.ID), &fetchedRecord)
			So(fetchedRecord.Data["name"], ShouldEqual, "John Doe")
		})

		Convey("check whether token is invalid", func(c C) {
			tokenString, err := jwt.NewWithClaims(
				jwt.SigningMethodHS256,
				ssoCustomTokenClaims{
					StandardClaims: jwt.StandardClaims{
						IssuedAt:  time.Now().Add(-time.Hour * 1).Unix(),
						ExpiresAt: time.Now().Add(-time.Minute * 30).Unix(),
						Subject:   "otherid1",
					},
					RawProfile: skydb.Data{
						"name": "John Doe",
					},
				},
			).SignedString([]byte("ssosecret"))
			So(err, ShouldBeNil)

			resp := r.POST(fmt.Sprintf(`{"token": "%s"}`, tokenString))

			c.Printf("Response: %s", string(resp.Body.Bytes()))
			So(resp.Code, ShouldEqual, 400)
		})
	})
}
