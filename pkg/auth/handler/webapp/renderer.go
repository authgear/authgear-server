package webapp

import (
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Renderer interface {
	// Render renders the template into response body.
	// Content-Length is set before calling beforeWrite.
	Render(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType, data interface{}, beforeWrite func(w http.ResponseWriter))
	// RenderHTML is a shorthand of Render that renders HTML.
	RenderHTML(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType, data interface{})
}

type ResponseRendererLogger struct{ *log.Logger }

func NewResponseRendererLogger(lf *log.Factory) ResponseRendererLogger {
	return ResponseRendererLogger{lf.New("renderer")}
}

type ResponseRenderer struct {
	TemplateEngine *template.Engine
	Logger         ResponseRendererLogger
}

func (r *ResponseRenderer) Render(w http.ResponseWriter, req *http.Request, templateType config.TemplateItemType, data interface{}, beforeWrite func(w http.ResponseWriter)) {
	r.Logger.WithFields(map[string]interface{}{
		"data": data,
	}).Debug("render with data")

	preferredLanguageTags := intl.GetPreferredLanguageTags(req.Context())
	out, err := r.TemplateEngine.WithValidatorOptions(
		template.AllowRangeNode(true),
		template.AllowTemplateNode(true),
		template.AllowDeclaration(true),
		template.MaxDepth(15),
	).WithPreferredLanguageTags(preferredLanguageTags).RenderTemplate(
		templateType,
		data,
	)
	if err != nil {
		panic(err)
	}

	body := []byte(out)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	if beforeWrite != nil {
		beforeWrite(w)
	}
	w.Write(body)
}

func (r *ResponseRenderer) RenderHTML(w http.ResponseWriter, req *http.Request, templateType config.TemplateItemType, data interface{}) {
	r.Render(w, req, templateType, data, func(w http.ResponseWriter) {
		// It is very important to specify the encoding because browsers assume ASCII if encoding is not specified.
		// No need to use FormatMediaType because the value is constant.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	})
}
