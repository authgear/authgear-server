package handler

import (
	"io"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachPresignUploadHandler(
	server *server.Server,
	dependencyMap inject.DependencyMap,
) *server.Server {
	server.Handle("/presign_upload", &PresignUploadHandlerFactory{
		dependencyMap,
	}).Methods("OPTIONS", "POST")
	return server
}

type PresignUploadHandlerFactory struct {
	DependencyMap inject.DependencyMap
}

func (f *PresignUploadHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &PresignUploadHandler{}
	inject.DefaultRequestInject(h, f.DependencyMap, request)
	return h.RequireAuthz(h, h)
}

type PresignUploadHandler struct {
	RequireAuthz         handler.RequireAuthz  `dependency:"RequireAuthz"`
	CloudStorageProvider cloudstorage.Provider `dependency:"CloudStorageProvider"`
	Validator            *validation.Validator `dependency:"Validator"`
}

func (h *PresignUploadHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
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
			"pattern": "^[^\\x00\\\\/:*'<>|]*$"
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
		response.Err = skyerr.MakeError(err)
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *PresignUploadHandler) ParsePresignUploadRequest(r io.Reader, p interface{}) error {
	return h.Validator.ParseReader("#PresignUploadRequest", r, p)
}

func (h *PresignUploadHandler) Handle(w http.ResponseWriter, r *http.Request) (result interface{}, err error) {
	// Parse request
	var payload cloudstorage.PresignUploadRequest
	err = handler.ParseJSONBody(r, w, h.ParsePresignUploadRequest, &payload)
	if err != nil {
		if validationError, ok := err.(validation.Error); ok {
			err = validationError.SkyErrInvalidArgument("Validation Error")
		}
		return
	}
	payload.SetDefaultValue()

	resp, err := h.CloudStorageProvider.PresignPutRequest(&payload)
	if err != nil {
		return
	}

	result = resp
	return
}
