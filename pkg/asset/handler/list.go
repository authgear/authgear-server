package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
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
		response.Err = skyerr.MakeError(err)
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
