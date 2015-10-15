package handler

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/authtoken"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
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

	info := skydb.UserInfo{}
	if p.IsAnonymous() {
		info = skydb.NewAnonymousUserInfo()
	} else if p.Provider() != "" {
		// Get AuthProvider and authenticates the user
		log.Debugf(`Client requested auth provider: "%v".`, p.Provider())
		authProvider := payload.ProviderRegistry.GetAuthProvider(p.Provider())
		principalID, authData, err := authProvider.Login(p.AuthData())
		if err != nil {
			response.Err = skyerr.ErrAuthFailure
			return
		}
		log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider())

		// Create new user info and set updated auth data
		info = skydb.NewProvidedAuthUserInfo(principalID, authData)
	} else {
		userID := p.UserID()
		email := p.Email()
		password := p.Password()

		if userID == "" || email == "" || password == "" {
			response.Err = skyerr.NewRequestInvalidErr(errors.New("empty user_id, email or password"))
			return
		}
		info = skydb.NewUserInfo(userID, email, password)
	}

	if err := payload.DBConn.CreateUser(&info); err != nil {
		if err == skydb.ErrUserDuplicated {
			response.Err = skyerr.ErrUserDuplicated
		} else {
			response.Err = skyerr.NewResourceSaveFailureErrWithStringID("user", p.UserID())
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

	info := skydb.UserInfo{}

	if p.Provider() != "" {
		// Get AuthProvider and authenticates the user
		log.Debugf(`Client requested auth provider: "%v".`, p.Provider())
		authProvider := payload.ProviderRegistry.GetAuthProvider(p.Provider())
		principalID, authData, err := authProvider.Login(p.AuthData())
		if err != nil {
			response.Err = skyerr.ErrAuthFailure
			return
		}
		log.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider())

		if err := payload.DBConn.GetUserByPrincipalID(principalID, &info); err != nil {
			// Create user if and only if no user found with the same principal
			if err != skydb.ErrUserNotFound {
				// TODO: more error handling here if necessary
				response.Err = skyerr.NewResourceFetchFailureErr("user", p.UserID())
				return
			}

			info = skydb.NewProvidedAuthUserInfo(principalID, authData)
			if err = payload.DBConn.CreateUser(&info); err != nil {
				if err == skydb.ErrUserDuplicated {
					response.Err = skyerr.ErrUserDuplicated
				} else {
					response.Err = skyerr.NewResourceSaveFailureErrWithStringID("user", p.UserID())
				}
				return
			}
		} else {
			info.SetProvidedAuthData(principalID, authData)
			if err := payload.DBConn.UpdateUser(&info); err != nil {
				response.Err = skyerr.NewUnknownErr(err)
				return
			}
		}
	} else {
		if err := payload.DBConn.GetUser(p.UserID(), &info); err != nil {
			if err == skydb.ErrUserNotFound {
				response.Err = skyerr.ErrUserNotFound
			} else {
				// TODO: more error handling here if necessary
				response.Err = skyerr.NewResourceFetchFailureErr("user", p.UserID())
			}
			return
		}

		if !info.IsSamePassword(p.Password()) {
			response.Err = skyerr.ErrInvalidLogin
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
			response.Err = skyerr.NewUnknownErr(err)
		}
	}
}

// Define the playload that change password handler will process
type passwordPayload struct {
	AppName    string
	Data       map[string]interface{}
	UserInfoID string
}

func (p *passwordPayload) RouteAction() string {
	return "auth:password"
}

func (p *passwordPayload) OldPassword() string {
	oldPassword, _ := p.Data["old_password"].(string)
	return oldPassword
}

func (p *passwordPayload) NewPassword() string {
	password, _ := p.Data["password"].(string)
	return password
}

func (p *passwordPayload) Invalidate() bool {
	invalidate, _ := p.Data["invalidate"].(bool)
	return invalidate
}

// PasswordHandler change the current user password
//
// PasswordHandler receives three parameters:
//
// * old_password (string, required)
// * password (string, required)
//
// If user is not logged in, an 404 not found will return.
//
//  Current implementation
//	curl -X POST -H "Content-Type: application/json" \
//	  -d @- http://localhost:3000/ <<EOF
//	{
//	    "action": "auth:password",
//	    "old_password": "rick.mak@gmail.com",
//	    "password": "123456"
//	}
//	EOF
// Response
// return existing access toektn if not invalidate
//
// TODO:
// Input accept `user_id` and `invalidate`.
// If `user_id` is supplied, will check authorization policy and see if existing
// accept `invalidate` and invaldate all existing access token.
// Return userInfoID with new AccessToken if the invalidate is true
func PasswordHandler(payload *router.Payload, response *router.Response) {
	log.Debugf("changing password")
	p := passwordPayload{
		AppName:    payload.AppName,
		Data:       payload.Data,
		UserInfoID: payload.UserInfoID,
	}
	info := skydb.UserInfo{}
	if err := payload.DBConn.GetUser(p.UserInfoID, &info); err != nil {
		if err == skydb.ErrUserNotFound {
			response.Err = skyerr.ErrUserNotFound
		} else {
			// TODO: more error handling here if necessary
			response.Err = skyerr.NewResourceFetchFailureErr("user", p.UserInfoID)
		}
		return
	}

	if !info.IsSamePassword(p.OldPassword()) {
		log.Debug("Incorrecly Old Password")
		response.Err = skyerr.NewUnknownErr(errors.New("Incorrecly Old Password"))
		return
	}
	info.SetPassword(p.NewPassword())
	if err := payload.DBConn.UpdateUser(&info); err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}

	if p.Invalidate() {
		log.Warningf("Invalidate is not yet implement")
		// TODO: invalidate all existing token and generate a new one for response
	}
	response.Result = authResponse{
		UserID:      info.ID,
		AccessToken: payload.AccessToken(),
	}
}
