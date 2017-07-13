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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/authtoken/authtokentest"
	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/plugin/provider"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/mock_skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skydbtest"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func tempDir() string {
	dir, err := ioutil.TempDir("", "skygear.skydb.handler.auth.test")
	if err != nil {
		panic(err)
	}
	return dir
}

func MakeEqualPredicateAssertion(key string, value string) func(predicate *skydb.Predicate) {
	return func(predicate *skydb.Predicate) {
		So(predicate.Operator, ShouldEqual, skydb.Equal)

		keyExp := predicate.Children[0].(skydb.Expression)
		valueExp := predicate.Children[1].(skydb.Expression)

		So(keyExp.Type, ShouldEqual, skydb.KeyPath)
		So(keyExp.Value, ShouldEqual, key)

		So(valueExp.Type, ShouldEqual, skydb.Literal)
		So(valueExp.Value, ShouldEqual, value)
	}
}

func MakeUsernameEmailQueryAssertion(username string, email string) func(query *skydb.Query) {
	return func(query *skydb.Query) {
		So(query.Type, ShouldEqual, "user")

		predicate := query.Predicate
		So(predicate.Operator, ShouldEqual, skydb.And)

		expectedChildrenCount := 0
		if username != "" {
			expectedChildrenCount = expectedChildrenCount + 1
		}

		if email != "" {
			expectedChildrenCount = expectedChildrenCount + 1
		}

		So(len(predicate.Children), ShouldEqual, expectedChildrenCount)

		for _, child := range predicate.Children {
			childPredicate := child.(skydb.Predicate)
			keyExp := childPredicate.Children[0].(skydb.Expression)
			if keyExp.Type == skydb.KeyPath && keyExp.Value == "username" {
				MakeEqualPredicateAssertion("username", username)(&childPredicate)
			} else if keyExp.Type == skydb.KeyPath && keyExp.Value == "email" {
				MakeEqualPredicateAssertion("email", email)(&childPredicate)
			} else {
				panic(fmt.Sprintf("Unexpected keypath"))
			}
		}
	}
}

func MakeUserRecordAssertion(authData skydb.AuthData) func(record *skydb.Record) {
	return func(record *skydb.Record) {
		So(record.ID.Type, ShouldEqual, "user")
		So(record.Data["username"], ShouldEqual, authData["username"])
		So(record.Data["email"], ShouldEqual, authData["email"])
	}
}

func ExpectDBSaveUserWithAuthData(db *mock_skydb.MockTxDatabase, authData skydb.AuthData) {
	userRecordSchema := skydb.RecordSchema{
		"username": skydb.FieldType{Type: skydb.TypeString},
		"email":    skydb.FieldType{Type: skydb.TypeString},
	}
	skydbtest.ExpectDBSaveUser(db, userRecordSchema, MakeUserRecordAssertion(authData))
}

// Seems like a memory imlementation of skydb will make tests
// faster and easier

