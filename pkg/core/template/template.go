package template

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/flosch/pongo2"
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

func ParseTextTemplateFromURL(url string, context map[string]interface{}) (string, error) {
	var body string
	var err error
	if body, err = DownloadTemplateFromURL(url); err != nil {
		return "", err
	}

	return ParseTextTemplate(body, context)
}

func ParseHTMLTemplateFromURL(url string, context map[string]interface{}) (string, error) {
	var body string
	var err error
	if body, err = DownloadTemplateFromURL(url); err != nil {
		return "", err
	}

	return ParseHTMLTemplate(body, context)
}

func ParseTextTemplate(templateString string, context map[string]interface{}) (out string, err error) {
	if templateString == "" {
		return
	}

	// turn off auto html escape
	autoEscapeOffTemplate := `{%% autoescape off %%}%s{%% endautoescape %%}`
	autoEscapeOffTemplateString := fmt.Sprintf(autoEscapeOffTemplate, templateString)

	return ParseHTMLTemplate(autoEscapeOffTemplateString, context)
}

func ParseHTMLTemplate(templateString string, context map[string]interface{}) (out string, err error) {
	if templateString == "" {
		return
	}

	tset := newTemplateSet()

	t, err := tset.FromString(templateString)
	if err != nil {
		return
	}

	defer func() {
		if rerr := recover(); rerr != nil {
			out = ""
			err = rerr.(error)
		}
	}()

	var buf bytes.Buffer
	if err = t.ExecuteWriterUnbuffered(context, &limitedWriter{w: &buf, n: MaxTemplateSize}); err != nil {
		return
	}

	out = string(buf.Bytes())
	return
}

func newTemplateSet() *pongo2.TemplateSet {
	tset := pongo2.NewSet("", nullTemplateLoader{})
	tset.BanTag("include")
	tset.BanTag("import")
	tset.BanTag("extends")
	tset.BanTag("ssi")
	return tset
}

var errLimitReached = skyerr.NewError(skyerr.UnexpectedError, "rendered template is too large")

type limitedWriter struct {
	w io.Writer
	n int64
}

func (l *limitedWriter) Write(p []byte) (n int, err error) {
	if l.n-int64(len(p)) <= 0 {
		// HACK(template): pongo2 does not handle write errors correctly,
		//                 so panic to abort template rendering early.
		panic(errLimitReached)
	}

	n, err = l.w.Write(p)
	l.n -= int64(n)

	return
}

type nullTemplateLoader struct{}

func (l nullTemplateLoader) Abs(base, name string) string       { return name }
func (l nullTemplateLoader) Get(path string) (io.Reader, error) { return l, nil }
func (l nullTemplateLoader) Read(p []byte) (int, error)         { return 0, io.EOF }
