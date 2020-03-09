package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

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

func AttachRegenerateRecoveryCodeHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/recovery_code/regenerate").
		Handler(server.FactoryToHandler(&RegenerateRecoveryCodeHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type RegenerateRecoveryCodeHandlerFactory struct {
	Dependency auth.DependencyMap
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
	TxContext        db.TxContext            `dependency:"TxContext"`
	AuthContext      coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	RequireAuthz     handler.RequireAuthz    `dependency:"RequireAuthz"`
	MFAProvider      mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration config.MFAConfiguration `dependency:"MFAConfiguration"`
}

func (h *RegenerateRecoveryCodeHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUser
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
		authInfo, _ := h.AuthContext.AuthInfo()
		userID := authInfo.ID
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
