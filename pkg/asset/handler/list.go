package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachListHandler(
	server *server.Server,
	dependencyMap inject.DependencyMap,
) *server.Server {
	server.Handle("/assets", &ListHandlerFactory{
		dependencyMap,
	}).Methods("OPTIONS", "GET")
	return server
}

type ListHandlerFactory struct {
	DependencyMap inject.DependencyMap
}

func (f *ListHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ListHandler{}
	inject.DefaultRequestInject(h, f.DependencyMap, request)
	return h.RequireAuthz(h, h)
}

// @JSONSchema
const ListAssetResponseSchema = `
{
	"$id": "#ListAssetResponse",
	"type": "object",
	"properties": {
		"result": {
			"type": "object",
			"properties": {
				"pagination_token": { "type": "string" },
				"assets": {
					"type": "array",
					"items": { "$ref": "#ListAssetItem" }
				}
			},
			"required": ["assets"]
		}
	}
}
`

// @JSONSchema
const ListAssetItemSchema = `
{
	"$id": "#ListAssetItem",
	"type": "object",
	"properties": {
		"asset_name": { "type": "string" },
		"size": { "type": "integer" }
	},
	"required": ["asset_name", "size"]
}
`

// nolint: deadcode
/*
	@ID ListAssetPaginationToken
	@Parameter pagination_token query
		The opaque pagination token.
		@JSONSchema
			{ "type": "string" }
*/
type listAssetPaginationToken string

// nolint: deadcode
/*
	@ID ListAssetPrefix
	@Parameter prefix query
		List on asset with the given prefix.
		@JSONSchema
			{ "type": "string" }
*/
type listAssetPrefix string

/*
	@Operation GET /assets - List assets.
		List assets.

		@SecurityRequirement master_key

		@Parameter {ListAssetPaginationToken}
		@Parameter {ListAssetPrefix}

		@Response 200
			@JSONSchema {ListAssetResponse}
*/
type ListHandler struct {
	RequireAuthz         handler.RequireAuthz  `dependency:"RequireAuthz"`
	CloudStorageProvider cloudstorage.Provider `dependency:"CloudStorageProvider"`
}

func (h *ListHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) (result interface{}, err error) {
	var payload cloudstorage.ListObjectsRequest
	q := r.URL.Query()
	payload.Prefix = q.Get("prefix")
	payload.PaginationToken = q.Get("pagination_token")

	resp, err := h.CloudStorageProvider.List(&payload)
	if err != nil {
		return
	}

	result = resp
	return
}
