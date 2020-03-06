package handler

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/asset/dependency/presign"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachPresignUploadFormHandler(
	router *mux.Router,
	dependencyMap inject.DependencyMap,
) {
	router.NewRoute().
		Path("/presign_upload_form").
		Handler(server.FactoryToHandler(&PresignUploadFormHandlerFactory{
			dependencyMap,
		})).
		Methods("OPTIONS", "POST")
}

type PresignUploadFormHandlerFactory struct {
	DependencyMap inject.DependencyMap
}

func (f *PresignUploadFormHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &PresignUploadFormHandler{}
	inject.DefaultRequestInject(h, f.DependencyMap, request)
	return h.RequireAuthz(h, h)
}

// @JSONSchema
const PresignUploadFormRequestSchema = `
{
	"$id": "#PresignUploadFormRequest",
	"type": "object",
	"additionalProperties": false
}
`

// @JSONSchema
const PresignUploadFormResponse = `
{
	"$id": "#PresignUploadFormResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"url": { "type": "string" }
			},
			"required": ["url"]
		}
	}
}
`

/*
	@Operation POST /presign_upload_form - Presign an upload form request.
		Presign an upload form request.

		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			@JSONSchema {PresignUploadFormRequest}

		@Response 200
			@JSONSchema {PresignUploadFormResponse}
*/
type PresignUploadFormHandler struct {
	RequireAuthz    handler.RequireAuthz `dependency:"RequireAuthz"`
	PresignProvider presign.Provider     `dependency:"PresignProvider"`
}

func (h *PresignUploadFormHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUserOrMasterKey
}

func (h *PresignUploadFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *PresignUploadFormHandler) Handle(w http.ResponseWriter, r *http.Request) (result interface{}, err error) {
	payload := handler.EmptyRequestPayload{}
	err = handler.DecodeJSONBody(r, w, &payload)
	if err != nil {
		return
	}

	u := &url.URL{
		Scheme: coreHttp.GetProto(r),
		User:   r.URL.User,
		Path:   "/_asset/upload_form",
	}
	if r.Host != "" {
		u.Host = r.Host
	} else {
		u.Host = r.URL.Host
	}

	req, _ := http.NewRequest("POST", u.String(), nil)

	h.PresignProvider.Presign(req, cloudstorage.PresignPutExpires)

	result = map[string]interface{}{
		"url": req.URL.String(),
	}

	return
}