func TestSignupHandler(t *testing.T) {
	Convey("SignupHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		conn := skydbtest.NewMapConn()
		tokenStore := authtokentest.SingleTokenStore{}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := mock_skydb.NewMockTxDatabase(ctrl)
		handler := &SignupHandler{
			TokenStore: &tokenStore,
		}

		Convey("sign up new account", func() {
			db.EXPECT().
				Query(gomock.Any()).
				Do(MakeUsernameEmailQueryAssertion("john.doe", "john.doe@example.com")).
				Return(skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{})), nil).
				AnyTimes()
			txBegin := db.EXPECT().Begin().AnyTimes()
			db.EXPECT().Commit().After(txBegin)

			ExpectDBSaveUserWithAuthData(db, skydb.AuthData{"username": "john.doe", "email": "john.doe@example.com"})

			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"username": "john.doe",
						"email":    "john.doe@example.com",
					},
					"password": "secret",
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Result, ShouldHaveSameTypeAs, AuthResponse{})
			authResp := resp.Result.(AuthResponse)
			So(authResp.Profile.ID, ShouldResemble, skydb.NewRecordID("user", authResp.UserID))
			So(authResp.Profile.DatabaseID, ShouldResemble, "_public")
			So(authResp.Profile.OwnerID, ShouldResemble, authResp.UserID)
			So(authResp.Profile.CreatorID, ShouldResemble, authResp.UserID)
			So(authResp.Profile.UpdaterID, ShouldResemble, authResp.UserID)
			So(authResp.Profile.CreatedAt, ShouldResemble, time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(authResp.Profile.UpdatedAt, ShouldResemble, time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(authResp.Profile.Data, ShouldResemble, skydb.Data{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			})
			So(authResp.AccessToken, ShouldNotBeEmpty)
			So(authResp.LastLoginAt, ShouldNotBeEmpty)
			So(authResp.LastSeenAt, ShouldNotBeEmpty)
			token := tokenStore.Token
			So(token.AuthInfoID, ShouldEqual, authResp.UserID)
			So(token.AccessToken, ShouldNotBeEmpty)

			authinfo := &skydb.AuthInfo{}
			err := conn.GetAuth(authResp.UserID, authinfo)
			So(err, ShouldBeNil)
			So(authinfo.Roles, ShouldBeNil)
		})

		Convey("sign up new account with profile", func() {
			db.EXPECT().
				Query(gomock.Any()).
				Do(MakeUsernameEmailQueryAssertion("john.doe", "john.doe@example.com")).
				Return(skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{})), nil).
				AnyTimes()
			txBegin := db.EXPECT().Begin().AnyTimes()
			db.EXPECT().Commit().After(txBegin)

			skydbtest.ExpectDBSaveUser(db, skydb.RecordSchema{
				"username": skydb.FieldType{Type: skydb.TypeString},
				"email":    skydb.FieldType{Type: skydb.TypeString},
				"nickname": skydb.FieldType{Type: skydb.TypeString},
				"number":   skydb.FieldType{Type: skydb.TypeNumber},
				"boolean":  skydb.FieldType{Type: skydb.TypeBoolean},
			}, MakeUserRecordAssertion(skydb.AuthData{
				"username": "john.doe",
				"email":    "john.doe@example.com",
				"nickname": "iamyourfather",
				"number":   float64(0),
				"boolean":  false,
			}))

			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"username": "john.doe",
						"email":    "john.doe@example.com",
					},
					"password": "secret",
					"profile": skydb.Data{
						"nickname": "iamyourfather",
						"number":   float64(0),
						"boolean":  false,
					},
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Result, ShouldHaveSameTypeAs, AuthResponse{})
			authResp := resp.Result.(AuthResponse)
			So(authResp.Profile.ID, ShouldResemble, skydb.NewRecordID("user", authResp.UserID))
			So(authResp.Profile.DatabaseID, ShouldResemble, "_public")
			So(authResp.Profile.OwnerID, ShouldResemble, authResp.UserID)
			So(authResp.Profile.CreatorID, ShouldResemble, authResp.UserID)
			So(authResp.Profile.UpdaterID, ShouldResemble, authResp.UserID)
			So(authResp.Profile.CreatedAt, ShouldResemble, time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(authResp.Profile.UpdatedAt, ShouldResemble, time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(authResp.Profile.Data, ShouldResemble, skydb.Data{
				"username": "john.doe",
				"email":    "john.doe@example.com",
				"nickname": "iamyourfather",
				"number":   float64(0),
				"boolean":  false,
			})
			So(authResp.AccessToken, ShouldNotBeEmpty)
			So(authResp.LastLoginAt, ShouldNotBeEmpty)
			So(authResp.LastSeenAt, ShouldNotBeEmpty)
			token := tokenStore.Token
			So(token.AuthInfoID, ShouldEqual, authResp.UserID)
			So(token.AccessToken, ShouldNotBeEmpty)

			authinfo := &skydb.AuthInfo{}
			err := conn.GetAuth(authResp.UserID, authinfo)
			So(err, ShouldBeNil)
			So(authinfo.Roles, ShouldBeNil)
		})

		Convey("sign up new account with profile, with auth data key but not duplicated", func() {
			db.EXPECT().
				Query(gomock.Any()).
				Do(MakeUsernameEmailQueryAssertion("john.doe", "")).
				Return(skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{})), nil).
				AnyTimes()
			txBegin := db.EXPECT().Begin().AnyTimes()
			db.EXPECT().Commit().After(txBegin)

			skydbtest.ExpectDBSaveUser(db, skydb.RecordSchema{
				"username": skydb.FieldType{Type: skydb.TypeString},
				"email":    skydb.FieldType{Type: skydb.TypeString},
				"nickname": skydb.FieldType{Type: skydb.TypeString},
			}, MakeUserRecordAssertion(skydb.AuthData{
				"username": "john.doe",
				"email":    "john.doe@example.com",
				"nickname": "iamyourfather",
			}))

			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"username": "john.doe",
					},
					"password": "secret",
					"profile": skydb.Data{
						"email":    "john.doe@example.com",
						"nickname": "iamyourfather",
					},
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Result, ShouldHaveSameTypeAs, AuthResponse{})
			authResp := resp.Result.(AuthResponse)
			So(authResp.Profile.ID, ShouldResemble, skydb.NewRecordID("user", authResp.UserID))
			So(authResp.Profile.DatabaseID, ShouldResemble, "_public")
			So(authResp.Profile.OwnerID, ShouldResemble, authResp.UserID)
			So(authResp.Profile.CreatorID, ShouldResemble, authResp.UserID)
			So(authResp.Profile.UpdaterID, ShouldResemble, authResp.UserID)
			So(authResp.Profile.CreatedAt, ShouldResemble, time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(authResp.Profile.UpdatedAt, ShouldResemble, time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(authResp.Profile.Data, ShouldResemble, skydb.Data{
				"username": "john.doe",
				"email":    "john.doe@example.com",
				"nickname": "iamyourfather",
			})
			So(authResp.AccessToken, ShouldNotBeEmpty)
			So(authResp.LastLoginAt, ShouldNotBeEmpty)
			So(authResp.LastSeenAt, ShouldNotBeEmpty)
			token := tokenStore.Token
			So(token.AuthInfoID, ShouldEqual, authResp.UserID)
			So(token.AccessToken, ShouldNotBeEmpty)

			authinfo := &skydb.AuthInfo{}
			err := conn.GetAuth(authResp.UserID, authinfo)
			So(err, ShouldBeNil)
			So(authinfo.Roles, ShouldBeNil)
		})

		Convey("sign up with invalid auth data", func() {
			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"iamyourfather": "john.doe",
					},
					"password": "secret",
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("sign up with duplicated keys in auth data and profile", func() {
			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"username": "john.doe",
					},
					"password": "secret",
					"profile": skydb.Data{
						"username": "iamyourfather",
					},
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("sign up new account with role base access control will have default role", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mock_skydb.NewMockTxDatabase(ctrl)
			db.EXPECT().
				Query(gomock.Any()).
				Do(MakeUsernameEmailQueryAssertion("john.doe", "john.doe@example.com")).
				Return(skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{})), nil).
				AnyTimes()
			txBegin := db.EXPECT().Begin().AnyTimes()
			db.EXPECT().Commit().After(txBegin)

			ExpectDBSaveUserWithAuthData(db, skydb.AuthData{"username": "john.doe", "email": "john.doe@example.com"})

			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"username": "john.doe",
						"email":    "john.doe@example.com",
					},
					"password": "secret",
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler := &SignupHandler{
				TokenStore:  &tokenStore,
				AccessModel: skydb.RoleBasedAccess,
			}
			handler.Handle(&req, &resp)
			authResp := resp.Result.(AuthResponse)

			authinfo := &skydb.AuthInfo{}
			err := conn.GetAuth(authResp.UserID, authinfo)
			So(err, ShouldBeNil)
			So(authinfo.Roles, ShouldResemble, []string{"user"})
		})

		Convey("sign up duplicate username", func() {
			authinfo := skydb.NewAuthInfo("secret")
			conn.CreateAuth(&authinfo)
			db.EXPECT().
				Query(gomock.Any()).
				Do(MakeUsernameEmailQueryAssertion("john.doe", "john.doe@example.com")).
				Return(skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{
					skydb.Record{
						ID: skydb.NewRecordID("user", authinfo.ID),
						Data: map[string]interface{}{
							"username": "john.doe",
							"email":    "john.doe@example.com",
						},
					},
				})), nil).
				AnyTimes()
			txBegin := db.EXPECT().Begin().AnyTimes()
			db.EXPECT().Rollback().After(txBegin)

			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"username": "john.doe",
						"email":    "john.doe@example.com",
					},
					"password": "secret",
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.Duplicated)
		})

		Convey("sign up duplicate email", func() {
			authinfo := skydb.NewAuthInfo("secret")
			conn.CreateAuth(&authinfo)
			db.EXPECT().
				Query(gomock.Any()).
				Do(MakeUsernameEmailQueryAssertion("john.doe", "john.doe@example.com")).
				Return(skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{
					skydb.Record{
						ID: skydb.NewRecordID("user", authinfo.ID),
						Data: map[string]interface{}{
							"username": "john.doe",
							"email":    "john.doe@example.com",
						},
					},
				})), nil).
				AnyTimes()
			txBegin := db.EXPECT().Begin().AnyTimes()
			db.EXPECT().Rollback().After(txBegin)

			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"username": "john.doe",
						"email":    "john.doe@example.com",
					},
					"password": "secret",
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.Duplicated)
		})
	})
}

