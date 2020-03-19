package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachListRecoveryCodeHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/recovery_code/list").
		Handler(server.FactoryToHandler(&ListRecoveryCodeHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type ListRecoveryCodeHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f ListRecoveryCodeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ListRecoveryCodeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type ListRecoveryCodeResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// @JSONSchema
const ListRecoveryCodeResponseSchema = `
{
	"$id": "#ListRecoveryCodeResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"recovery_codes": {
					"type": "array",
					"items": { "type": "string" }
				}
			}
		}
	}
}
`

/*
	@Operation POST /mfa/recovery_code/list - List recovery codes
		List recovery codes if allowed.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200
			List of recovery codes.
			@JSONSchema {ListRecoveryCodeResponse}
*/
type ListRecoveryCodeHandler struct {
	TxContext        db.TxContext            `dependency:"TxContext"`
	RequireAuthz     handler.RequireAuthz    `dependency:"RequireAuthz"`
	MFAProvider      mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration config.MFAConfiguration `dependency:"MFAConfiguration"`
}

func (h *ListRecoveryCodeHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUser
}

func (h *ListRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *ListRecoveryCodeHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload struct{}
	if err := handler.DecodeJSONBody(r, w, &payload); err != nil {
		return nil, err
	}

	if !h.MFAConfiguration.RecoveryCode.ListEnabled {
		return nil, skyerr.NewNotFound("listing recovery code is disabled")
	}

	err = db.WithTx(h.TxContext, func() error {
		userID := auth.GetAuthInfo(r.Context()).ID
		codes, err := h.MFAProvider.GetRecoveryCode(userID)
		if err != nil {
			return err
		}
		resp = ListRecoveryCodeResponse{
			RecoveryCodes: codes,
		}
		return nil
	})
	return
}
