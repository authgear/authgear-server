package template

import (
	"bytes"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"io"
	textTemplate "text/template"
)

const MaxTemplateSize = 1024 * 1024 * 1

type RenderOptions struct {
	// The name of the main template
	Name string
	// The template body of the main template
	TemplateBody string
	// The additional templates to parse.
	Defines []string
	// The data for rendering the template
	Data interface{}
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
		err = fmt.Errorf("failed to parse template: %w", err)
		return
	}
	// Parse defines
	for _, define := range opts.Defines {
		_, err = template.Parse(define)
		if err != nil {
			err = fmt.Errorf("failed to parse template: %w", err)
			return
		}
	}

	// Validate all templates
	validator := NewValidator(opts.ValidatorOpts...)
	err = validator.ValidateTextTemplate(template)
	if err != nil {
		err = fmt.Errorf("failed to validate template: %w", err)
		return
	}

	var buf bytes.Buffer
	if err = template.Execute(&limitedWriter{w: &buf, n: MaxTemplateSize}, opts.Data); err != nil {
		err = fmt.Errorf("failed to execute template: %w", err)
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
		err = fmt.Errorf("failed to parse template: %w", err)
		return
	}
	// Parse defines
	for _, define := range opts.Defines {
		_, err = template.Parse(define)
		if err != nil {
			err = fmt.Errorf("failed to parse template: %w", err)
			return
		}
	}

	// Validate all templates
	validator := NewValidator(opts.ValidatorOpts...)
	err = validator.ValidateHTMLTemplate(template)
	if err != nil {
		err = fmt.Errorf("failed to validate template: %w", err)
		return
	}

	var buf bytes.Buffer
	if err = template.Execute(&limitedWriter{w: &buf, n: MaxTemplateSize}, opts.Data); err != nil {
		err = fmt.Errorf("failed to execute template: %w", err)
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