func TestLoginHandler(t *testing.T) {
	Convey("LoginHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		conn := skydbtest.NewMapConn()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		db := mock_skydb.NewMockDatabase(ctrl)

		tokenStore := authtokentest.SingleTokenStore{}
		handler := &LoginHandler{
			TokenStore: &tokenStore,
		}

		Convey("login user", func() {
			authinfo := skydb.NewAuthInfo("secret")
			authinfo.Roles = []string{
				"Programmer",
				"Tester",
			}
			conn.CreateAuth(&authinfo)

			db.EXPECT().
				Query(gomock.Any()).
				Do(MakeUsernameEmailQueryAssertion("john.doe", "")).
				Return(skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{skydb.Record{
					ID:   skydb.NewRecordID("user", authinfo.ID),
					Data: map[string]interface{}{"username": "john.doe", "email": "john.doe@example.com"},
				}})), nil).
				AnyTimes()

			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"username": "john.doe",
					},
					"password": "secret",
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Result, ShouldHaveSameTypeAs, AuthResponse{})

			authResp := resp.Result.(AuthResponse)
			So(authResp.Profile.ID, ShouldResemble, skydb.NewRecordID("user", authResp.UserID))
			So(authResp.Profile.Data, ShouldResemble, skydb.Data{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			})
			So(authResp.AccessToken, ShouldNotBeEmpty)
			So(authResp.Roles, ShouldContain, "Programmer")
			So(authResp.Roles, ShouldContain, "Tester")

			token := tokenStore.Token
			So(token.AuthInfoID, ShouldEqual, authResp.UserID)
			So(token.AccessToken, ShouldNotBeEmpty)
		})

		Convey("login with invalid auth data", func() {
			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"iamyourfather": "john.doe",
					},
					"password": "secret",
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("login user wrong password", func() {
			authinfo := skydb.NewAuthInfo("secret")
			conn.CreateAuth(&authinfo)

			db.EXPECT().
				Query(gomock.Any()).
				Do(MakeUsernameEmailQueryAssertion("john.doe", "")).
				Return(skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{skydb.Record{
					ID:   skydb.NewRecordID("user", authinfo.ID),
					Data: map[string]interface{}{"username": "john.doe", "email": "john.doe@example.com"},
				}})), nil).
				AnyTimes()

			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"username": "john.doe",
					},
					"password": "wrongsecret",
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.InvalidCredentials)
		})

		Convey("login user not found", func() {
			db.EXPECT().
				Query(gomock.Any()).
				Do(MakeUsernameEmailQueryAssertion("john.doe", "")).
				Return(skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{})), nil).
				AnyTimes()

			req := router.Payload{
				Data: map[string]interface{}{
					"auth_data": skydb.AuthData{
						"username": "john.doe",
					},
					"password": "secret",
				},
				DBConn:   conn,
				Database: db,
			}
			resp := router.Response{}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.ResourceNotFound)
		})
	})
}

