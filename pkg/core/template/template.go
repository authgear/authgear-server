package template

import (
	"bytes"
	htmlTemplate "html/template"
	"io"
	textTemplate "text/template"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

// TODO(template): Apply MaxTemplateSize on remote template.
const MaxTemplateSize = 1024 * 1024 * 1

func RenderTextTemplate(id string, templateString string, context map[string]interface{}) (out string, err error) {
	if templateString == "" {
		return
	}

	template, err := textTemplate.New(id).Parse(templateString)
	if err != nil {
		err = errors.Newf("failed to parse template: %w", err)
		return
	}

	err = ValidateTextTemplate(template)
	if err != nil {
		err = errors.Newf("failed to validate template: %w", err)
		return
	}

	var buf bytes.Buffer
	if err = template.Execute(&limitedWriter{w: &buf, n: MaxTemplateSize}, context); err != nil {
		err = errors.Newf("failed to execute template: %w", err)
		return
	}

	out = string(buf.Bytes())
	return
}

func RenderHTMLTemplate(id string, templateString string, context map[string]interface{}) (out string, err error) {
	if templateString == "" {
		return
	}

	template, err := htmlTemplate.New(id).Parse(templateString)
	if err != nil {
		err = errors.Newf("failed to parse template: %w", err)
		return
	}

	err = ValidateHTMLTemplate(template)
	if err != nil {
		err = errors.Newf("failed to validate template: %w", err)
		return
	}

	var buf bytes.Buffer
	if err = template.Execute(&limitedWriter{w: &buf, n: MaxTemplateSize}, context); err != nil {
		err = errors.Newf("failed to execute template: %w", err)
		return
	}

	out = string(buf.Bytes())
	return
}

var errLimitReached = errors.New("rendered template is too large")

type limitedWriter struct {
	w io.Writer
	n int64
}

func (l *limitedWriter) Write(p []byte) (n int, err error) {
	if l.n-int64(len(p)) <= 0 {
		return 0, errLimitReached
	}

	n, err = l.w.Write(p)
	l.n -= int64(n)

	return
}
