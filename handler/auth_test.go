package handler

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oddb/fs"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

func tempDir() string {
	dir, err := ioutil.TempDir("", "ourd.oddb.handler.auth.test")
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
	dir := tempDir()
	defer os.RemoveAll(dir)

	conn, err := fs.Open("com.oursky.ourd.handler.auth", dir)
	if err != nil {
		panic(err)
	}

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
	dir := tempDir()
	defer os.RemoveAll(dir)

	conn, err := fs.Open("com.oursky.ourd.handler.auth", dir)
	if err != nil {
		panic(err)
	}

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

	errorResponse, ok := resp.Err.(oderr.Error)
	if !ok {
		t.Fatalf("got type = %T, want type oderr.Error", resp.Err)
	}

	if errorResponse.Code() != 101 {
		t.Fatalf("got errorResponse.Code() = %v, want 101", errorResponse.Code())
	}
}

func TestLoginHandler(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)

	conn, err := fs.Open("com.oursky.ourd.handler.auth", dir)
	if err != nil {
		panic(err)
	}

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
	dir := tempDir()
	defer os.RemoveAll(dir)

	conn, err := fs.Open("com.oursky.ourd.handler.auth", dir)
	if err != nil {
		panic(err)
	}

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

	errorResponse, ok := resp.Err.(oderr.Error)
	if !ok {
		t.Fatalf("got type = %T, want type oderr.Error", resp.Err)
	}

	if errorResponse != oderr.ErrAuthFailure {
		t.Fatalf("got resp.Err = %v, want ErrAuthFailure", errorResponse)
	}
}

func TestLoginHandlerNotFound(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)

	conn, err := fs.Open("com.oursky.ourd.handler.auth", dir)
	if err != nil {
		panic(err)
	}

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

	errorResponse, ok := resp.Err.(oderr.Error)
	if !ok {
		t.Fatalf("got type = %T, want type oderr.Error", resp.Err)
	}

	if errorResponse != oderr.ErrUserNotFound {
		t.Fatalf("got resp.Err = %v, want ErrUserNotFound", errorResponse)
	}
}
