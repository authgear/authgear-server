package webapp

import (
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/intl"
	"github.com/authgear/authgear-server/pkg/template"
)

type Renderer interface {
	Render(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType, data interface{})
}

type HTMLRenderer struct {
	TemplateEngine *template.Engine
}

func (r *HTMLRenderer) Render(w http.ResponseWriter, req *http.Request, templateType config.TemplateItemType, data interface{}) {
	preferredLanguageTags := intl.GetPreferredLanguageTags(req.Context())
	out, err := r.TemplateEngine.WithValidatorOptions(
		template.AllowRangeNode(true),
		template.AllowTemplateNode(true),
		template.AllowDeclaration(true),
		template.MaxDepth(15),
	).WithPreferredLanguageTags(preferredLanguageTags).RenderTemplate(
		templateType,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		panic(err)
	}

	body := []byte(out)
	// It is very important to specify the encoding
	// because browsers assume ASCII if encoding is not specified.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Write(body)
}
