package mfa

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachCreateTOTPHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/totp/new", &CreateTOTPHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type CreateTOTPHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f CreateTOTPHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &CreateTOTPHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

func (h *CreateTOTPHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.DenyInvalidSession),
	)
}

type CreateTOTPRequest struct {
	AuthnSessionToken string `json:"authn_session_token"`
	DisplayName       string `json:"display_name"`
	AccountName       string `json:"account_name"`
	Issuer            string `json:"issuer"`
}

type CreateTOTPResponse struct {
	AuthenticatorID   string `json:"authenticator_id"`
	AuthenticatorType string `json:"authenticator_type"`
	Secret            string `json:"secret"`
	OTPAuthURI        string `json:"otpauth_uri"`
	QRCodeImageURI    string `json:"qr_code_image_uri"`
}

// @JSONSchema
const CreateTOTPRequestSchema = `
{
	"$id": "#CreateTOTPRequest",
	"type": "object",
	"properties": {
		"display_name": { "type": "string", "minLength": 1 },
		"account_name": { "type": "string", "minLength": 1 },
		"issuer": { "type": "string", "minLength": 1 },
		"authn_session_token": { "type": "string", "minLength": 1 }
	},
	"required": ["display_name"]
}
`

// @JSONSchema
const CreateTOTPResponseSchema = `
{
	"$id": "#CreateTOTPResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"authenticator_id": { "type": "string" },
				"authenticator_type": { "type": "string" },
				"secret": { "type": "string" },
				"otpauth_uri": { "type": "string" },
				"qr_code_image_uri": { "type": "string" }
			}
		}
	}
}
`

/*
	@Operation POST /mfa/totp/new - Create TOTP authenticator.
		Create inactive TOTP authenticator. It must be activated later.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			@JSONSchema {CreateTOTPRequest}
		@Response 200
			Details of the authenticator
			@JSONSchema {CreateTOTPResponse}
*/
type CreateTOTPHandler struct {
	TxContext            db.TxContext            `dependency:"TxContext"`
	Validator            *validation.Validator   `dependency:"Validator"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	RequireAuthz         handler.RequireAuthz    `dependency:"RequireAuthz"`
	MFAProvider          mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration     config.MFAConfiguration `dependency:"MFAConfiguration"`
	AuthnSessionProvider authnsession.Provider   `dependency:"AuthnSessionProvider"`
}

func (h *CreateTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *CreateTOTPHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload CreateTOTPRequest
	if err := handler.BindJSONBody(r, w, h.Validator, "#CreateTOTPRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() error {
		userID, _, _, err := h.AuthnSessionProvider.Resolve(h.AuthContext, payload.AuthnSessionToken, authnsession.ResolveOptions{
			MFAOption: authnsession.ResolveMFAOptionOnlyWhenNoAuthenticators,
		})
		if err != nil {
			return err
		}
		a, err := h.MFAProvider.CreateTOTP(userID, payload.DisplayName)
		if err != nil {
			return err
		}
		keyURI := mfa.NewKeyURI(payload.Issuer, payload.AccountName, a.Secret)
		qrCodeImageURI, err := keyURI.QRCodeDataURI()
		if err != nil {
			return err
		}
		resp = CreateTOTPResponse{
			AuthenticatorID:   a.ID,
			AuthenticatorType: string(a.Type),
			Secret:            a.Secret,
			OTPAuthURI:        keyURI.String(),
			QRCodeImageURI:    qrCodeImageURI,
		}
		return nil
	})
	return
}
