package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	coreurl "github.com/skygeario/skygear-server/pkg/core/url"
)

const AuthorizationResultHTML = `<!DOCTYPE html>
<head>
<meta http-equiv="refresh" content="0;url={{ .redirect_uri }}" />
</head>
<script>
window.location.href = "{{ .redirect_uri }}"
</script>`

type AuthorizationResult interface {
	WriteResponse(rw http.ResponseWriter, r *http.Request)
	IsInternalError() bool
}

type (
	authorizationResultRedirect struct {
		RedirectURI *url.URL
		Response    protocol.AuthorizationResponse
	}
	authorizationResultError struct {
		RedirectURI   *url.URL
		InternalError bool
		Response      protocol.ErrorResponse
	}
	authorizationResultRequireAuthn struct {
		AuthenticateURI *url.URL
	}
)

func (a authorizationResultRedirect) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	redirectURI := coreurl.WithQueryParamsAdded(a.RedirectURI, a.Response)
	redirect(rw, redirectURI.String())
}

func (a authorizationResultRedirect) IsInternalError() bool {
	return false
}

func (a authorizationResultError) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	if a.RedirectURI != nil {
		redirectURI := coreurl.WithQueryParamsAdded(a.RedirectURI, a.Response)
		redirect(rw, redirectURI.String())
	} else {
		err := "Invalid OAuth authorization request:\n"
		keys := make([]string, 0, len(a.Response))
		for k := range a.Response {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i != 0 {
				err += "\n"
			}
			err += fmt.Sprintf("%s: %s", k, a.Response[k])
		}
		http.Error(rw, err, http.StatusBadRequest)
	}
}

func (a authorizationResultError) IsInternalError() bool {
	return a.InternalError
}

func (a authorizationResultRequireAuthn) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, a.AuthenticateURI.String(), http.StatusFound)
}

func (a authorizationResultRequireAuthn) IsInternalError() bool {
	return false
}

func redirect(rw http.ResponseWriter, redirectURI string) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	// NOTE(authui): XHR and redirect
	// Normally we should use HTTP 302 to redirect.
	// However, when XHR is used, redirect is followed automatically.
	// The final redirect URI may be custom URI which is considered unsecure by user agent.
	// Therefore, we write HTML and use <meta http-equiv="refresh"> to redirect.
	// rw.Header().Set("Location", redirectURI)
	// rw.WriteHeader(http.StatusFound)

	tmpl, err := template.New("authorization_result").Parse(AuthorizationResultHTML)
	if err != nil {
		panic("oauth: invalid authorization result page template")
	}
	err = tmpl.Execute(rw, map[string]string{
		"redirect_uri": redirectURI,
	})
	if err != nil {
		panic("oauth: failed to load authorization result page")
	}
}
