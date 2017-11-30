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
	"encoding/json"
	"fmt"
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

func TestLoginProviderHandler(t *testing.T) {
	Convey("LoginProviderHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		tokenStore := authtokentest.SingleTokenStore{}
		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()
		txdb := skydbtest.NewMockTxDatabase(db)

		r := handlertest.NewSingleRouteRouter(&LoginProviderHandler{
			TokenStore: &tokenStore,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = txdb
			p.AccessKey = router.MasterAccessKey
		})

		Convey("login provider with non existent principal id", func() {
			resp := r.POST(`
				{
					"principal_id": "non-existent",
					"provider": "skygear",
					"provider_profile": {"name": "chima ceo"},
					"token_response": {"access_token": "token"}
				}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"error": {
						"name": "InvalidCredentials",
						"code": 105,
						"message": "no connected user"
					}
				}
				`)
			So(resp.Code, ShouldEqual, 401)
		})

		Convey("login provider with existent principal id", func() {
			authinfo := skydb.NewAnonymousAuthInfo()
			anHourAgo := timeNow().Add(-1 * time.Hour)
			authinfo.LastSeenAt = &anHourAgo
			conn.CreateAuth(&authinfo)

			userRecordID := skydb.NewRecordID("user", authinfo.ID)
			db.Save(&skydb.Record{
				ID:         userRecordID,
				DatabaseID: db.ID(),
				OwnerID:    authinfo.ID,
				CreatorID:  authinfo.ID,
				UpdaterID:  authinfo.ID,
				CreatedAt:  anHourAgo,
				UpdatedAt:  anHourAgo,
				Data: map[string]interface{}{
					"last_login_at": anHourAgo,
				},
			})

			oauth := skydb.OAuthInfo{
				UserID:          authinfo.ID,
				Provider:        "skygear",
				PrincipalID:     "chima",
				TokenResponse:   map[string]interface{}{"access_token": "token"},
				ProviderProfile: map[string]interface{}{"name": "chima ceo"},
				CreatedAt:       &anHourAgo,
				UpdatedAt:       &anHourAgo,
			}
			conn.CreateOAuthInfo(&oauth)

			resp := r.POST(`
				{
					"principal_id": "chima",
					"provider": "skygear",
					"provider_profile": {"name": "new chima ceo"},
					"token_response": {"access_token": "new token"}
				}`)
			token := tokenStore.Token
			So(token.AccessToken, ShouldNotBeBlank)

			newOAuthInfo := skydb.OAuthInfo{}
			conn.GetOAuthInfo(oauth.Provider, oauth.PrincipalID, &newOAuthInfo)
			So(newOAuthInfo.UserID, ShouldNotBeBlank)

			profile := newOAuthInfo.ProviderProfile
			profileJSON, _ := json.Marshal(&profile)
			So(profileJSON, ShouldEqualJSON, `{"name": "new chima ceo"}`)

			tokenResponse := newOAuthInfo.TokenResponse
			tokenResponseJSON, _ := json.Marshal(&tokenResponse)
			So(tokenResponseJSON, ShouldEqualJSON, `{"access_token": "new token"}`)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`
				{
					"result": {
						"user_id": "%v",
						"profile": {
							"_type": "record",
							"_id": "user/%v",
							"_created_by": "%v",
							"_ownerID": "%v",
							"_updated_by": "%v",
							"_access": null,
							"_created_at": "2006-01-02T14:04:05Z",
							"_updated_at": "2006-01-02T14:04:05Z",
							"last_login_at": {
								"$date": "2006-01-02T14:04:05Z",
								"$type": "date"
							}
						},
						"access_token": "%v",
						"last_login_at": "2006-01-02T14:04:05Z",
						"last_seen_at": "2006-01-02T14:04:05Z"
					}
				}`,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				token.AccessToken,
			))
		})
	})
}

