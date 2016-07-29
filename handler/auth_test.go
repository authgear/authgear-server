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

	"github.com/skygeario/skygear-server/authtoken"
	"github.com/skygeario/skygear-server/authtoken/authtokentest"
	"github.com/skygeario/skygear-server/handler/handlertest"
	"github.com/skygeario/skygear-server/plugin/provider"
	"github.com/skygeario/skygear-server/router"
	"github.com/skygeario/skygear-server/skydb"
	"github.com/skygeario/skygear-server/skydb/skydbtest"
	"github.com/skygeario/skygear-server/skyerr"
	. "github.com/skygeario/skygear-server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func tempDir() string {
	dir, err := ioutil.TempDir("", "skygear.skydb.handler.auth.test")
	if err != nil {
		panic(err)
	}
	return dir
}

// Seems like a memory imlementation of skydb will make tests
// faster and easier

func TestSignupHandler(t *testing.T) {
	Convey("SignupHandler", t, func() {
		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()
		txdb := skydbtest.NewMockTxDatabase(db)
		tokenStore := authtokentest.SingleTokenStore{}

		Convey("sign up new account", func() {
			req := router.Payload{
				Data: map[string]interface{}{
					"username": "john.doe",
					"email":    "john.doe@example.com",
					"password": "secret",
				},
				DBConn:   conn,
				Database: txdb,
			}
			resp := router.Response{}
			handler := &SignupHandler{
				TokenStore: &tokenStore,
			}
			handler.Handle(&req, &resp)

			So(txdb.DidBegin, ShouldBeTrue)
			So(txdb.DidCommit, ShouldBeTrue)

			So(resp.Result, ShouldHaveSameTypeAs, authResponse{})
			authResp := resp.Result.(authResponse)
			So(authResp.Username, ShouldEqual, "john.doe")
			So(authResp.Email, ShouldEqual, "john.doe@example.com")
			So(authResp.AccessToken, ShouldNotBeEmpty)
			token := tokenStore.Token
			So(token.UserInfoID, ShouldEqual, authResp.UserID)
			So(token.AccessToken, ShouldNotBeEmpty)

			userinfo := &skydb.UserInfo{}
			err := conn.GetUserByUsernameEmail("john.doe", "", userinfo)
			So(err, ShouldBeNil)
			So(userinfo.Roles, ShouldBeNil)

			_, ok := db.RecordMap[fmt.Sprintf("user/%s", token.UserInfoID)]
			So(ok, ShouldBeTrue)
		})

		Convey("sign up new account with role base access control will have defautl role", func() {
			req := router.Payload{
				Data: map[string]interface{}{
					"username": "john.doe",
					"email":    "john.doe@example.com",
					"password": "secret",
				},
				DBConn:   conn,
				Database: txdb,
			}
			resp := router.Response{}
			handler := &SignupHandler{
				TokenStore:  &tokenStore,
				AccessModel: skydb.RoleBasedAccess,
			}
			handler.Handle(&req, &resp)

			userinfo := &skydb.UserInfo{}
			err := conn.GetUserByUsernameEmail("john.doe", "", userinfo)
			So(err, ShouldBeNil)
			So(userinfo.Roles, ShouldResemble, []string{"user"})
		})

		Convey("sign up duplicate username", func() {
			userinfo := skydb.NewUserInfo("john.doe", "", "secret")
			conn.CreateUser(&userinfo)

			req := router.Payload{
				Data: map[string]interface{}{
					"username": "john.doe",
					"email":    "john.doe@example.com",
					"password": "secret",
				},
				DBConn:   conn,
				Database: txdb,
			}
			resp := router.Response{}
			handler := &SignupHandler{
				TokenStore: &tokenStore,
			}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.Duplicated)
		})

		Convey("sign up duplicate email", func() {
			userinfo := skydb.NewUserInfo("", "john.doe@example.com", "secret")
			conn.CreateUser(&userinfo)

			req := router.Payload{
				Data: map[string]interface{}{
					"username": "john.doe",
					"email":    "john.doe@example.com",
					"password": "secret",
				},
				DBConn:   conn,
				Database: txdb,
			}
			resp := router.Response{}
			handler := &SignupHandler{
				TokenStore: &tokenStore,
			}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.Duplicated)
		})
	})
}

