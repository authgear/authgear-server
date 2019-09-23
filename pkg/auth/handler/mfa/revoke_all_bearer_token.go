package mfa

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachRevokeAllBearerTokenHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/bearer_token/revoke_all", &RevokeAllBearerTokenHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type RevokeAllBearerTokenHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f RevokeAllBearerTokenHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RevokeAllBearerTokenHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h.AuthContext, h)
}

/*
	@Operation POST /mfa/bearer_token/revoke_all - Revoke all bearer tokens.
		Revoke all bearer tokens.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200 {EmptyResponse}
*/
type RevokeAllBearerTokenHandler struct {
	TxContext        db.TxContext            `dependency:"TxContext"`
	AuthContext      coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	MFAProvider      mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration config.MFAConfiguration `dependency:"MFAConfiguration"`
}

func (h *RevokeAllBearerTokenHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h *RevokeAllBearerTokenHandler) WithTx() bool {
	return true
}

func (h *RevokeAllBearerTokenHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := handler.EmptyRequestPayload{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h *RevokeAllBearerTokenHandler) Handle(req interface{}) (resp interface{}, err error) {
	authInfo, _ := h.AuthContext.AuthInfo()
	userID := authInfo.ID
	err = h.MFAProvider.DeleteAllBearerToken(userID)
	if err != nil {
		return
	}
	resp = map[string]interface{}{}
	return resp, nil
}
