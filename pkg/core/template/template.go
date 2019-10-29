package template

import (
	"bytes"
	htmlTemplate "html/template"
	"io"
	"io/ioutil"
	"net/http"
	textTemplate "text/template"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

const MaxTemplateSize = 1024 * 1024 * 1

func DownloadTemplateFromURL(url string) (string, error) {
	// FIXME(sec): validate URL to be trusted URL
	// nolint: gosec
	resp, err := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return "", errors.Newf("failed to request: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, MaxTemplateSize))
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ParseTextTemplate(id string, templateString string, context map[string]interface{}) (out string, err error) {
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

func ParseHTMLTemplate(id string, templateString string, context map[string]interface{}) (out string, err error) {
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
