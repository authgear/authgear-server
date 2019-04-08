package sso

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/customtoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	signUpHandler "github.com/skygeario/skygear-server/pkg/auth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
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

// AttachCustomTokenLoginHandler attaches CustomTokenLoginHandler to server
func AttachCustomTokenLoginHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/custom_token/login", &CustomTokenLoginHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// CustomTokenLoginHandlerFactory creates CustomTokenLoginHandler
type CustomTokenLoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new CustomTokenLoginHandler
func (f CustomTokenLoginHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &CustomTokenLoginHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f CustomTokenLoginHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

type customTokenLoginPayload struct {
	TokenString string                           `json:"token"`
	Claims      customtoken.SSOCustomTokenClaims `json:"-"`
}

func (payload customTokenLoginPayload) Validate() error {
	if err := payload.Claims.Validate(); err != nil {
		return skyerr.NewError(
			skyerr.InvalidCredentials,
			err.Error(),
		)
	}
	return nil
}

/*
CustomTokenLoginHandler authenticates the user with a custom token

An external server is responsible for generating the custom token which
contains a Principal ID and a signature. It is required that the token
has issued-at and expired-at claims.

The custom token is signed by a shared secret and encoded in JWT format.

The claims of the custom token is as follows:

    {
      "sub": "id1234567800",
      "iat": 1513316033,
      "exp": 1828676033,
      "skyprofile": {
        "name": "John Doe"
      }
    }

When signing the above claims with the custom token secret `ssosecret` using
HS256 as algorithm, the following JWT token is produced:

	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJpZDEyMzQ1Njc4MDAiLCJpYXQiOjE1MTMzMTYwMzMsImV4cCI6MTgyODY3NjAzMywic2t5cHJvZmlsZSI6eyJuYW1lIjoiSm9obiBEb2UifX0.JRAwXPF4CDWCpMCvemCBPrUAQAXPV9qVWeAYo1vBAqQ

This token can be used to log in to Skygear Server. If there is no user
associated with the Token Principal ID (the subject/sub claim), a new user is
created.

curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/sso/custom_token/login <<EOF
{
	"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJpZDEyMzQ1Njc4MDAiLCJpYXQiOjE1MTMzMTYwMzMsImV4cCI6MTgyODY3NjAzMywic2t5cHJvZmlsZSI6eyJuYW1lIjoiSm9obiBEb2UifX0.JRAwXPF4CDWCpMCvemCBPrUAQAXPV9qVWeAYo1vBAqQ"
}
EOF
*/
type CustomTokenLoginHandler struct {
	TxContext               db.TxContext         `dependency:"TxContext"`
	UserProfileStore        userprofile.Store    `dependency:"UserProfileStore"`
	TokenStore              authtoken.Store      `dependency:"TokenStore"`
	AuthInfoStore           authinfo.Store       `dependency:"AuthInfoStore"`
	CustomTokenAuthProvider customtoken.Provider `dependency:"CustomTokenAuthProvider"`
	UserVerifyKeys          []string             `dependency:"UserVerifyKeys"`
	WelcomeEmailEnabled     bool                 `dependency:"WelcomeEmailEnabled"`
	AuditTrail              audit.Trail          `dependency:"AuditTrail"`
	TaskQueue               async.Queue          `dependency:"AsyncTaskQueue"`
}

func (h CustomTokenLoginHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h CustomTokenLoginHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := customTokenLoginPayload{}
	var err error

	defer func() {
		if err != nil {
			h.AuditTrail.Log(audit.Entry{
				Event: audit.EventLoginFailure,
				Data: map[string]interface{}{
					"type": "custom_token",
				},
			})
		}
	}()

	if err = json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	payload.Claims, err = h.CustomTokenAuthProvider.Decode(payload.TokenString)
	if err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, err.Error())
	}
	return payload, err
}

// Handle function handle custom token login
func (h CustomTokenLoginHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(customTokenLoginPayload)
	var info authinfo.AuthInfo
	var userProfile userprofile.UserProfile
	var createNewUser bool

	defer func() {
		var event audit.Event
		if err != nil {
			event = audit.EventLoginFailure
		} else {
			if createNewUser {
				event = audit.EventSignup
			} else {
				event = audit.EventLoginSuccess
			}
		}

		h.AuditTrail.Log(audit.Entry{
			AuthID: info.ID,
			Event:  event,
			Data: map[string]interface{}{
				"type": "custom_token",
			},
		})
	}()

	createNewUser, err = h.handleLogin(payload, &info, &userProfile)
	if err != nil {
		return
	}

	// TODO: check disable

	// Create auth token
	tkn, err := h.TokenStore.NewToken(info.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&tkn); err != nil {
		panic(err)
	}

	respFactory := response.AuthResponseFactory{}
	resp = respFactory.NewAuthResponse(info, userProfile, tkn.AccessToken)

	// Populate the activity time to user
	now := timeNow()
	info.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(&info); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	// TODO: audit trail

	if createNewUser && h.WelcomeEmailEnabled {
		h.sendWelcomeEmail(userProfile)
	}

	return
}

func (h CustomTokenLoginHandler) handleLogin(payload customTokenLoginPayload, info *authinfo.AuthInfo, userProfile *userprofile.UserProfile) (createNewUser bool, err error) {
	createNewUser = false
	principal, err := h.CustomTokenAuthProvider.GetPrincipalByTokenPrincipalID(payload.Claims.Subject)
	if err != nil {
		if err != skydb.ErrUserNotFound {
			return
		}

		err = nil
		createNewUser = true
	}

	if createNewUser {
		now := timeNow()
		*info = authinfo.NewAuthInfo()
		info.LastLoginAt = &now

		// Initialise verify state
		info.VerifyInfo = map[string]bool{}
		for _, key := range h.UserVerifyKeys {
			info.VerifyInfo[key] = false
		}

		// Create AuthInfo
		if e := h.AuthInfoStore.CreateAuth(info); e != nil {
			if e == skydb.ErrUserDuplicated {
				err = signUpHandler.ErrUserDuplicated
				return
			}

			// TODO:
			// return proper error
			err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save auth info")
			return
		}

		principal := customtoken.NewPrincipal()
		principal.TokenPrincipalID = payload.Claims.Subject
		principal.UserID = info.ID
		err = h.CustomTokenAuthProvider.CreatePrincipal(principal)
	} else {
		if e := h.AuthInfoStore.GetAuth(principal.UserID, info); e != nil {
			if err == skydb.ErrUserNotFound {
				err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
				return
			}
			err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
			return
		}
	}

	// Create Profile
	userProfileFunc := func(userID string, authInfo *authinfo.AuthInfo, data userprofile.Data) (userprofile.UserProfile, error) {
		if createNewUser {
			return h.UserProfileStore.CreateUserProfile(userID, authInfo, data)
		}

		return h.UserProfileStore.UpdateUserProfile(userID, authInfo, data)
	}

	if *userProfile, err = userProfileFunc(info.ID, info, payload.Claims.RawProfile); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save user profile")
		return
	}

	return
}

func (h CustomTokenLoginHandler) sendWelcomeEmail(userProfile userprofile.UserProfile) {
	if email, ok := userProfile.Data["email"].(string); ok {
		h.TaskQueue.Enqueue(task.WelcomeEmailSendTaskName, task.WelcomeEmailSendTaskParam{
			Email:       email,
			UserProfile: userProfile,
		}, nil)
	}
}
