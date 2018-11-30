package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// AttachLoginHandler attach login handler to server
func AttachLoginHandler(
	server *server.Server,
	authDependency auth.RequestDependencyMap,
) *server.Server {
	server.Handle("/login", &LoginHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// LoginHandlerFactory creates new handler
type LoginHandlerFactory struct {
	Dependency auth.RequestDependencyMap
}

func (f LoginHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LoginHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f LoginHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

// LoginRequestPayload login handler request payload
type LoginRequestPayload struct {
	AuthData map[string]interface{} `json:"auth_data"`
	Password string                 `json:"password"`
}

// Validate request payload
func (p LoginRequestPayload) Validate() error {
	if len(p.AuthData) == 0 {
		return skyerr.NewInvalidArgument("empty auth data", []string{"auth_data"})
	}

	if p.Password == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}

	return nil
}

// LoginHandler handles login request
type LoginHandler struct {
	TokenStore           authtoken.Store   `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store    `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider `dependency:"PasswordAuthProvider"`
	UserProfileStore     userprofile.Store `dependency:"UserProfileStore"`
	AuditTrail           audit.Trail       `dependency:"AuditTrail"`
	TxContext            db.TxContext      `dependency:"TxContext"`
}

func (h LoginHandler) WithTx() bool {
	return true
}

// ProvideAuthzPolicy provides authorization policy
func (h LoginHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

// DecodeRequest decode request payload
func (h LoginHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := LoginRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

// Handle api request
func (h LoginHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(LoginRequestPayload)
	fetchedAuthInfo := authinfo.AuthInfo{}

	defer func() {
		if err != nil {
			h.AuditTrail.Log(audit.Entry{
				AuthID: fetchedAuthInfo.ID,
				Event:  audit.EventLoginFailure,
			})
		} else {
			h.AuditTrail.Log(audit.Entry{
				AuthID: fetchedAuthInfo.ID,
				Event:  audit.EventLoginSuccess,
			})
		}
	}()

	if valid := h.PasswordAuthProvider.IsAuthDataValid(payload.AuthData); !valid {
		err = skyerr.NewInvalidArgument("invalid auth data", []string{"auth_data"})
		return
	}

	userID, err := h.getUserID(payload.Password, payload.AuthData)
	if err != nil {
		return
	}

	if err = h.AuthInfoStore.GetAuth(userID, &fetchedAuthInfo); err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return
		}
		// TODO: more error handling here if necessary
		err = skyerr.NewResourceFetchFailureErr("auth_data", payload.AuthData)
		return
	}

	if err = checkUserIsNotDisabled(&fetchedAuthInfo); err != nil {
		return
	}

	// generate access-token
	token, err := h.TokenStore.NewToken(fetchedAuthInfo.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	// Get Profile
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.GetUserProfile(fetchedAuthInfo.ID, token.AccessToken); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to fetch user profile")
		return
	}

	resp = response.NewAuthResponse(fetchedAuthInfo, userProfile, token.AccessToken)

	// Populate the activity time to user
	now := timeNow()
	fetchedAuthInfo.LastLoginAt = &now
	fetchedAuthInfo.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(&fetchedAuthInfo); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	return
}

func (h LoginHandler) getUserID(pwd string, authData map[string]interface{}) (userID string, err error) {
	principals, err := h.PasswordAuthProvider.GetPrincipalsByAuthData(authData)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return
		}
		// TODO: more error handling here if necessary
		err = skyerr.NewResourceFetchFailureErr("auth_data", authData)
		return
	}

	for _, principal := range principals {
		if !principal.IsSamePassword(pwd) {
			err = skyerr.NewError(skyerr.InvalidCredentials, "auth_data or password incorrect")
			return
		}
	}

	userID = principals[0].UserID

	return
}
