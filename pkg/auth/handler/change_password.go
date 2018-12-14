package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
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

func AttachChangePasswordHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/change_password", &ChangePasswordHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// ChangePasswordHandlerFactory creates ChangePasswordHandler
type ChangePasswordHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new handler
func (f ChangePasswordHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ChangePasswordHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f ChangePasswordHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type ChangePasswordRequestPayload struct {
	NewPassword string `json:"password"`
	OldPassword string `json:"old_password"`
}

func (p ChangePasswordRequestPayload) Validate() error {
	if p.OldPassword == "" {
		return skyerr.NewInvalidArgument("empty old password", []string{"old_password"})
	}
	if p.NewPassword == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}
	return nil
}

// ChangePasswordHandler change the current user password
//
// ChangePasswordHandler receives old and new password:
//
// * old_password (string, required)
// * password (string, required)
//
// If user is not logged in, an 401 unauthorized will return.
//
//  Current implementation
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/change_password <<EOF
//  {
//      "old_password": "oldpassword",
//      "password": "newpassword"
//  }
//  EOF
// Response
// return auth response with new access token
type ChangePasswordHandler struct {
	AuditTrail           audit.Trail                `dependency:"AuditTrail"`
	AuthContext          coreAuth.ContextGetter     `dependency:"AuthContextGetter"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
	PasswordChecker      *authAudit.PasswordChecker `dependency:"PasswordChecker"`
	TokenStore           authtoken.Store            `dependency:"TokenStore"`
	TxContext            db.TxContext               `dependency:"TxContext"`
	UserProfileStore     userprofile.Store          `dependency:"UserProfileStore"`
}

func (h ChangePasswordHandler) WithTx() bool {
	return true
}

// DecodeRequest decode the request payload
func (h ChangePasswordHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ChangePasswordRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h ChangePasswordHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ChangePasswordRequestPayload)
	authinfo := h.AuthContext.AuthInfo()

	if err = h.PasswordChecker.ValidatePassword(authAudit.ValidatePasswordPayload{
		PlainPassword: payload.NewPassword,
		AuthID:        authinfo.ID,
	}); err != nil {
		return
	}

	principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(authinfo.ID)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return
		}
		return
	}

	for _, p := range principals {
		if !p.IsSamePassword(payload.OldPassword) {
			err = skyerr.NewError(skyerr.InvalidCredentials, "Incorrect old password")
			return
		}
		p.PlainPassword = payload.NewPassword
		err = h.PasswordAuthProvider.UpdatePrincipal(*p)
		if err != nil {
			return
		}
	}

	now := timeNow()
	authinfo.TokenValidSince = &now
	err = h.AuthInfoStore.UpdateAuth(authinfo)
	if err != nil {
		return
	}

	// generate access-token
	token, err := h.TokenStore.NewToken(authinfo.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	// Get Profile
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.GetUserProfile(authinfo.ID, token.AccessToken); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to fetch user profile")
		return
	}

	resp = response.NewAuthResponse(*authinfo, userProfile, token.AccessToken)
	h.AuditTrail.Log(audit.Entry{
		AuthID: authinfo.ID,
		Event:  audit.EventChangePassword,
	})

	return
}
