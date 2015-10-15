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
	"github.com/oursky/skygear/oddb"
	"github.com/oursky/skygear/oddb/oddbtest"
	. "github.com/oursky/skygear/ourtest"
	"github.com/oursky/skygear/provider"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

func tempDir() string {
	dir, err := ioutil.TempDir("", "skygear.oddb.handler.auth.test")
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

// Seems like a memory imlementation of oddb will make tests
// faster and easier

func TestHomeHandler(t *testing.T) {
	req := router.Payload{}
	resp := router.Response{}

	HomeHandler(&req, &resp)
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
	conn := oddbtest.NewMapConn()

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"user_id":  "userinfoid",
			"email":    "john.doe@example.com",
			"password": "secret",
		},
		DBConn:     conn,
		TokenStore: &tokenStore,
	}
	resp := router.Response{}
	SignupHandler(&req, &resp)

	authResp, ok := resp.Result.(authResponse)
	if !ok {
		t.Fatalf("got type = %T, want type authResponse", resp.Result)
	}

	if authResp.UserID != "userinfoid" {
		t.Fatalf("got authResp.UserID = %v, want userinfoid", authResp.UserID)
	}

	if authResp.Email != "john.doe@example.com" {
		t.Fatalf("got authResp.Email = %v, want john.doe@example.com", authResp.Email)
	}

	if authResp.AccessToken == "" {
		t.Fatal("got authResp.AccessToken, want non-empty value")
	}

	token := authtoken.Token(tokenStore)
	if token.UserInfoID != "userinfoid" {
		t.Fatalf("got token.UserInfoID = %v, want userinfoid", token.UserInfoID)
	}

	if token.AccessToken == "" {
		t.Fatal("got empty token.AccessToken, want non-empty value")
	}
}

func TestSignupHandlerDuplicated(t *testing.T) {
	conn := oddbtest.NewMapConn()

	userinfo := oddb.NewUserInfo("userinfoid", "john.doe@example.com", "secret")
	conn.CreateUser(&userinfo)

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"user_id":  "userinfoid",
			"email":    "john.doe@example.com",
			"password": "secret",
		},
		DBConn:     conn,
		TokenStore: &tokenStore,
	}
	resp := router.Response{}
	SignupHandler(&req, &resp)

	errorResponse, ok := resp.Err.(skyerr.Error)
	if !ok {
		t.Fatalf("got type = %T, want type skyerr.Error", resp.Err)
	}

	if errorResponse.Code() != 101 {
		t.Fatalf("got errorResponse.Code() = %v, want 101", errorResponse.Code())
	}
}

func TestLoginHandler(t *testing.T) {
	conn := oddbtest.NewMapConn()

	userinfo := oddb.NewUserInfo("userinfoid", "john.doe@example.com", "secret")
	conn.CreateUser(&userinfo)

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"user_id":  "userinfoid",
			"password": "secret",
		},
		DBConn:     conn,
		TokenStore: &tokenStore,
	}
	resp := router.Response{}
	LoginHandler(&req, &resp)

	authResp, ok := resp.Result.(authResponse)
	if !ok {
		t.Fatalf("got type = %T, want type authResponse", resp.Result)
	}

	if authResp.UserID != "userinfoid" {
		t.Fatalf("got authResp.UserID = %v, want userinfoid", authResp.UserID)
	}

	if authResp.Email != "john.doe@example.com" {
		t.Fatalf("got authResp.Email = %v, want john.doe@example.com", authResp.Email)
	}

	if authResp.AccessToken == "" {
		t.Fatal("got authResp.AccessToken, want non-empty value")
	}

	token := authtoken.Token(tokenStore)
	if token.UserInfoID != "userinfoid" {
		t.Fatalf("got token.UserInfoID = %v, want userinfoid", token.UserInfoID)
	}

	if token.AccessToken == "" {
		t.Fatal("got empty token.AccessToken, want non-empty value")
	}
}

func TestLoginHandlerWrongPassword(t *testing.T) {
	conn := oddbtest.NewMapConn()

	userinfo := oddb.NewUserInfo("userinfoid", "john.doe@example.com", "secret")
	conn.CreateUser(&userinfo)

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"user_id":  "userinfoid",
			"password": "wrongsecret",
		},
		DBConn:     conn,
		TokenStore: &tokenStore,
	}
	resp := router.Response{}
	LoginHandler(&req, &resp)

	errorResponse, ok := resp.Err.(skyerr.Error)
	if !ok {
		t.Fatalf("got type = %T, want type skyerr.Error", resp.Err)
	}

	if errorResponse != skyerr.ErrInvalidLogin {
		t.Fatalf("got resp.Err = %v, want ErrInvalidLogin", errorResponse)
	}
}

func TestLoginHandlerNotFound(t *testing.T) {
	conn := oddbtest.NewMapConn()

	tokenStore := singleTokenStore{}
	req := router.Payload{
		Data: map[string]interface{}{
			"user_id":  "userinfoid",
			"password": "secret",
		},
		DBConn:     conn,
		TokenStore: &tokenStore,
	}
	resp := router.Response{}
	LoginHandler(&req, &resp)

	errorResponse, ok := resp.Err.(skyerr.Error)
	if !ok {
		t.Fatalf("got type = %T, want type skyerr.Error", resp.Err)
	}

	if errorResponse != skyerr.ErrUserNotFound {
		t.Fatalf("got resp.Err = %v, want ErrUserNotFound", errorResponse)
	}
}

