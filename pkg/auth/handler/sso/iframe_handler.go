package sso

import (
	"io"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachIFrameHandlerFactory(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/iframe_handler", &IFrameHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "GET")
	return server
}

type IFrameHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f IFrameHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &IFrameHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h
}

func (f IFrameHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf()
}

type IFrameHandler struct {
	IFrameHTMLProvider sso.IFrameHTMLProvider `dependency:"IFrameHTMLProvider"`
}

// ServeHTTP provides html of iframe handler that are used in js sdk for
// retrieving oauth result
// curl -X GET http://localhost:3000/sso/iframe_handler
//
func (h IFrameHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	html, _ := h.IFrameHTMLProvider.HTML()
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(rw, html)
	return
}
