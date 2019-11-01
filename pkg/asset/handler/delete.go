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
)

func AttachDeleteHandler(
	server *server.Server,
	dependencyMap inject.DependencyMap,
) *server.Server {
	server.Handle("/delete/{asset_name}", &DeleteHandlerFactory{
		dependencyMap,
	}).Methods("OPTIONS", "DELETE")
	return server
}

type DeleteHandlerFactory struct {
	DependencyMap inject.DependencyMap
}

func (f *DeleteHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &DeleteHandler{}
	inject.DefaultRequestInject(h, f.DependencyMap, request)
	return h.RequireAuthz(h, h)
}

/*
	@Operation DELETE /delete/{asset_name} - Delete the given asset.
		Delete the given asset.

		@SecurityRequirement master_key

		@Parameter asset_name path
			Name of asset
			@JSONSchema
				{ "type": "string" }

		@Response 200
*/
type DeleteHandler struct {
	RequireAuthz         handler.RequireAuthz  `dependency:"RequireAuthz"`
	CloudStorageProvider cloudstorage.Provider `dependency:"CloudStorageProvider"`
}

func (h *DeleteHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) (result interface{}, err error) {

	vars := mux.Vars(r)
	assetName := vars["asset_name"]
	err = h.CloudStorageProvider.Delete(assetName)
	if err != nil {
		return
	}
	result = map[string]interface{}{}
	return
}