func TestLoginHandlerWithProvider(t *testing.T) {
	Convey("LoginHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		tokenStore := authtokentest.SingleTokenStore{}
		conn := singleUserConn{}
		db := skydbtest.NewMapDB()
		txdb := skydbtest.NewMockTxDatabase(db)
		providerRegistry := provider.NewRegistry()
		providerRegistry.RegisterAuthProvider("com.example", handlertest.NewSingleUserAuthProvider("com.example", "johndoe"))

		r := handlertest.NewSingleRouteRouter(&LoginHandler{
			TokenStore:       &tokenStore,
			ProviderRegistry: providerRegistry,
		}, func(p *router.Payload) {
			p.DBConn = &conn
			p.Database = txdb
		})

		Convey("login in non-existent provider", func() {
			resp := r.POST(`{"provider": "com.non-existent", "provider_auth_data": {"name": "johndoe"}}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 108,
		"name": "InvalidArgument",
		"info": {"arguments": ["provider"]},
		"message": "no auth provider of name \"com.non-existent\""
	}
}`)
			So(resp.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("login in existing", func() {
			authinfo := skydb.NewProviderInfoAuthInfo("com.example:johndoe", map[string]interface{}{"name": "boo"})
			n := timeNow()
			authinfo.LastLoginAt = &n
			authinfo.LastSeenAt = &n
			conn.authinfo = &authinfo
			defer func() {
				conn.authinfo = nil
			}()

			db.Save(&skydb.Record{
				ID:         skydb.NewRecordID("user", authinfo.ID),
				DatabaseID: db.ID(),
				OwnerID:    authinfo.ID,
				CreatorID:  authinfo.ID,
				UpdaterID:  authinfo.ID,
				CreatedAt:  n,
				UpdatedAt:  n,
				Data:       map[string]interface{}{},
			})

			resp := r.POST(`{"provider": "com.example", "provider_auth_data": {"name": "johndoe"}}`)

			token := tokenStore.Token
			So(token.AccessToken, ShouldNotBeBlank)
			So(conn.authinfo, ShouldNotBeNil)
			authData := conn.authinfo.ProviderInfo["com.example:johndoe"]
			authDataJSON, _ := json.Marshal(&authData)
			So(authDataJSON, ShouldEqualJSON, `{"name": "johndoe"}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"result": {
		"user_id": "%v",
		"profile": {
			"_type": "record",
			"_id": "user/%v",
			"_created_by": "%v",
			"_ownerID": "%v",
			"_updated_by": "%v",
			"_access": null,
			"_created_at": "2006-01-02T15:04:05Z",
			"_updated_at": "2006-01-02T15:04:05Z"
		},
		"access_token": "%v",
		"last_login_at": "%v",
		"last_seen_at": "%v"
	}
}`,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				token.AccessToken,
				n.Format(time.RFC3339Nano),
				n.Format(time.RFC3339Nano),
			))
			So(resp.Code, ShouldEqual, 200)
			// The LastLoginAt should updated
			So(conn.authinfo.LastLoginAt, ShouldNotEqual, n)
		})

		Convey("login in and create", func() {
			resp := r.POST(`{"provider": "com.example", "provider_auth_data": {"name": "johndoe"}}`)

			So(txdb.DidBegin, ShouldBeTrue)
			So(txdb.DidCommit, ShouldBeTrue)

			token := tokenStore.Token
			authinfo := conn.authinfo

			So(token.AccessToken, ShouldNotBeBlank)
			So(conn.authinfo, ShouldNotBeNil)
			authData := conn.authinfo.ProviderInfo["com.example:johndoe"]
			authDataJSON, _ := json.Marshal(&authData)
			So(authDataJSON, ShouldEqualJSON, `{"name": "johndoe"}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"result": {
		"user_id": "%v",
		"profile": {
			"_type": "record",
			"_id": "user/%v",
			"_created_by": "%v",
			"_ownerID": "%v",
			"_updated_by": "%v",
			"_access": null,
			"_created_at": "2006-01-02T15:04:05Z",
			"_updated_at": "2006-01-02T15:04:05Z"
		},
		"access_token": "%v"
	}
}`,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				token.AccessToken,
			))
			So(resp.Code, ShouldEqual, 200)
			So(authinfo.LastLoginAt, ShouldNotBeNil)

			_, ok := db.RecordMap[fmt.Sprintf("user/%s", authinfo.ID)]
			So(ok, ShouldBeTrue)
		})
	})
}