func TestLoginHandler(t *testing.T) {
	Convey("LoginHandler", t, func() {
		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()
		txdb := skydbtest.NewMockTxDatabase(db)
		tokenStore := authtokentest.SingleTokenStore{}

		Convey("login user", func() {
			userinfo := skydb.NewUserInfo("john.doe", "john.doe@example.com", "secret")
			conn.CreateUser(&userinfo)

			req := router.Payload{
				Data: map[string]interface{}{
					"username": "john.doe",
					"password": "secret",
				},
				DBConn:   conn,
				Database: txdb,
			}
			resp := router.Response{}
			handler := &LoginHandler{
				TokenStore: &tokenStore,
			}
			handler.Handle(&req, &resp)

			So(resp.Result, ShouldHaveSameTypeAs, authResponse{})
			authResp := resp.Result.(authResponse)
			So(authResp.Username, ShouldEqual, "john.doe")
			So(authResp.Email, ShouldEqual, "john.doe@example.com")
			So(authResp.AccessToken, ShouldNotBeEmpty)
			token := tokenStore.Token
			So(token.UserInfoID, ShouldEqual, authResp.UserID)
			So(token.AccessToken, ShouldNotBeEmpty)
		})

		Convey("login user with username in different case should ok", func() {
			userinfo := skydb.NewUserInfo("john.doe", "john.doe@example.com", "secret")
			conn.CreateUser(&userinfo)

			req := router.Payload{
				Data: map[string]interface{}{
					"username": "john.DOE",
					"password": "secret",
				},
				DBConn:   conn,
				Database: txdb,
			}
			resp := router.Response{}
			handler := &LoginHandler{
				TokenStore: &tokenStore,
			}
			handler.Handle(&req, &resp)

			So(resp.Result, ShouldHaveSameTypeAs, authResponse{})
			authResp := resp.Result.(authResponse)
			So(authResp.Username, ShouldEqual, "john.doe")
			So(authResp.Email, ShouldEqual, "john.doe@example.com")
			So(authResp.AccessToken, ShouldNotBeEmpty)
			token := tokenStore.Token
			So(token.UserInfoID, ShouldEqual, authResp.UserID)
			So(token.AccessToken, ShouldNotBeEmpty)
		})

		Convey("login user with email in different case should ok", func() {
			userinfo := skydb.NewUserInfo("john.doe", "john.doe@example.com", "secret")
			conn.CreateUser(&userinfo)

			req := router.Payload{
				Data: map[string]interface{}{
					"email":    "john.DOE@example.com",
					"password": "secret",
				},
				DBConn:   conn,
				Database: txdb,
			}
			resp := router.Response{}
			handler := &LoginHandler{
				TokenStore: &tokenStore,
			}
			handler.Handle(&req, &resp)

			So(resp.Result, ShouldHaveSameTypeAs, authResponse{})
			authResp := resp.Result.(authResponse)
			So(authResp.Username, ShouldEqual, "john.doe")
			So(authResp.Email, ShouldEqual, "john.doe@example.com")
			So(authResp.AccessToken, ShouldNotBeEmpty)
			token := tokenStore.Token
			So(token.UserInfoID, ShouldEqual, authResp.UserID)
			So(token.AccessToken, ShouldNotBeEmpty)
		})
		Convey("login user wrong password", func() {
			userinfo := skydb.NewUserInfo("john.doe", "john.doe@example.com", "secret")
			conn.CreateUser(&userinfo)

			req := router.Payload{
				Data: map[string]interface{}{
					"username": "john.doe",
					"password": "wrongsecret",
				},
				DBConn:   conn,
				Database: txdb,
			}
			resp := router.Response{}
			handler := &LoginHandler{
				TokenStore: &tokenStore,
			}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.InvalidCredentials)
		})

		Convey("login user not found", func() {
			req := router.Payload{
				Data: map[string]interface{}{
					"username": "john.doe",
					"password": "secret",
				},
				DBConn:   conn,
				Database: txdb,
			}
			resp := router.Response{}
			handler := &LoginHandler{
				TokenStore: &tokenStore,
			}
			handler.Handle(&req, &resp)

			So(resp.Err, ShouldImplement, (*skyerr.Error)(nil))
			errorResponse := resp.Err.(skyerr.Error)
			So(errorResponse.Code(), ShouldEqual, skyerr.ResourceNotFound)
		})
	})
}

func TestLoginHandlerWithProvider(t *testing.T) {
	Convey("LoginHandler", t, func() {
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
			resp := r.POST(`{"provider": "com.non-existent", "auth_data": {"name": "johndoe"}}`)
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
			userinfo := skydb.NewProvidedAuthUserInfo("com.example:johndoe", map[string]interface{}{"name": "boo"})
			conn.userinfo = &userinfo
			defer func() {
				conn.userinfo = nil
			}()

			resp := r.POST(`{"provider": "com.example", "auth_data": {"name": "johndoe"}}`)

			token := tokenStore.Token
			So(token.AccessToken, ShouldNotBeBlank)
			So(conn.userinfo, ShouldNotBeNil)
			authData := conn.userinfo.Auth["com.example:johndoe"]
			authDataJSON, _ := json.Marshal(&authData)
			So(authDataJSON, ShouldEqualJSON, `{"name": "johndoe"}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"result": {
		"user_id": "%v",
		"access_token": "%v"
	}
}`, userinfo.ID, token.AccessToken))
			So(resp.Code, ShouldEqual, 200)
		})

		Convey("login in and create", func() {
			resp := r.POST(`{"provider": "com.example", "auth_data": {"name": "johndoe"}}`)

			So(txdb.DidBegin, ShouldBeTrue)
			So(txdb.DidCommit, ShouldBeTrue)

			token := tokenStore.Token
			userinfo := conn.userinfo

			So(token.AccessToken, ShouldNotBeBlank)
			So(conn.userinfo, ShouldNotBeNil)
			authData := conn.userinfo.Auth["com.example:johndoe"]
			authDataJSON, _ := json.Marshal(&authData)
			So(authDataJSON, ShouldEqualJSON, `{"name": "johndoe"}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"result": {
		"user_id": "%v",
		"access_token": "%v"
	}
}`, userinfo.ID, token.AccessToken))
			So(resp.Code, ShouldEqual, 200)

			_, ok := db.RecordMap[fmt.Sprintf("user/%s", userinfo.ID)]
			So(ok, ShouldBeTrue)
		})
	})
}

