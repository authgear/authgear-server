package handler

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/asset/dependency/presign"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachPresignUploadFormHandler(
	server *server.Server,
	dependencyMap inject.DependencyMap,
) *server.Server {
	server.Handle("/presign_upload_form", &PresignUploadFormHandlerFactory{
		dependencyMap,
	}).Methods("OPTIONS", "POST")
	return server
}

type PresignUploadFormHandlerFactory struct {
	DependencyMap inject.DependencyMap
}

func (f *PresignUploadFormHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &PresignUploadFormHandler{}
	inject.DefaultRequestInject(h, f.DependencyMap, request)
	return h.RequireAuthz(h, h)
}

type PresignUploadFormHandler struct {
	RequireAuthz    handler.RequireAuthz `dependency:"RequireAuthz"`
	PresignProvider presign.Provider     `dependency:"PresignProvider"`
}

func (h *PresignUploadFormHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AnyOf(
		authz.PolicyFunc(policy.RequireMasterKey),
		policy.AllOf(
			authz.PolicyFunc(policy.DenyNoAccessKey),
			authz.PolicyFunc(policy.RequireAuthenticated),
			authz.PolicyFunc(policy.DenyDisabledUser),
		),
	)
}

func (h *PresignUploadFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Err = skyerr.MakeError(err)
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

	h.PresignProvider.Presign(req)

	result = map[string]interface{}{
		"url": req.URL.String(),
	}

	return
}