func TestSignupProviderHandler(t *testing.T) {
	Convey("SignupProviderHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		tokenStore := authtokentest.SingleTokenStore{}
		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()
		txdb := skydbtest.NewMockTxDatabase(db)

		r := handlertest.NewSingleRouteRouter(&SignupProviderHandler{
			TokenStore: &tokenStore,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = txdb
			p.AccessKey = router.MasterAccessKey
		})

		Convey("signup provider with non existent principal id", func() {
			resp := r.POST(`
				{
					"principal_id": "chima",
					"provider": "skygear",
					"provider_profile": {"name": "chima ceo", "email": "chima@skygeario.com"},
					"token_response": {"access_token": "token"},
					"profile": {"email": "chima@skygeario.com"}
				}`)

			token := tokenStore.Token
			So(token.AccessToken, ShouldNotBeBlank)
			So(resp.Code, ShouldEqual, 200)

			newOAuthInfo := skydb.OAuthInfo{}
			conn.GetOAuthInfo("skygear", "chima", &newOAuthInfo)
			So(newOAuthInfo.UserID, ShouldNotBeBlank)

			newAuthInfo := skydb.AuthInfo{}
			conn.GetAuth(newOAuthInfo.UserID, &newAuthInfo)
			So(newOAuthInfo.UserID, ShouldNotBeBlank)
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
									"email": "chima@skygeario.com"
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

			profile := newOAuthInfo.ProviderProfile
			profileJSON, _ := json.Marshal(&profile)
			So(profileJSON, ShouldEqualJSON, `{"name": "chima ceo", "email": "chima@skygeario.com"}`)

			tokenResponse := newOAuthInfo.TokenResponse
			tokenResponseJSON, _ := json.Marshal(&tokenResponse)
			So(tokenResponseJSON, ShouldEqualJSON, `{"access_token": "token"}`)
		})

		Convey("signup provider with existent principal id", func() {
			authinfo := skydb.NewAnonymousAuthInfo()
			anHourAgo := timeNow().Add(-1 * time.Hour)
			authinfo.LastSeenAt = &anHourAgo
			conn.CreateAuth(&authinfo)

			userRecordID := skydb.NewRecordID("user", authinfo.ID)
			db.Save(&skydb.Record{
				ID:         userRecordID,
				DatabaseID: db.ID(),
				OwnerID:    authinfo.ID,
				CreatorID:  authinfo.ID,
				UpdaterID:  authinfo.ID,
				CreatedAt:  anHourAgo,
				UpdatedAt:  anHourAgo,
				Data: map[string]interface{}{
					"last_login_at": anHourAgo,
				},
			})

			oauth := skydb.OAuthInfo{
				UserID:          authinfo.ID,
				Provider:        "skygear",
				PrincipalID:     "chima",
				TokenResponse:   map[string]interface{}{"access_token": "token"},
				ProviderProfile: map[string]interface{}{"name": "chima ceo"},
				CreatedAt:       &anHourAgo,
				UpdatedAt:       &anHourAgo,
			}
			conn.CreateOAuthInfo(&oauth)

			resp := r.POST(`
				{
					"principal_id": "chima",
					"provider": "skygear",
					"provider_profile": {"name": "new chima ceo"},
					"token_response": {"access_token": "new token"}
				}`)
			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"error": {
						"name": "InvalidArgument",
						"code": 108,
						"message": "user already connected"
					}
				}`)
		})
	})
}

func TestLinkProviderHandler(t *testing.T) {
	Convey("LinkProviderHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()
		txdb := skydbtest.NewMockTxDatabase(db)

		r := handlertest.NewSingleRouteRouter(&LinkProviderHandler{}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = txdb
			p.AccessKey = router.MasterAccessKey
		})

		Convey("connect provider with non existent user", func() {
			resp := r.POST(`
				{
					"principal_id": "non-existent",
					"provider": "skygear",
					"provider_auth_data": {"name": "chima ceo"},
					"user_id": "non-existent"
				}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"error": {
							"name": "ResourceNotFound",
							"code": 110,
							"message": "user not found"
					}
				}`)
			So(resp.Code, ShouldEqual, 404)
		})

		Convey("provider account already linked with existing user", func() {
			authinfo := skydb.NewAnonymousAuthInfo()
			anHourAgo := timeNow().Add(-1 * time.Hour)
			authinfo.LastSeenAt = &anHourAgo
			conn.CreateAuth(&authinfo)

			userRecordID := skydb.NewRecordID("user", authinfo.ID)
			db.Save(&skydb.Record{
				ID:         userRecordID,
				DatabaseID: db.ID(),
				OwnerID:    authinfo.ID,
				CreatorID:  authinfo.ID,
				UpdaterID:  authinfo.ID,
				CreatedAt:  anHourAgo,
				UpdatedAt:  anHourAgo,
				Data: map[string]interface{}{
					"last_login_at": anHourAgo,
				},
			})

			oauth := skydb.OAuthInfo{
				UserID:          authinfo.ID,
				Provider:        "skygear",
				PrincipalID:     "chima",
				TokenResponse:   map[string]interface{}{"access_token": "token"},
				ProviderProfile: map[string]interface{}{"name": "chima ceo"},
				CreatedAt:       &anHourAgo,
				UpdatedAt:       &anHourAgo,
			}
			conn.CreateOAuthInfo(&oauth)

			resp := r.POST(`
				{
					"principal_id": "chima",
					"provider": "skygear",
					"provider_profile": {"name": "new chima ceo"},
					"token_response": {"access_token": "new token"},
					"user_id": "non-existent"
				}`)
			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"error": {
						"name": "InvalidArgument",
						"code": 108,
						"message": "provider account already linked with existing user"
					}
				}`)
		})

		Convey("user linked to the provider already", func() {
			authinfo := skydb.NewAnonymousAuthInfo()
			anHourAgo := timeNow().Add(-1 * time.Hour)
			authinfo.LastSeenAt = &anHourAgo

			conn.CreateAuth(&authinfo)
			userRecordID := skydb.NewRecordID("user", authinfo.ID)
			db.Save(&skydb.Record{
				ID:         userRecordID,
				DatabaseID: db.ID(),
				OwnerID:    authinfo.ID,
				CreatorID:  authinfo.ID,
				UpdaterID:  authinfo.ID,
				CreatedAt:  anHourAgo,
				UpdatedAt:  anHourAgo,
				Data: map[string]interface{}{
					"last_login_at": anHourAgo,
				},
			})

			oauth := skydb.OAuthInfo{
				UserID:          authinfo.ID,
				Provider:        "skygear",
				PrincipalID:     "chima",
				TokenResponse:   map[string]interface{}{"access_token": "token"},
				ProviderProfile: map[string]interface{}{"name": "chima ceo"},
				CreatedAt:       &anHourAgo,
				UpdatedAt:       &anHourAgo,
			}
			conn.CreateOAuthInfo(&oauth)

			resp := r.POST(fmt.Sprintf(`
				{
					"principal_id": "non-existent",
					"provider": "skygear",
					"provider_profile": {"name": "new chima ceo"},
					"token_response": {"access_token": "new token"},
					"user_id": "%v"
				}`, authinfo.ID))

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"error": {
						"name": "InvalidArgument",
						"code": 108,
						"message": "user linked to the provider already"
					}
				}`)
		})

		Convey("connect provider", func() {
			authinfo := skydb.NewAnonymousAuthInfo()
			anHourAgo := timeNow().Add(-1 * time.Hour)
			authinfo.LastSeenAt = &anHourAgo

			conn.CreateAuth(&authinfo)
			userRecordID := skydb.NewRecordID("user", authinfo.ID)
			db.Save(&skydb.Record{
				ID:         userRecordID,
				DatabaseID: db.ID(),
				OwnerID:    authinfo.ID,
				CreatorID:  authinfo.ID,
				UpdaterID:  authinfo.ID,
				CreatedAt:  anHourAgo,
				UpdatedAt:  anHourAgo,
				Data: map[string]interface{}{
					"last_login_at": anHourAgo,
				},
			})

			resp := r.POST(fmt.Sprintf(`
				{
					"principal_id": "non-existent",
					"provider": "skygear",
					"provider_auth_data": {"name": "chima ceo"},
					"user_id": "%v"
				}`, authinfo.ID))

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result": "OK"}`)
		})
	})
}