func TestLoginHandlerWithProvider(t *testing.T) {
	Convey("LoginHandler", t, func() {
		tokenStore := singleTokenStore{}
		conn := singleUserConn{}
		providerRegistry := provider.NewRegistry()
		providerRegistry.RegisterAuthProvider("com.example", handlertest.NewSingleUserAuthProvider("com.example", "johndoe"))

		r := handlertest.NewSingleRouteRouter(LoginHandler, func(p *router.Payload) {
			p.TokenStore = &tokenStore
			p.DBConn = &conn
			p.ProviderRegistry = providerRegistry
		})

		Convey("login in existing", func() {
			userinfo := oddb.NewProvidedAuthUserInfo("com.example:johndoe", map[string]interface{}{"name": "boo"})
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
	userinfo *oddb.UserInfo
	oddb.Conn
}

func (conn *singleUserConn) UpdateUser(userinfo *oddb.UserInfo) error {
	if conn.userinfo != nil && conn.userinfo.ID == userinfo.ID {
		conn.userinfo = userinfo
		return nil
	} else {
		return oddb.ErrUserNotFound
	}
}

func (conn *singleUserConn) CreateUser(userinfo *oddb.UserInfo) error {
	if conn.userinfo == nil {
		conn.userinfo = userinfo
		return nil
	} else {
		return oddb.ErrUserDuplicated
	}
}

func (conn *singleUserConn) GetUser(id string, userinfo *oddb.UserInfo) error {
	if conn.userinfo == nil {
		return oddb.ErrUserNotFound
	} else {
		*userinfo = *conn.userinfo
		return nil
	}
}

func (conn *singleUserConn) GetUserByPrincipalID(principalID string, userinfo *oddb.UserInfo) error {
	if conn.userinfo == nil {
		return oddb.ErrUserNotFound
	} else {
		*userinfo = *conn.userinfo
		return nil
	}
}

func TestSignupHandlerAsAnonymous(t *testing.T) {
	Convey("SignupHandler", t, func() {
		tokenStore := singleTokenStore{}
		conn := singleUserConn{}

		r := handlertest.NewSingleRouteRouter(SignupHandler, func(p *router.Payload) {
			p.TokenStore = &tokenStore
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

		Convey("errors when user id is missing", func() {
			resp := r.POST(`{
				"email": "john.doe@example.com",
				"password": "iamyourfather"
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 101,
		"type": "RequestInvalid",
		"message": "empty user_id, email or password"
	}
}`)
			So(resp.Code, ShouldEqual, 400)
		})

		Convey("errors when email is missing", func() {
			resp := r.POST(`{
				"userid": "someuserid",
				"password": "iamyourfather"
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 101,
		"type": "RequestInvalid",
		"message": "empty user_id, email or password"
	}
}`)
			So(resp.Code, ShouldEqual, 400)
		})

		Convey("errors when password is missing", func() {
			resp := r.POST(`{
				"userid": "someuserid",
				"email": "john.doe@example.com"
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 101,
		"type": "RequestInvalid",
		"message": "empty user_id, email or password"
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

		r := handlertest.NewSingleRouteRouter(SignupHandler, func(p *router.Payload) {
			p.TokenStore = &tokenStore
			p.DBConn = &conn
			p.ProviderRegistry = providerRegistry
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
		"code": 101,
		"type": "AuthenticationError",
		"message": "authentication failed"
	}
}`))
			So(resp.Code, ShouldEqual, 400)
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
		conn := oddbtest.NewMapConn()

		r := handlertest.NewSingleRouteRouter(LogoutHandler, func(p *router.Payload) {
			p.TokenStore = tokenStore
			p.DBConn = conn
		})

		Convey("deletes existing access token", func() {
			resp := r.POST(`{
	"access_token": "someaccesstoken"
}`)
			So(tokenStore.deletedAccessToken, ShouldEqual, "someaccesstoken")
			So(resp.Body.Bytes(), ShouldEqualJSON, `{}`)
			So(resp.Code, ShouldEqual, 200)
		})

		Convey("deletes non-existing access token without error", func() {
			tokenStore.errToReturn = &authtoken.NotFoundError{}
			resp := r.POST(`{
	"access_token": "notexistaccesstoken"
}`)
			So(tokenStore.deletedAccessToken, ShouldEqual, "notexistaccesstoken")
			So(resp.Body.Bytes(), ShouldEqualJSON, `{}`)
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
		"code": 1,
		"type": "UnknownError",
		"message": "some interesting error"
	}
}`)
			So(resp.Code, ShouldEqual, 400)
		})
	})
}

func TestPasswordHandlerWithProvider(t *testing.T) {
	Convey("PasswordHandler", t, func() {
		conn := singleUserConn{}
		userinfo := oddb.NewUserInfo("lord-of-skygear", "limouren@skygear.io", "chima")
		conn.CreateUser(&userinfo)
		tokenStore := singleTokenStore{}
		token := authtoken.New("_", userinfo.ID, time.Time{})
		tokenStore.Put(&token)

		r := handlertest.NewSingleRouteRouter(PasswordHandler, func(p *router.Payload) {
			p.TokenStore = &tokenStore
			p.DBConn = &conn
		})

		Convey("change password success", func() {
			resp := r.POST(fmt.Sprintf(`{
	"access_token": "%s",
	"user_id": "lord-of-skygear",
	"old_password": "chima",
	"password": "faseng"
}`, token.AccessToken))

			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
	"result": {
		"user_id": "lord-of-skygear",
		"access_token": "%s"
	}
}`, token.AccessToken))
			So(resp.Code, ShouldEqual, 200)
		})

	})
}
