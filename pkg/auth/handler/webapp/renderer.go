package webapp

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/intl"
	"github.com/authgear/authgear-server/pkg/log"
	"github.com/authgear/authgear-server/pkg/template"
)

type Renderer interface {
	Render(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType, data interface{})

	RenderAttachment(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType, data interface{}, filename string)
}

type HTMLRendererLogger struct{ *log.Logger }

func NewHTMLRendererLogger(lf *log.Factory) HTMLRendererLogger {
	return HTMLRendererLogger{lf.New("renderer")}
}

type HTMLRenderer struct {
	TemplateEngine *template.Engine
	Logger         HTMLRendererLogger
}

func (r *HTMLRenderer) Render(w http.ResponseWriter, req *http.Request, templateType config.TemplateItemType, data interface{}) {
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
	// It is very important to specify the encoding
	// because browsers assume ASCII if encoding is not specified.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Write(body)
}

func (r *HTMLRenderer) RenderAttachment(w http.ResponseWriter, req *http.Request, templateType config.TemplateItemType, data interface{}, filename string) {
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Write(body)
}
