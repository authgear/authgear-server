package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// AttachLoginHandler attach login handler to server
func AttachLoginHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/login", &LoginHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// LoginHandlerFactory creates new handler
type LoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LoginHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LoginHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	h.AuditTrail = h.AuditTrail.WithRequest(request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f LoginHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

// LoginRequestPayload login handler request payload
type LoginRequestPayload struct {
	LoginIDKey string `json:"login_id_key"`
	LoginID    string `json:"login_id"`
	Realm      string `json:"realm"`
	Password   string `json:"password"`
}

// Validate request payload
func (p LoginRequestPayload) Validate() error {
	if p.LoginID == "" {
		return skyerr.NewInvalidArgument("empty login ID", []string{"login_id"})
	}

	if p.Password == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}

	return nil
}

// LoginHandler handles login request
type LoginHandler struct {
	TokenStore           authtoken.Store            `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
	IdentityProvider     principal.IdentityProvider `dependency:"IdentityProvider"`
	UserProfileStore     userprofile.Store          `dependency:"UserProfileStore"`
	AuditTrail           audit.Trail                `dependency:"AuditTrail"`
	TxContext            db.TxContext               `dependency:"TxContext"`
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

	if err != nil {
		return nil, err
	}

	if payload.Realm == "" {
		payload.Realm = password.DefaultRealm
	}
	return payload, nil
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

	if payload.LoginIDKey != "" {
		loginIDs := []password.LoginID{
			password.LoginID{Key: payload.LoginIDKey, Value: payload.LoginID},
		}
		if err = h.PasswordAuthProvider.ValidateLoginIDs(loginIDs); err != nil {
			return
		}
	}

	if valid := h.PasswordAuthProvider.IsRealmValid(payload.Realm); !valid {
		err = skyerr.NewInvalidArgument("realm is not allowed", []string{"realm"})
		return
	}

	principal, err := h.getPrincipal(payload.Password, payload.LoginIDKey, payload.LoginID, payload.Realm)
	if err != nil {
		return
	}

	if err = h.AuthInfoStore.GetAuth(principal.UserID, &fetchedAuthInfo); err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return
		}
		// TODO: more error handling here if necessary
		err = skyerr.NewResourceFetchFailureErr("login_id", payload.LoginID)
		return
	}

	if err = checkUserIsNotDisabled(&fetchedAuthInfo); err != nil {
		return
	}

	// generate access-token
	token, err := h.TokenStore.NewToken(fetchedAuthInfo.ID, principal.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	// Get Profile
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.GetUserProfile(fetchedAuthInfo.ID); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to fetch user profile")
		return
	}

	// Populate the activity time to user
	now := timeNow()
	fetchedAuthInfo.LastLoginAt = &now
	fetchedAuthInfo.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(&fetchedAuthInfo); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	user := model.NewUser(fetchedAuthInfo, userProfile)
	identity := model.NewIdentity(h.IdentityProvider, &principal)
	resp = model.NewAuthResponse(user, identity, token.AccessToken)

	return
}

func (h LoginHandler) getPrincipal(pwd string, loginIDKey string, loginID string, realm string) (principal password.Principal, err error) {
	err = h.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm(loginIDKey, loginID, realm, &principal)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return
		}
		// TODO: more error handling here if necessary
		err = skyerr.NewResourceFetchFailureErr("login_id", loginID)
		return
	}

	if !principal.IsSamePassword(pwd) {
		err = skyerr.NewError(skyerr.InvalidCredentials, "login_id or password incorrect")
		return
	}

	return
}
