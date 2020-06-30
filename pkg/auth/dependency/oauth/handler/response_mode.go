package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	coreurl "github.com/authgear/authgear-server/pkg/core/url"
)

const htmlRedirectTemplateString = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="refresh" content="0;url={{ .redirect_uri }}" />
</head>
<body>
<script>
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
<script>
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

func writeResponse(w http.ResponseWriter, r *http.Request, redirectURI *url.URL, responseMode string, response map[string]string) {
	if responseMode == "" {
		responseMode = "query"
	}

	switch responseMode {
	case "query":
		htmlRedirect(w, coreurl.WithQueryParamsAdded(redirectURI, response).String())
	case "fragment":
		htmlRedirect(w, coreurl.WithQueryParamsSetToFragment(redirectURI, response).String())
	case "form_post":
		formPost(w, redirectURI, response)
	default:
		http.Error(w, fmt.Sprintf("oauth: invalid response_mode %s", responseMode), http.StatusBadRequest)
	}
}

func htmlRedirect(rw http.ResponseWriter, redirectURI string) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	// NOTE(authui): XHR and redirect
	// Normally we should use HTTP 302 to redirect.
	// However, when XHR is used, redirect is followed automatically.
	// The final redirect URI may be custom URI which is considered unsecure by user agent.
	// Therefore, we write HTML and use <meta http-equiv="refresh"> to redirect.
	// rw.Header().Set("Location", redirectURI)
	// rw.WriteHeader(http.StatusFound)
	err := htmlRedirectTemplate.Execute(rw, map[string]string{
		"redirect_uri": redirectURI,
	})
	if err != nil {
		panic(fmt.Errorf("oauth: failed to execute html_redirect template: %w", err))
	}
}

func formPost(w http.ResponseWriter, redirectURI *url.URL, response map[string]string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := formPostTemplate.Execute(w, map[string]interface{}{
		"redirect_uri": redirectURI.String(),
		"response":     response,
	})
	if err != nil {
		panic(fmt.Errorf("oauth: failed to execute form_post template: %w", err))
	}
}
