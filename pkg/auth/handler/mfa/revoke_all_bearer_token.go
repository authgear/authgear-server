package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachRevokeAllBearerTokenHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/bearer_token/revoke_all").
		Handler(server.FactoryToHandler(&RevokeAllBearerTokenHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type RevokeAllBearerTokenHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f RevokeAllBearerTokenHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RevokeAllBearerTokenHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
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
	RequireAuthz     handler.RequireAuthz    `dependency:"RequireAuthz"`
	MFAProvider      mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration config.MFAConfiguration `dependency:"MFAConfiguration"`
}

func (h *RevokeAllBearerTokenHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h *RevokeAllBearerTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *RevokeAllBearerTokenHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload struct{}
	if err := handler.DecodeJSONBody(r, w, &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() error {
		userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
		err = h.MFAProvider.DeleteAllBearerToken(userID)
		if err != nil {
			return err
		}
		resp = struct{}{}
		return nil
	})
	return
}
