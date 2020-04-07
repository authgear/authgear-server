package template

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	htmlTemplate "html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	textTemplate "text/template"
	"unicode/utf8"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

const MaxTemplateSize = 1024 * 1024 * 1

// DownloadStringFromAssuminglyTrustedURL downloads the content of url.
// url is assumed to be trusted.
func DownloadStringFromAssuminglyTrustedURL(url string) (content string, err error) {
	// nolint: gosec
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = errors.Newf("unexpected status code: %d", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, MaxTemplateSize))
	if err != nil {
		return
	}

	if !utf8.Valid(body) {
		err = errors.New("expected content to be UTF-8 encoded")
		return
	}

	content = string(body)
	return
}

// EncodeContextToURLQueryParamValue encodes context into URL query param value.
// Specifially, the context is first encoded into JSON and then base64url encoded.
func EncodeContextToURLQueryParamValue(context map[string]interface{}) (val string, err error) {
	if context == nil {
		return
	}
	bytes, err := json.Marshal(context)
	if err != nil {
		return
	}
	val = base64.RawURLEncoding.EncodeToString(bytes)
	return
}

// DecodeURLQueryParamValueToContext is the inverse of EncodeContextToURLQueryParamValue.
func DecodeURLQueryParamValueToContext(val string) (context map[string]interface{}, err error) {
	if val == "" {
		return
	}
	bytes, err := base64.RawURLEncoding.DecodeString(val)
	if err != nil {
		return
	}
	err = json.Unmarshal(bytes, &context)
	if err != nil {
		return
	}
	return
}

func SetContextToURLQuery(u *url.URL, context map[string]interface{}) error {
	encoded, err := EncodeContextToURLQueryParamValue(context)
	if err != nil {
		return err
	}
	query := u.Query()
	query.Set("x-skygear-redirect-data", encoded)
	u.RawQuery = query.Encode()
	return nil
}

type RenderOptions struct {
	// The name of the main template
	Name string
	// The template body of the main template
	TemplateBody string
	// The additional templates to parse.
	Defines []string
	// The context for rendering the template
	Context map[string]interface{}
	// The options to Validator
	ValidatorOpts []ValidatorOption
}

func RenderTextTemplate(opts RenderOptions) (out string, err error) {
	if opts.TemplateBody == "" {
		return
	}

	validator := NewValidator(opts.ValidatorOpts...)

	// Parse the main template
	template, err := textTemplate.New(opts.Name).Parse(opts.TemplateBody)
	if err != nil {
		err = errors.Newf("failed to parse template: %w", err)
		return
	}
	// Parse defines
	for _, define := range opts.Defines {
		_, err = template.Parse(define)
		if err != nil {
			err = errors.Newf("failed to parse template: %w", err)
			return
		}
	}

	// Validate all templates
	err = validator.ValidateTextTemplate(template)
	if err != nil {
		err = errors.Newf("failed to validate template: %w", err)
		return
	}

	var buf bytes.Buffer
	if err = template.Execute(&limitedWriter{w: &buf, n: MaxTemplateSize}, opts.Context); err != nil {
		err = errors.Newf("failed to execute template: %w", err)
		return
	}

	out = string(buf.Bytes())
	return
}

func RenderHTMLTemplate(opts RenderOptions) (out string, err error) {
	if opts.TemplateBody == "" {
		return
	}

	validator := NewValidator(opts.ValidatorOpts...)

	// Parse the main template
	template, err := htmlTemplate.New(opts.Name).Parse(opts.TemplateBody)
	if err != nil {
		err = errors.Newf("failed to parse template: %w", err)
		return
	}
	// Parse defines
	for _, define := range opts.Defines {
		_, err = template.Parse(define)
		if err != nil {
			err = errors.Newf("failed to parse template: %w", err)
			return
		}
	}

	// Validate all templates
	err = validator.ValidateHTMLTemplate(template)
	if err != nil {
		err = errors.Newf("failed to validate template: %w", err)
		return
	}

	var buf bytes.Buffer
	if err = template.Execute(&limitedWriter{w: &buf, n: MaxTemplateSize}, opts.Context); err != nil {
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
