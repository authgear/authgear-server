package template

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	textTemplate "text/template"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

const MaxTemplateSize = 1024 * 1024 * 1

func DownloadTemplateFromFilePath(filePath string) (string, error) {
	filePath = filepath.Clean(filePath)

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(io.LimitReader(f, MaxTemplateSize))
	if err != nil {
		return "", err
	}

	return string(b), nil
}

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
		return "", fmt.Errorf("unsuccessful request: %s", resp.Status)
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
		return
	}

	err = ValidateTextTemplate(template)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	if err = template.Execute(&limitedWriter{w: &buf, n: MaxTemplateSize}, context); err != nil {
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
		return
	}

	err = ValidateHTMLTemplate(template)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	if err = template.Execute(&limitedWriter{w: &buf, n: MaxTemplateSize}, context); err != nil {
		return
	}

	out = string(buf.Bytes())
	return
}

var errLimitReached = skyerr.NewError(skyerr.UnexpectedError, "rendered template is too large")

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
