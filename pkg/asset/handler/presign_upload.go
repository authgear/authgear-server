package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachPresignUploadHandler(
	router *mux.Router,
	dependencyMap inject.DependencyMap,
) {
	router.NewRoute().
		Path("/presign_upload").
		Handler(server.FactoryToHandler(&PresignUploadHandlerFactory{
			dependencyMap,
		})).
		Methods("OPTIONS", "POST")
}

type PresignUploadHandlerFactory struct {
	DependencyMap inject.DependencyMap
}

func (f *PresignUploadHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &PresignUploadHandler{}
	inject.DefaultRequestInject(h, f.DependencyMap, request)
	return h.RequireAuthz(h, h)
}

// @JSONSchema
const PresignUploadResponseSchema = `
{
	"$id": "#PresignUploadResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"asset_name": { "type": "string" },
				"url": { "type": "string" },
				"method": { "type": "string" },
				"headers": {
					"type": "array",
					"items": {
						"type": "object",
						"properties": {
							"name": { "type": "string" },
							"value": { "type": "string" }
						},
						"required": ["name", "value"]
					}
				}
			},
			"required": ["asset_name", "url", "method", "headers"]
		}
	}
}
`

/*
	@Operation POST /presign_upload - Presign an upload request.
		Presign an upload request.

		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			@JSONSchema {PresignUploadRequest}

		@Response 200
			@JSONSchema {PresignUploadResponse}
*/
type PresignUploadHandler struct {
	RequireAuthz         handler.RequireAuthz  `dependency:"RequireAuthz"`
	CloudStorageProvider cloudstorage.Provider `dependency:"CloudStorageProvider"`
	Validator            *validation.Validator `dependency:"Validator"`
}

func (h *PresignUploadHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUserOrMasterKey
}

// @JSONSchema
const PresignUploadRequestSchema = `
{
	"$id": "#PresignUploadRequest",
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"prefix": {
			"type": "string",
			"pattern": "^[-_.a-zA-Z0-9]*$"
		},
		"access": { "type": "string", "enum": ["public", "private"] },
		"headers": {
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"content-type": { "type": "string" },
				"content-disposition": { "type": "string" },
				"content-encoding": { "type": "string" },
				"content-length": { "type": "string" },
				"content-md5": { "type": "string" },
				"cache-control": { "type": "string" },
				"access-control-allow-origin": { "type": "string" },
				"access-control-expose-headers": { "type": "string" },
				"access-control-max-age": { "type": "string" },
				"access-control-allow-credentials": { "type": "string" },
				"access-control-allow-methods": { "type": "string" },
				"access-control-allow-headers": { "type": "string" }
			},
			"required": ["content-length"]
		}
	},
	"required": ["headers"]
}
`

func (h *PresignUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *PresignUploadHandler) Handle(w http.ResponseWriter, r *http.Request) (result interface{}, err error) {
	// Parse request
	var payload cloudstorage.PresignUploadRequest
	err = handler.BindJSONBody(r, w, h.Validator, "#PresignUploadRequest", &payload)
	if err != nil {
		return
	}

	resp, err := h.CloudStorageProvider.PresignPutRequest(&payload)
	if err != nil {
		return
	}

	result = resp
	return
}
