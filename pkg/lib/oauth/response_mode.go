package oauth

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

const htmlRedirectTemplateString = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="refresh" content="0;url={{ .redirect_uri }}" />
</head>
<body>
{{- if $.CSPNonce }}
<script nonce="{{ $.CSPNonce }}">
{{- else }}
<script>
{{- end }}
window.location.href = "{{ .redirect_uri }}"
</script>
</body>
</html>
`

const formPostTemplateString = `<!DOCTYPE html>
<html>
<head>
<title>Submit this form</title>
</head>
<body>
<noscript>Please submit this form to proceed</noscript>
<form method="post" action="{{ .redirect_uri }}">
{{- range $name, $value := .response }}
<input type="hidden" name="{{ $name }}" value="{{ $value }}">
{{- end }}
<button type="submit" name="" value="">Submit</button>
</form>
{{- if $.CSPNonce }}
<script nonce="{{ $.CSPNonce }}">
{{- else }}
<script>
{{- end }}
document.forms[0].submit();
</script>
</body>
</html>
`

var htmlRedirectTemplate *template.Template
var formPostTemplate *template.Template

func init() {
	var err error

	htmlRedirectTemplate, err = template.New("html_redirect").Parse(htmlRedirectTemplateString)
	if err != nil {
		panic(fmt.Errorf("oauth: invalid html_redirect template: %w", err))
	}

	formPostTemplate, err = template.New("form_post").Parse(formPostTemplateString)
	if err != nil {
		panic(fmt.Errorf("oauth: invalid form_post template: %w", err))
	}
}

type WriteResponseOptions struct {
	RedirectURI  *url.URL
	ResponseMode string
	UseHTTP200   bool
	Response     map[string]string
}

func WriteResponse(w http.ResponseWriter, r *http.Request, options WriteResponseOptions) {
	responseMode := options.ResponseMode
	if responseMode == "" {
		responseMode = "query"
	}

	useHTTP200 := options.UseHTTP200

	switch responseMode {
	case "query":
		switch useHTTP200 {
		case true:
			HTTP200HTMLRedirect(w, r, urlutil.WithQueryParamsAdded(options.RedirectURI, options.Response).String())
		default:
			HTTP303HTMLRedirect(w, r, urlutil.WithQueryParamsAdded(options.RedirectURI, options.Response).String())
		}
	case "fragment":
		switch useHTTP200 {
		case true:
			HTTP200HTMLRedirect(w, r, urlutil.WithQueryParamsSetToFragment(options.RedirectURI, options.Response).String())
		default:
			HTTP303HTMLRedirect(w, r, urlutil.WithQueryParamsSetToFragment(options.RedirectURI, options.Response).String())
		}
	case "cookie":
		switch useHTTP200 {
		case true:
			HTTP200HTMLRedirect(w, r, urlutil.WithQueryParamsAdded(options.RedirectURI, options.Response).String())
		default:
			HTTP303HTMLRedirect(w, r, urlutil.WithQueryParamsAdded(options.RedirectURI, options.Response).String())
		}
	case "form_post":
		FormPost(w, r, options.RedirectURI, options.Response)
	default:
		http.Error(w, fmt.Sprintf("oauth: invalid response_mode %s", responseMode), http.StatusBadRequest)
	}
}

func HTTP200HTMLRedirect(rw http.ResponseWriter, r *http.Request, redirectURI string) {
	// This redirect approach is kept for backward compatibility with custom UI is in use.
	// See https://linear.app/authgear/issue/DEV-2544/revisit-oauth-redirect-approach
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	err := htmlRedirectTemplate.Execute(rw, map[string]string{
		"CSPNonce":     httputil.GetCSPNonce(r.Context()),
		"redirect_uri": redirectURI,
	})
	if err != nil {
		panic(fmt.Errorf("oauth: failed to execute html_redirect template: %w", err))
	}
}

func HTTP303HTMLRedirect(rw http.ResponseWriter, r *http.Request, redirectURI string) {
	// About this redirect approach.
	// We use a combination of redirects in this approach.
	//
	// 1. HTTP 303 with Location header. The use of 303 forces the use of GET method.
	//    This redirect is preferred when the browser respects it.
	// 2. <meta http-equiv="refresh">. In case 1 fails, this does the redirect.
	// 3. window.location.href. In case 2 fails, this does the redirect.
	//
	// Using iframe is also supported, see PROJECT_ROOT/experiments/DEV-2544
	// When an iframe is used to load the response, the iframe must have allow-top-navigation set.
	// Then the window.location.href will navigate the top-level frame.

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Header().Set("Location", redirectURI)
	// 303, not 302.
	rw.WriteHeader(http.StatusSeeOther)

	err := htmlRedirectTemplate.Execute(rw, map[string]string{
		"CSPNonce":     httputil.GetCSPNonce(r.Context()),
		"redirect_uri": redirectURI,
	})
	if err != nil {
		panic(fmt.Errorf("oauth: failed to execute html_redirect template: %w", err))
	}
}

func FormPost(w http.ResponseWriter, r *http.Request, redirectURI *url.URL, response map[string]string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := formPostTemplate.Execute(w, map[string]interface{}{
		"CSPNonce":     httputil.GetCSPNonce(r.Context()),
		"redirect_uri": redirectURI.String(),
		"response":     response,
	})
	if err != nil {
		panic(fmt.Errorf("oauth: failed to execute form_post template: %w", err))
	}
}
