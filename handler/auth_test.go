package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/oursky/skygear/authtoken"
	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/plugin/provider"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skydb/skydbtest"
	"github.com/oursky/skygear/skyerr"
	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func tempDir() string {
	dir, err := ioutil.TempDir("", "skygear.skydb.handler.auth.test")
	if err != nil {
		panic(err)
	}
	return dir
}

// singleTokenStore implementassigns to and returns itself.
type singleTokenStore authtoken.Token

func (s *singleTokenStore) Get(accessToken string, token *authtoken.Token) error {
	*token = authtoken.Token(*s)
	return nil
}

func (s *singleTokenStore) Put(token *authtoken.Token) error {
	*s = singleTokenStore(*token)
	return nil
}

func (s *singleTokenStore) Delete(accessToken string) error {
	panic("Thou shalt not call Delete")
}

// Seems like a memory imlementation of skydb will make tests
// faster and easier

func TestHomeHandler(t *testing.T) {
	req := router.Payload{}
	resp := router.Response{}

	handler := &HomeHandler{}
	handler.Handle(&req, &resp)
	var s statusResponse

	switch pt := resp.Result.(type) {
	default:
		t.Fatalf("unexpected type %T", pt)
	case statusResponse:
		s = resp.Result.(statusResponse)
	}

	if s.Status != "OK" {
		t.Fatalf("got response %v, want `OK`", s.Status)
	}
}

func TestSignupHandler(t *testing.T) {
	conn := skydbtest.NewMapConn()

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"username": "john.doe",
			"email":    "john.doe@example.com",
			"password": "secret",
		},
		DBConn: conn,
	}
	resp := router.Response{}
	handler := &SignupHandler{
		TokenStore: &tokenStore,
	}
	handler.Handle(&req, &resp)

	authResp, ok := resp.Result.(authResponse)
	if !ok {
		t.Fatalf("got type = %T, want type authResponse", resp.Result)
	}

	if authResp.Username != "john.doe" {
		t.Fatalf("got authResp.Username = %v, want john.doe", authResp.Username)
	}

	if authResp.Email != "john.doe@example.com" {
		t.Fatalf("got authResp.Email = %v, want john.doe@example.com", authResp.Email)
	}

	if authResp.AccessToken == "" {
		t.Fatal("got authResp.AccessToken, want non-empty value")
	}

	token := authtoken.Token(tokenStore)
	if token.UserInfoID != authResp.UserID {
		t.Fatalf("Token userID don't match with response %v, %v", token.UserInfoID, authResp.UserID)
	}

	if token.AccessToken == "" {
		t.Fatal("got empty token.AccessToken, want non-empty value")
	}
}

func TestSignupHandlerDuplicatedUsername(t *testing.T) {
	conn := skydbtest.NewMapConn()

	userinfo := skydb.NewUserInfo("john.doe", "", "secret")
	conn.CreateUser(&userinfo)

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"username": "john.doe",
			"email":    "john.doe@example.com",
			"password": "secret",
		},
		DBConn: conn,
	}
	resp := router.Response{}
	handler := &SignupHandler{
		TokenStore: &tokenStore,
	}
	handler.Handle(&req, &resp)

	errorResponse, ok := resp.Err.(skyerr.Error)
	if !ok {
		t.Fatalf("got type = %T, want type skyerr.Error", resp.Err)
	}

	if errorResponse.Code() != 109 {
		t.Fatalf("got errorResponse.Code() = %v, want 109", errorResponse.Code())
	}
}

func TestSignupHandlerDuplicatedEmail(t *testing.T) {
	conn := skydbtest.NewMapConn()

	userinfo := skydb.NewUserInfo("", "john.doe@example.com", "secret")
	conn.CreateUser(&userinfo)

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"username": "john.doe",
			"email":    "john.doe@example.com",
			"password": "secret",
		},
		DBConn: conn,
	}
	resp := router.Response{}
	handler := &SignupHandler{
		TokenStore: &tokenStore,
	}
	handler.Handle(&req, &resp)

	errorResponse, ok := resp.Err.(skyerr.Error)
	if !ok {
		t.Fatalf("got type = %T, want type skyerr.Error", resp.Err)
	}

	if errorResponse.Code() != 109 {
		t.Fatalf("got errorResponse.Code() = %v, want 109", errorResponse.Code())
	}
}

func TestLoginHandler(t *testing.T) {
	conn := skydbtest.NewMapConn()

	userinfo := skydb.NewUserInfo("john.doe", "john.doe@example.com", "secret")
	conn.CreateUser(&userinfo)

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"username": "john.doe",
			"password": "secret",
		},
		DBConn: conn,
	}
	resp := router.Response{}
	handler := &LoginHandler{
		TokenStore: &tokenStore,
	}
	handler.Handle(&req, &resp)

	authResp, ok := resp.Result.(authResponse)
	if !ok {
		t.Fatalf("got type = %T, want type authResponse", resp.Result)
	}

	if authResp.Username != "john.doe" {
		t.Fatalf("got authResp.UserID = %v, want userinfoid", authResp.Username)
	}

	if authResp.Email != "john.doe@example.com" {
		t.Fatalf("got authResp.Email = %v, want john.doe@example.com", authResp.Email)
	}

	if authResp.AccessToken == "" {
		t.Fatal("got authResp.AccessToken, want non-empty value")
	}

	token := authtoken.Token(tokenStore)
	if token.UserInfoID != authResp.UserID {
		t.Fatalf("Token userID don't match with response %v, %v", token.UserInfoID, authResp.UserID)
	}

	if token.AccessToken == "" {
		t.Fatal("got empty token.AccessToken, want non-empty value")
	}
}