func TestUnlinkProviderHandler(t *testing.T) {
	Convey("UnlinkProviderHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()
		txdb := skydbtest.NewMockTxDatabase(db)

		r := handlertest.NewSingleRouteRouter(&UnlinkProviderHandler{}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = txdb
			p.AccessKey = router.MasterAccessKey
		})

		Convey("unlink provider with non existent user", func() {
			resp := r.POST(`
				{
					"provider": "skygear",
					"user_id": "non-existent"
				}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"error": {
							"name": "ResourceNotFound",
							"code": 110,
							"message": "user not found"
					}
				}`)
			So(resp.Code, ShouldEqual, 404)
		})

		Convey("unlink provider success", func() {
			authinfo := skydb.NewAnonymousAuthInfo()
			anHourAgo := timeNow().Add(-1 * time.Hour)
			authinfo.LastSeenAt = &anHourAgo
			conn.CreateAuth(&authinfo)

			userRecordID := skydb.NewRecordID("user", authinfo.ID)
			db.Save(&skydb.Record{
				ID:         userRecordID,
				DatabaseID: db.ID(),
				OwnerID:    authinfo.ID,
				CreatorID:  authinfo.ID,
				UpdaterID:  authinfo.ID,
				CreatedAt:  anHourAgo,
				UpdatedAt:  anHourAgo,
				Data: map[string]interface{}{
					"last_login_at": anHourAgo,
				},
			})

			oauth := skydb.OAuthInfo{
				UserID:          authinfo.ID,
				Provider:        "skygear",
				PrincipalID:     "chima",
				TokenResponse:   map[string]interface{}{"access_token": "token"},
				ProviderProfile: map[string]interface{}{"name": "chima ceo"},
				CreatedAt:       &anHourAgo,
				UpdatedAt:       &anHourAgo,
			}
			conn.CreateOAuthInfo(&oauth)

			oauth = skydb.OAuthInfo{
				UserID:          authinfo.ID,
				Provider:        "cats",
				PrincipalID:     "faseng",
				TokenResponse:   map[string]interface{}{"access_token": "token"},
				ProviderProfile: map[string]interface{}{"name": "faseng cto"},
				CreatedAt:       &anHourAgo,
				UpdatedAt:       &anHourAgo,
			}
			conn.CreateOAuthInfo(&oauth)

			resp := r.POST(fmt.Sprintf(`
				{
					"provider": "skygear",
					"user_id": "%v"
				}`, authinfo.ID))
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result": "OK"}`)

			newOAuthInfo := skydb.OAuthInfo{}
			conn.GetOAuthInfo("skygear", "chima", &newOAuthInfo)
			So(newOAuthInfo.UserID, ShouldBeBlank)

			newOAuthInfo = skydb.OAuthInfo{}
			conn.GetOAuthInfo("cats", "faseng", &newOAuthInfo)
			So(newOAuthInfo.UserID, ShouldNotBeBlank)
		})
	})
}