type singleUserConn struct {
	authinfo *skydb.AuthInfo
	skydb.Conn
}

func (conn *singleUserConn) UpdateAuth(authinfo *skydb.AuthInfo) error {
	if conn.authinfo != nil && conn.authinfo.ID == authinfo.ID {
		conn.authinfo = authinfo
		return nil
	}
	return skydb.ErrUserNotFound
}

func (conn *singleUserConn) CreateAuth(authinfo *skydb.AuthInfo) error {
	if conn.authinfo == nil {
		conn.authinfo = authinfo
		return nil
	}
	return skydb.ErrUserDuplicated
}

func (conn *singleUserConn) GetAuth(id string, authinfo *skydb.AuthInfo) error {
	if conn.authinfo != nil {
		*authinfo = *conn.authinfo
		return nil
	}
	return skydb.ErrUserNotFound
}

func (conn *singleUserConn) GetAuthByPrincipalID(principalID string, authinfo *skydb.AuthInfo) error {
	if conn.authinfo != nil {
		*authinfo = *conn.authinfo
		return nil
	}
	return skydb.ErrUserNotFound
}

func (conn *singleUserConn) GetRecordAccess(recordType string) (skydb.RecordACL, error) {
	return skydb.NewRecordACL([]skydb.RecordACLEntry{}), nil
}