func TestLoginHandlerWrongPassword(t *testing.T) {
	conn := skydbtest.NewMapConn()

	userinfo := skydb.NewUserInfo("john.doe", "john.doe@example.com", "secret")
	conn.CreateUser(&userinfo)

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"username": "john.doe",
			"password": "wrongsecret",
		},
		DBConn: conn,
	}
	resp := router.Response{}
	handler := &LoginHandler{
		TokenStore: &tokenStore,
	}
	handler.Handle(&req, &resp)

	errorResponse, ok := resp.Err.(skyerr.Error)
	if !ok {
		t.Fatalf("got type = %T, want type skyerr.Error", resp.Err)
	}

	if errorResponse.Code() != 105 {
		t.Fatalf("got resp.Err.Code() = %v, want 105", errorResponse.Code())
	}
}

func TestLoginHandlerNotFound(t *testing.T) {
	conn := skydbtest.NewMapConn()

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"username": "john.doe",
			"password": "secret",
		},
		DBConn: conn,
	}
	resp := router.Response{}
	handler := &LoginHandler{
		TokenStore: &tokenStore,
	}
	handler.Handle(&req, &resp)

	errorResponse, ok := resp.Err.(skyerr.Error)
	if !ok {
		t.Fatalf("got type = %T, want type skyerr.Error", resp.Err)
	}

	if errorResponse.Code() != skyerr.ResourceNotFound {
		t.Fatalf("got resp.Err.Code() = %v, want %v", errorResponse.Code(), skyerr.ResourceNotFound)
	}
}

func TestLoginHandlerWithProvider(t *testing.T) {
	Convey("LoginHandler", t, func() {
		tokenStore := singleTokenStore{}
		conn := singleUserConn{}
		providerRegistry := provider.NewRegistry()
		providerRegistry.RegisterAuthProvider("com.example", handlertest.NewSingleUserAuthProvider("com.example", "johndoe"))

		r := handlertest.NewSingleRouteRouter(&LoginHandler{
			TokenStore:       &tokenStore,
			ProviderRegistry: providerRegistry,
		}, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("login in existing", func() {
			userinfo := skydb.NewProvidedAuthUserInfo("com.example:johndoe", map[string]interface{}{"name": "boo"})
			conn.userinfo = &userinfo
			defer func() {
				conn.userinfo = nil
			}()

			resp := r.POST(`{"provider": "com.example", "auth_data": {"name": "johndoe"}}`)

			token := authtoken.Token(tokenStore)
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

			token := authtoken.Token(tokenStore)
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

func TestSignupHandlerAsAnonymous(t *testing.T) {
	Convey("SignupHandler", t, func() {
		tokenStore := singleTokenStore{}
		conn := singleUserConn{}

		r := handlertest.NewSingleRouteRouter(&SignupHandler{
			TokenStore: &tokenStore,
		}, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("signs up anonymously", func() {
			resp := r.POST(`{}`)

			token := authtoken.Token(tokenStore)
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
		})

		Convey("errors when both usename and email is missing", func() {
			resp := r.POST(`{
				"password": "iamyourfather"
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 108,
		"name": "InvalidArgument",
		"message": "empty identifier(username, email) or password"
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
		"message": "empty identifier(username, email) or password"
	}
}`)
			So(resp.Code, ShouldEqual, 400)
		})
	})
}

func TestSignupHandlerWithProvider(t *testing.T) {
	Convey("SignupHandler", t, func() {
		tokenStore := singleTokenStore{}
		conn := singleUserConn{}
		providerRegistry := provider.NewRegistry()
		providerRegistry.RegisterAuthProvider("com.example", handlertest.NewSingleUserAuthProvider("com.example", "johndoe"))

		r := handlertest.NewSingleRouteRouter(&SignupHandler{
			TokenStore:       &tokenStore,
			ProviderRegistry: providerRegistry,
		}, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("signs up with user", func() {
			resp := r.POST(`{"provider": "com.example", "auth_data": {"name": "johndoe"}}`)

			token := authtoken.Token(tokenStore)
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

		r := handlertest.NewSingleRouteRouter(&LogoutHandler{
			TokenStore: tokenStore,
		}, func(p *router.Payload) {
			p.DBConn = conn
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
		tokenStore := singleTokenStore{}
		token := authtoken.New("_", userinfo.ID, time.Time{})
		tokenStore.Put(&token)

		r := handlertest.NewSingleRouteRouter(&PasswordHandler{}, func(p *router.Payload) {
			p.DBConn = &conn
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