type singleUserConn struct {
	userinfo *skydb.UserInfo
	skydb.Conn
}

func (conn *singleUserConn) UpdateUser(userinfo *skydb.UserInfo) error {
	if conn.userinfo != nil && conn.userinfo.ID == userinfo.ID {
		conn.userinfo = userinfo
		return nil
	}
	return skydb.ErrUserNotFound
}

func (conn *singleUserConn) CreateUser(userinfo *skydb.UserInfo) error {
	if conn.userinfo == nil {
		conn.userinfo = userinfo
		return nil
	}
	return skydb.ErrUserDuplicated
}

func (conn *singleUserConn) GetUser(id string, userinfo *skydb.UserInfo) error {
	if conn.userinfo != nil {
		*userinfo = *conn.userinfo
		return nil
	}
	return skydb.ErrUserNotFound
}

func (conn *singleUserConn) GetUserByPrincipalID(principalID string, userinfo *skydb.UserInfo) error {
	if conn.userinfo != nil {
		*userinfo = *conn.userinfo
		return nil
	}
	return skydb.ErrUserNotFound
}

func (conn *singleUserConn) GetRecordAccess(recordType string) (skydb.RecordACL, error) {
	return skydb.NewRecordACL([]skydb.RecordACLEntry{}), nil
}

func TestSignupHandlerAsAnonymous(t *testing.T) {
	Convey("SignupHandler", t, func() {
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
			userinfo := conn.userinfo

			So(token.AccessToken, ShouldNotBeBlank)
			So(conn.userinfo.ID, ShouldNotBeBlank)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"result": {
		"user_id": "%v",
		"access_token": "%v"
	}
}`, userinfo.ID, token.AccessToken))
			So(resp.Code, ShouldEqual, 200)

			_, ok := db.RecordMap[fmt.Sprintf("user/%s", userinfo.ID)]
			So(ok, ShouldBeTrue)
		})

		Convey("errors when both usename and email is missing", func() {
			resp := r.POST(`{
				"password": "iamyourfather"
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 108,
		"name": "InvalidArgument",
		"info": {"arguments": ["username","email"]},
		"message": "empty username and empty email"
	}
}`)
			So(resp.Code, ShouldEqual, 400)
		})

		Convey("errors when password is missing", func() {
			resp := r.POST(`{
				"username": "john.doe",
				"email": "john.doe@example.com"
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
			resp := r.POST(`{"provider": "com.non-existent", "auth_data": {"name": "johndoe"}}`)
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
			resp := r.POST(`{"provider": "com.example", "auth_data": {"name": "johndoe"}}`)

			So(txdb.DidBegin, ShouldBeTrue)
			So(txdb.DidCommit, ShouldBeTrue)

			token := tokenStore.Token
			userinfo := conn.userinfo

			So(token.AccessToken, ShouldNotBeBlank)
			authData := conn.userinfo.Auth["com.example:johndoe"]
			authDataJSON, _ := json.Marshal(&authData)
			So(authDataJSON, ShouldEqualJSON, `{"name": "johndoe"}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"result": {
		"user_id": "%v",
		"access_token": "%v"
	}
}`, userinfo.ID, token.AccessToken))
			So(resp.Code, ShouldEqual, 200)

			_, ok := db.RecordMap[fmt.Sprintf("user/%s", userinfo.ID)]
			So(ok, ShouldBeTrue)
		})

		Convey("signs up with incorrect user", func() {
			resp := r.POST(`{"provider": "com.example", "auth_data": {"name": "janedoe"}}`)

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

func (store *deleteTokenStore) NewToken(appName string, userInfoID string) (authtoken.Token, error) {
	return authtoken.New(appName, userInfoID, time.Time{}), nil
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
		userinfo := skydb.NewUserInfo("lord-of-skygear", "limouren@skygear.io", "chima")
		userinfo.ID = "user-uuid"
		conn.CreateUser(&userinfo)
		tokenStore := authtokentest.SingleTokenStore{}
		token := authtoken.New("_", userinfo.ID, time.Time{})
		tokenStore.Put(&token)

		r := handlertest.NewSingleRouteRouter(&PasswordHandler{}, func(p *router.Payload) {
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
}`, token.AccessToken))
			So(resp.Code, ShouldEqual, 200)
		})

	})
}
