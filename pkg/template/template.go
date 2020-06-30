package template

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	htmlTemplate "html/template"
	"io"
	"net/url"
	textTemplate "text/template"

	"github.com/authgear/authgear-server/pkg/core/errors"
)

const MaxTemplateSize = 1024 * 1024 * 1

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
	// Funcs injects custom functions
	Funcs map[string]interface{}
}

func RenderTextTemplate(opts RenderOptions) (out string, err error) {
	if opts.TemplateBody == "" {
		return
	}

	// Initialize the template object
	template := textTemplate.New(opts.Name)

	// Inject the funcs map before parsing any templates.
	// This is required by the documentation.
	if opts.Funcs != nil {
		template.Funcs(opts.Funcs)
	}

	// Parse the main template
	_, err = template.Parse(opts.TemplateBody)
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
	validator := NewValidator(opts.ValidatorOpts...)
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

	// Initialize the template object
	template := htmlTemplate.New(opts.Name)

	// Inject the funcs map before parsing any templates.
	// This is required by the documentation.
	if opts.Funcs != nil {
		template.Funcs(opts.Funcs)
	}

	// Parse the main template
	_, err = template.Parse(opts.TemplateBody)
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
	validator := NewValidator(opts.ValidatorOpts...)
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
