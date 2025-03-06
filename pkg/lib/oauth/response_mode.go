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

func WriteResponse(w http.ResponseWriter, r *http.Request, redirectURI *url.URL, responseMode string, response map[string]string) {
	if responseMode == "" {
		responseMode = "query"
	}

	switch responseMode {
	case "query":
		HTMLRedirect(w, r, urlutil.WithQueryParamsAdded(redirectURI, response).String())
	case "fragment":
		HTMLRedirect(w, r, urlutil.WithQueryParamsSetToFragment(redirectURI, response).String())
	case "cookie":
		HTMLRedirect(w, r, urlutil.WithQueryParamsAdded(redirectURI, response).String())
	case "form_post":
		FormPost(w, r, redirectURI, response)
	default:
		http.Error(w, fmt.Sprintf("oauth: invalid response_mode %s", responseMode), http.StatusBadRequest)
	}
}

func HTMLRedirect(rw http.ResponseWriter, r *http.Request, redirectURI string) {
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
