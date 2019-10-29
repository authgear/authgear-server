package handler

import (
	"io"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachSignHandler(
	server *server.Server,
	dependencyMap inject.DependencyMap,
) *server.Server {
	server.Handle("/get_signed_url", &SignHandlerFactory{
		dependencyMap,
	}).Methods("OPTIONS", "POST")
	return server
}

type SignHandlerFactory struct {
	DependencyMap inject.DependencyMap
}

func (f *SignHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SignHandler{}
	inject.DefaultRequestInject(h, f.DependencyMap, request)
	return h.RequireAuthz(h, h)
}

// @JSONSchema
const SignAssetResponseSchema = `
{
	"$id": "#SignAssetResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"assets": {
					"type": "array",
					"items": { "$ref": "#SignAssetItem" }
				}
			},
			"required": ["assets"]
		}
	}
}
`

// @JSONSchema
const SignAssetItemSchema = `
{
	"$id": "#SignAssetItem",
	"type": "object",
	"properties": {
		"asset_name": { "type": "string" },
		"url": { "type": "string" }
	},
	"required": ["asset_name", "url"]
}
`

/*
	@Operation POST /get_signed_url - Get signed URL
		Get signed URL of private assets.

		@SecurityRequirement master_key

		@RequestBody
			A list of asset names to be signed.
			@JSONSchema {SignAssetRequest}

		@Response 200
			A list of signed asset urls.
			@JSONSchema {SignAssetResponse}
*/
type SignHandler struct {
	RequireAuthz         handler.RequireAuthz  `dependency:"RequireAuthz"`
	CloudStorageProvider cloudstorage.Provider `dependency:"CloudStorageProvider"`
	Validator            *validation.Validator `dependency:"Validator"`
}

func (h *SignHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

// @JSONSchema
const SignRequestSchema = `
{
	"$id": "#SignAssetRequest",
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"assets": {
			"type": "array",
			"items": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"asset_name": {
						"type": "string",
						"minLength": 1
					}
				},
				"required": ["asset_name"]
			}
		}
	},
	"required": ["assets"]
}
`

func (h *SignHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Err = skyerr.MakeError(err)
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *SignHandler) ParseSignRequest(r io.Reader, p interface{}) error {
	return h.Validator.ParseReader("#SignAssetRequest", r, p)
}

func (h *SignHandler) Handle(w http.ResponseWriter, r *http.Request) (result interface{}, err error) {
	var payload cloudstorage.SignRequest
	err = handler.ParseJSONBody(r, w, h.ParseSignRequest, &payload)
	if err != nil {
		if validationError, ok := err.(validation.Error); ok {
			err = validationError.SkyErrInvalidArgument("Validation Error")
		}
		return
	}

	scheme := coreHttp.GetProto(r)
	host := coreHttp.GetHost(r)
	resp, err := h.CloudStorageProvider.Sign(scheme, host, &payload)
	if err != nil {
		return
	}

	result = resp
	return
}