func (conn *singleUserConn) GetRecordDefaultAccess(recordType string) (skydb.RecordACL, error) {
	return nil, nil
}

func (conn *singleUserConn) GetRecordFieldAccess() (skydb.FieldACL, error) {
	return skydb.FieldACL{}, nil
}

func TestSignupHandlerAsAnonymous(t *testing.T) {
	Convey("SignupHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		tokenStore := authtokentest.SingleTokenStore{}
		conn := singleUserConn{}
		db := skydbtest.NewMapDB()
		txdb := skydbtest.NewMockTxDatabase(db)

		r := handlertest.NewSingleRouteRouter(&SignupHandler{
			TokenStore: &tokenStore,
		}, func(p *router.Payload) {
			p.DBConn = &conn
			p.Database = txdb
		})

		Convey("signs up anonymously", func() {
			resp := r.POST(`{}`)

			So(txdb.DidBegin, ShouldBeTrue)
			So(txdb.DidCommit, ShouldBeTrue)

			token := tokenStore.Token
			authinfo := conn.authinfo

			So(token.AccessToken, ShouldNotBeBlank)
			So(conn.authinfo.ID, ShouldNotBeBlank)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"result": {
		"user_id": "%v",
		"profile": {
			"_type": "record",
			"_id": "user/%v",
			"_created_by": "%v",
			"_ownerID": "%v",
			"_updated_by": "%v",
			"_access": null,
			"_created_at": "2006-01-02T15:04:05Z",
			"_updated_at": "2006-01-02T15:04:05Z"
		},
		"access_token": "%v",
		"last_login_at": "%v",
		"last_seen_at": "%v"
	}
}`,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				token.AccessToken,
				authinfo.LastLoginAt.Format(time.RFC3339Nano),
				authinfo.LastSeenAt.Format(time.RFC3339Nano),
			))
			So(resp.Code, ShouldEqual, 200)

			user, ok := db.RecordMap[fmt.Sprintf("user/%s", authinfo.ID)]
			So(ok, ShouldBeTrue)
			So(len(user.Data) == 0, ShouldBeTrue)
		})

		Convey("errors when both usename and email is missing", func() {
			resp := r.POST(`{
				"password": "iamyourfather"
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 108,
		"name": "InvalidArgument",
		"info": {"arguments": ["auth_data"]},
		"message": "invalid auth data"
	}
}`)
			So(resp.Code, ShouldEqual, 400)
		})

		Convey("errors when password is missing", func() {
			resp := r.POST(`{
				"auth_data": {
					"username": "john.doe",
					"email": "john.doe@example.com"
				}
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 108,
		"name": "InvalidArgument",
		"info": {"arguments": ["password"]},
		"message": "empty password"
	}
}`)
			So(resp.Code, ShouldEqual, 400)
		})
	})
}

func TestSignupHandlerWithProvider(t *testing.T) {
	Convey("SignupHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		tokenStore := authtokentest.SingleTokenStore{}
		conn := singleUserConn{}
		db := skydbtest.NewMapDB()
		txdb := skydbtest.NewMockTxDatabase(db)
		providerRegistry := provider.NewRegistry()
		providerRegistry.RegisterAuthProvider("com.example", handlertest.NewSingleUserAuthProvider("com.example", "johndoe"))

		r := handlertest.NewSingleRouteRouter(&SignupHandler{
			TokenStore:       &tokenStore,
			ProviderRegistry: providerRegistry,
		}, func(p *router.Payload) {
			p.DBConn = &conn
			p.Database = txdb
		})

		Convey("signs up with non-existent provider", func() {
			resp := r.POST(`{"provider": "com.non-existent", "provider_auth_data": {"name": "johndoe"}}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 108,
		"name": "InvalidArgument",
		"info": {"arguments": ["provider"]},
		"message": "no auth provider of name \"com.non-existent\""
	}
}`)
			So(resp.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("signs up with user", func() {
			resp := r.POST(`{"provider": "com.example", "provider_auth_data": {"name": "johndoe"}}`)

			So(txdb.DidBegin, ShouldBeTrue)
			So(txdb.DidCommit, ShouldBeTrue)

			token := tokenStore.Token
			authinfo := conn.authinfo

			So(token.AccessToken, ShouldNotBeBlank)
			authData := conn.authinfo.ProviderInfo["com.example:johndoe"]
			authDataJSON, _ := json.Marshal(&authData)
			So(authDataJSON, ShouldEqualJSON, `{"name": "johndoe"}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"result": {
		"user_id": "%v",
		"profile": {
			"_type": "record",
			"_id": "user/%v",
			"_created_by": "%v",
			"_ownerID": "%v",
			"_updated_by": "%v",
			"_access": null,
			"_created_at": "2006-01-02T15:04:05Z",
			"_updated_at": "2006-01-02T15:04:05Z"
		},
		"access_token": "%v",
		"last_login_at": "%v",
		"last_seen_at": "%v"
	}
}`,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				authinfo.ID,
				token.AccessToken,
				authinfo.LastLoginAt.Format(time.RFC3339Nano),
				authinfo.LastSeenAt.Format(time.RFC3339Nano),
			))
			So(resp.Code, ShouldEqual, 200)

			_, ok := db.RecordMap[fmt.Sprintf("user/%s", authinfo.ID)]
			So(ok, ShouldBeTrue)
		})

		Convey("signs up with incorrect user", func() {
			resp := r.POST(`{"provider": "com.example", "provider_auth_data": {"name": "janedoe"}}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"error": {
		"code": 105,
		"name": "InvalidCredentials",
		"message": "unable to login with the given credentials"
	}
}`))
			So(resp.Code, ShouldEqual, 401)
		})
	})
}

type deleteTokenStore struct {
	deletedAccessToken string
	errToReturn        error
}

func (store *deleteTokenStore) NewToken(appName string, authInfoID string) (authtoken.Token, error) {
	return authtoken.New(appName, authInfoID, time.Time{}), nil
}

func (store *deleteTokenStore) Get(accessToken string, token *authtoken.Token) error {
	panic("Thou shalt not call Get")
}

func (store *deleteTokenStore) Put(token *authtoken.Token) error {
	panic("Thou shalt not call Put")
}

func (store *deleteTokenStore) Delete(accessToken string) error {
	store.deletedAccessToken = accessToken
	return store.errToReturn
}

func TestLogoutHandler(t *testing.T) {
	Convey("LogoutHandler", t, func() {
		tokenStore := &deleteTokenStore{}
		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()

		r := handlertest.NewSingleRouteRouter(&LogoutHandler{
			TokenStore: tokenStore,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = db
		})

		Convey("deletes existing access token", func() {
			resp := r.POST(`{
	"access_token": "someaccesstoken"
}`)
			So(tokenStore.deletedAccessToken, ShouldEqual, "someaccesstoken")
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result":{"status":"OK"}
}`)
			So(resp.Code, ShouldEqual, 200)
		})

		Convey("deletes non-existing access token without error", func() {
			tokenStore.errToReturn = &authtoken.NotFoundError{}
			resp := r.POST(`{
	"access_token": "notexistaccesstoken"
}`)
			So(tokenStore.deletedAccessToken, ShouldEqual, "notexistaccesstoken")
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result":{"status":"OK"}
}`)
			So(resp.Code, ShouldEqual, 200)
		})

		Convey("fails to delete due to unknown error", func() {
			tokenStore.errToReturn = errors.New("some interesting error")
			resp := r.POST(`{
	"access_token": "someaccesstoken"
}`)
			So(tokenStore.deletedAccessToken, ShouldEqual, "someaccesstoken")
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 10000,
		"name": "UnexpectedError",
		"message": "some interesting error"
	}
}`)
			So(resp.Code, ShouldEqual, 500)
		})
	})
}

func TestPasswordHandlerWithProvider(t *testing.T) {
	Convey("PasswordHandler", t, func() {
		conn := singleUserConn{}
		authinfo := skydb.NewAuthInfo("chima")
		authinfo.ID = "user-uuid"
		conn.CreateAuth(&authinfo)
		tokenStore := authtokentest.SingleTokenStore{}
		token := authtoken.New("_", authinfo.ID, time.Time{})
		tokenStore.Put(&token)

		r := handlertest.NewSingleRouteRouter(&PasswordHandler{
			TokenStore: &tokenStore,
		}, func(p *router.Payload) {
			p.DBConn = &conn
			p.Database = skydbtest.NewMapDB()
		})

		Convey("change password success", func() {
			resp := r.POST(fmt.Sprintf(`{
	"access_token": "%s",
	"username": "lord-of-skygear",
	"old_password": "chima",
	"password": "faseng"
}`, token.AccessToken))

			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"result": {
		"user_id": "user-uuid",
		"access_token": "%s"
	}
}`, tokenStore.Token.AccessToken))
			So(resp.Code, ShouldEqual, 200)
		})

	})
}
