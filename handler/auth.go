package handler

import (
	_ "log"
	"time"

	"github.com/oursky/ourd/auth"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/router"
)

// TokenStore is the interface for access and storage of access token.
type TokenStore auth.TokenStore

// Authentication generates handlers used for authentication purposes.
//
// Authentication relies on a TokenStore to get and set access token
// of users.
type Authentication struct {
	TokenStore
}

// SignupHandler returns a handler to sign up user using own TokenStore
func (au *Authentication) SignupHandler() func(*router.Payload, *router.Response) {
	return func(payload *router.Payload, response *router.Response) {
		SignupHandler(payload, response, au.TokenStore)
	}
}

// LoginHandler returns a handler to log user in using own TokenStore
func (au *Authentication) LoginHandler() func(*router.Payload, *router.Response) {
	return func(payload *router.Payload, response *router.Response) {
		LoginHandler(payload, response, au.TokenStore)
	}
}

type authResponse struct {
	UserID      string `json:"user_id,omitempty"`
	Email       string `json:"email,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

type signupPayload struct {
	Meta map[string]interface{}
	Data map[string]interface{}
}

func (p *signupPayload) RouteAction() string {
	return "auth:signup"
}

func (p *signupPayload) Email() string {
	email, _ := p.Data["email"].(string)
	return email
}

func (p *signupPayload) Password() string {
	password, _ := p.Data["password"].(string)
	return password
}

func (p *signupPayload) UserID() string {
	userID, _ := p.Data["user_id"].(string)
	return userID
}

func (p *signupPayload) IsAnonymous() bool {
	return p.UserID() == ""
}

// SignupHandler creates an UserInfo with the supplied information.
//
// SignupHandler receives three parameters:
//
// * user_id (string, unique, optional)
// * email  (string, optional)
// * password (string, optional)
//
// If user_id is not supplied, an anonymous user is created and
// have user_id auto-generated. SignupHandler writes an error to
// response.Result if the supplied user_id collides with an existing
// user_id.
//
//	curl -X POST -H "Content-Type: application/json" \
//	  -d @- http://localhost:3000/ <<EOF
//	{
//	    "action": "auth:signup",
//	    "user_id": "rick.mak@gmail.com",
//	    "email": "rick.mak@gmail.com",
//	    "password": "123456"
//	}
//	EOF
func SignupHandler(payload *router.Payload, response *router.Response, store TokenStore) {
	p := signupPayload{
		Meta: payload.Meta,
		Data: payload.Data,
	}

	info := oddb.UserInfo{}
	if p.IsAnonymous() {
		info = oddb.NewAnonymousUserInfo()
	} else {
		userID := p.UserID()
		email := p.Email()
		password := p.Password()

		info = oddb.NewUserInfo(userID, email, password)
	}

	if err := payload.DBConn.CreateUser(&info); err != nil {
		if err == oddb.ErrUserDuplicated {
			response.Result = NewError(UserIDDuplicatedErr, "User with the same ID already existed")
		} else {
			// TODO: more error handling here if necessary
			response.Result = NewError(UnknownErr, "Unknown error occurred.")
		}
		return
	}

	// generate access-token
	token := auth.NewToken(info.ID, time.Time{})
	if err := store.Put(&token); err != nil {
		panic(err)
	}

	response.Result = authResponse{
		UserID:      info.ID,
		Email:       info.Email,
		AccessToken: token.AccessToken,
	}
}

type loginPayload struct {
	Meta map[string]interface{}
	Data map[string]interface{}
}

func (p *loginPayload) RouteAction() string {
	return "auth:login"
}

func (p *loginPayload) UserID() string {
	userID, _ := p.Data["user_id"].(string)
	return userID
}

func (p *loginPayload) Password() string {
	password, _ := p.Data["password"].(string)
	return password
}

/*
LoginHandler is dummy implementation on handling login
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "auth:login",
    "user_id": "rick.mak@gmail.com",
    "password": "123456"
}
EOF
*/
func LoginHandler(payload *router.Payload, response *router.Response, store TokenStore) {
	p := loginPayload{
		Meta: payload.Meta,
		Data: payload.Data,
	}

	info := oddb.UserInfo{}
	if err := payload.DBConn.GetUser(p.UserID(), &info); err != nil {
		if err == oddb.ErrUserNotFound {
			response.Result = NewError(UserIDNotFoundErr, "Cannot find User with the specified ID")
		} else {
			// TODO: more error handling here if necessary
			response.Result = NewError(UnknownErr, "Unknown error")
		}
		return
	}

	if !info.IsSamePassword(p.Password()) {
		response.Result = NewError(RequestInvalidErr, "Invalid login information")
		return
	}

	// generate access-token
	token := auth.NewToken(info.ID, time.Time{})
	if err := store.Put(&token); err != nil {
		panic(err)
	}

	response.Result = authResponse{
		UserID:      info.ID,
		Email:       info.Email,
		AccessToken: token.AccessToken,
	}
}
