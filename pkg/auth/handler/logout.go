package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// AttachLogoutHandler attach logout handler to server
func AttachLogoutHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/logout", &LogoutHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// LogoutHandlerFactory creates new handler
type LogoutHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new handler
func (f LogoutHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &LogoutHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	h.AuditTrail = h.AuditTrail.WithRequest(request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f LogoutHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

// LogoutRequestPayload is request payload of logout handler
type LogoutRequestPayload struct {
	AccessToken string
}

// Validate request payload
func (p LogoutRequestPayload) Validate() error {
	if p.AccessToken == "" {
		return skyerr.NewError(skyerr.AccessTokenNotAccepted, "missing access token")
	}
	return nil
}

// LogoutHandler handles logout request
type LogoutHandler struct {
	AuthContext coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	TokenStore  authtoken.Store        `dependency:"TokenStore"`
	AuditTrail  audit.Trail            `dependency:"AuditTrail"`
	TxContext   db.TxContext           `dependency:"TxContext"`
}

func (h LogoutHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h LogoutHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := LogoutRequestPayload{}
	payload.AccessToken = model.GetAccessToken(request)
	return payload, nil
}

// Handle api request
func (h LogoutHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(LogoutRequestPayload)

	accessToken := payload.AccessToken

	if err = h.TokenStore.Delete(accessToken); err != nil {
		if _, notfound := err.(*authtoken.NotFoundError); notfound {
			err = nil
		}
	}
	if err != nil {
		err = skyerr.MakeError(err)
	} else {
		resp = "OK"
	}

	h.AuditTrail.Log(audit.Entry{
		AuthID: h.AuthContext.AuthInfo().ID,
		Event:  audit.EventLogout,
	})

	return
}
