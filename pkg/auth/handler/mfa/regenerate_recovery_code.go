package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachRegenerateRecoveryCodeHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/recovery_code/regenerate").
		Handler(server.FactoryToHandler(&RegenerateRecoveryCodeHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type RegenerateRecoveryCodeHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f RegenerateRecoveryCodeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RegenerateRecoveryCodeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type RegenerateRecoveryCodeResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// @JSONSchema
const RegenerateRecoveryCodeResponseSchema = `
{
	"$id": "#RegenerateRecoveryCodeResponse",
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
	@Operation POST /mfa/recovery_code/regenerate - Regenerate recovery codes
		Regenerate recovery codes. The old ones will no longer valid.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Response 200
			List of newly generated recovery codes.
			@JSONSchema {RegenerateRecoveryCodeResponse}
*/
type RegenerateRecoveryCodeHandler struct {
	TxContext    db.TxContext         `dependency:"TxContext"`
	RequireAuthz handler.RequireAuthz `dependency:"RequireAuthz"`
	MFAProvider  mfa.Provider         `dependency:"MFAProvider"`
}

func (h *RegenerateRecoveryCodeHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h *RegenerateRecoveryCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *RegenerateRecoveryCodeHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload struct{}
	if err := handler.DecodeJSONBody(r, w, &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() error {
		userID := auth.GetSession(r.Context()).AuthnAttrs().UserID
		codes, err := h.MFAProvider.GenerateRecoveryCode(userID)
		if err != nil {
			return err
		}
		resp = RegenerateRecoveryCodeResponse{
			RecoveryCodes: codes,
		}
		return nil
	})
	return
}
