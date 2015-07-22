package handler

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

type authResponse struct {
	UserID      string `json:"user_id,omitempty"`
	Email       string `json:"email,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

type signupPayload struct {
	AppName string
	Meta    map[string]interface{}
	Data    map[string]interface{}
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
	return p.Email() == "" && p.Password() == "" && p.UserID() == "" && p.Provider() == ""
}

func (p *signupPayload) Provider() string {
	provider, _ := p.Data["provider"].(string)
	return provider
}

func (p *signupPayload) AuthData() map[string]interface{} {
	authData, _ := p.Data["auth_data"].(map[string]interface{})
	return authData
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
func SignupHandler(payload *router.Payload, response *router.Response) {
	store := payload.TokenStore

	p := signupPayload{
		AppName: payload.AppName,
		Meta:    payload.Meta,
		Data:    payload.Data,
	}

	info := oddb.UserInfo{}
	if p.IsAnonymous() {
		info = oddb.NewAnonymousUserInfo()
	} else if p.Provider() != "" {
		// Get AuthProvider and authenticates the user
		log.Debugf(`Client requested auth provider: "%v".`, p.Provider())
		authProvider := payload.ProviderRegistry.GetAuthProvider(p.Provider())
		principalID, authData, err := authProvider.Login(p.AuthData())
		if err != nil {
			response.Err = oderr.ErrAuthFailure
			return
		}
		log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider())

		// Create new user info and set updated auth data
		info = oddb.NewProvidedAuthUserInfo(principalID, authData)
	} else {
		userID := p.UserID()
		email := p.Email()
		password := p.Password()

		if userID == "" || email == "" || password == "" {
			response.Err = oderr.NewRequestInvalidErr(errors.New("empty user_id, email or password"))
			return
		}
		info = oddb.NewUserInfo(userID, email, password)
	}

	if err := payload.DBConn.CreateUser(&info); err != nil {
		if err == oddb.ErrUserDuplicated {
			response.Err = oderr.ErrUserDuplicated
		} else {
			response.Err = oderr.NewResourceSaveFailureErrWithStringID("user", p.UserID())
		}
		return
	}

	// generate access-token
	token := authtoken.New(p.AppName, info.ID, time.Time{})
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
	AppName string
	Meta    map[string]interface{}
	Data    map[string]interface{}
}

func (p *loginPayload) RouteAction() string {
	return "auth:login"
}

func (p *loginPayload) Provider() string {
	provider, _ := p.Data["provider"].(string)
	return provider
}

func (p *loginPayload) AuthData() map[string]interface{} {
	authData, _ := p.Data["auth_data"].(map[string]interface{})
	return authData
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
func LoginHandler(payload *router.Payload, response *router.Response) {
	store := payload.TokenStore

	p := loginPayload{
		AppName: payload.AppName,
		Meta:    payload.Meta,
		Data:    payload.Data,
	}

	info := oddb.UserInfo{}

	if p.Provider() != "" {
		// Get AuthProvider and authenticates the user
		log.Debugf(`Client requested auth provider: "%v".`, p.Provider())
		authProvider := payload.ProviderRegistry.GetAuthProvider(p.Provider())
		principalID, _, err := authProvider.Login(p.AuthData())
		if err != nil {
			response.Err = oderr.ErrAuthFailure
			return
		}
		log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider())

		if err := payload.DBConn.GetUserByPrincipalID(principalID, &info); err != nil {
			if err == oddb.ErrUserNotFound {
				response.Err = oderr.ErrUserNotFound
			} else {
				// TODO: more error handling here if necessary
				response.Err = oderr.NewResourceFetchFailureErr("user", p.UserID())
			}
			return
		}
	} else {
		if err := payload.DBConn.GetUser(p.UserID(), &info); err != nil {
			if err == oddb.ErrUserNotFound {
				response.Err = oderr.ErrUserNotFound
			} else {
				// TODO: more error handling here if necessary
				response.Err = oderr.NewResourceFetchFailureErr("user", p.UserID())
			}
			return
		}

		if !info.IsSamePassword(p.Password()) {
			response.Err = oderr.ErrInvalidLogin
			return
		}
	}

	// generate access-token
	token := authtoken.New(p.AppName, info.ID, time.Time{})
	if err := store.Put(&token); err != nil {
		panic(err)
	}

	response.Result = authResponse{
		UserID:      info.ID,
		Email:       info.Email,
		AccessToken: token.AccessToken,
	}
}

// LogoutHandler receives an access token and invalidates it
func LogoutHandler(payload *router.Payload, response *router.Response) {
	store := payload.TokenStore
	accessToken := payload.AccessToken()

	if err := store.Delete(accessToken); err != nil {
		if _, notfound := err.(*authtoken.NotFoundError); !notfound {
			response.Err = oderr.NewUnknownErr(err)
		}
	}
}
