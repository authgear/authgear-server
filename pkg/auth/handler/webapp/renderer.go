package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/template"
)

type Renderer interface {
	// Render renders the template into response body.
	Render(w http.ResponseWriter, r *http.Request, tpl template.Resource, data interface{})
	// RenderHTML is a shorthand of Render that renders HTML.
	RenderHTML(w http.ResponseWriter, r *http.Request, tpl *template.HTML, data interface{})
	RenderStatus(w http.ResponseWriter, req *http.Request, status int, tpl template.Resource, data interface{})
}
