package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// AttachLogoutHandler attach logout handler to server
func AttachLogoutHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/logout", &LogoutHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

// LogoutHandlerFactory creates new handler
type LogoutHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new handler
func (f LogoutHandlerFactory) NewHandler(request *http.Request) handler.Handler {
	h := &LogoutHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h)
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
	AccessToken         string
}

// Validate request payload
func (p LogoutRequestPayload) Validate() error {
	return nil
}

// LogoutHandler handles logout request
type LogoutHandler struct{
	TokenStore           authtoken.Store            `dependency:"TokenStore"`
}

// DecodeRequest decode request payload
func (h LogoutHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := LogoutRequestPayload{}
	payload.AccessToken = model.GetAccessToken(request)
	return payload, nil
}

// Handle api request
func (h LogoutHandler) Handle(req interface{}, ctx handler.AuthContext) (resp interface{}, err error) {
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

	// TODO: Audit
	return
}
