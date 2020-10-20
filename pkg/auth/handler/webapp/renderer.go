package webapp

import (
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type Renderer interface {
	// Render renders the template into response body.
	// Content-Length is set before calling beforeWrite.
	Render(w http.ResponseWriter, r *http.Request, tpl template.Resource, data interface{}, beforeWrite func(w http.ResponseWriter))
	// RenderHTML is a shorthand of Render that renders HTML.
	RenderHTML(w http.ResponseWriter, r *http.Request, tpl *template.HTML, data interface{})
}

type ResponseRendererLogger struct{ *log.Logger }

func NewResponseRendererLogger(lf *log.Factory) ResponseRendererLogger {
	return ResponseRendererLogger{lf.New("renderer")}
}

type ResponseRenderer struct {
	TemplateEngine *template.Engine
	Logger         ResponseRendererLogger
}

func (r *ResponseRenderer) Render(w http.ResponseWriter, req *http.Request, tpl template.Resource, data interface{}, beforeWrite func(w http.ResponseWriter)) {
	r.Logger.WithFields(map[string]interface{}{
		"data": data,
	}).Debug("render with data")

	preferredLanguageTags := intl.GetPreferredLanguageTags(req.Context())
	out, err := r.TemplateEngine.Render(
		tpl,
		preferredLanguageTags,
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
	_, err = w.Write(body)
	if err != nil {
		panic(err)
	}
}

func (r *ResponseRenderer) RenderHTML(w http.ResponseWriter, req *http.Request, tpl *template.HTML, data interface{}) {
	r.Render(w, req, tpl, data, func(w http.ResponseWriter) {
		// It is very important to specify the encoding because browsers assume ASCII if encoding is not specified.
		// No need to use FormatMediaType because the value is constant.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	})
}
