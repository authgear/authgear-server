package web

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/template"
)

type ResponseRenderer struct {
	TemplateEngine *template.Engine
}

func (r *ResponseRenderer) Render(w http.ResponseWriter, req *http.Request, tpl template.Resource, data interface{}) {
	r.RenderStatus(w, req, http.StatusOK, tpl, data)
}

func (r *ResponseRenderer) RenderStatus(w http.ResponseWriter, req *http.Request, status int, tpl template.Resource, data interface{}) {
	r.TemplateEngine.RenderStatus(w, req, status, tpl, data)
}

func (r *ResponseRenderer) RenderHTML(w http.ResponseWriter, req *http.Request, tpl *template.HTML, data interface{}) {
	// It is very important to specify the encoding because browsers assume ASCII if encoding is not specified.
	// No need to use FormatMediaType because the value is constant.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Render(w, req, tpl, data)
}
