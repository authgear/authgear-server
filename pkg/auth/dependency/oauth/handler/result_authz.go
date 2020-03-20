package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
)

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
		AuthorizeURI    *url.URL
		Request         protocol.AuthorizationRequest
	}
)

func (a authorizationResultRedirect) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	query := a.RedirectURI.Query()
	for k, v := range a.Response {
		query.Set(k, v)
	}
	a.RedirectURI.RawQuery = query.Encode()
	http.Redirect(rw, r, a.RedirectURI.String(), http.StatusFound)
}

func (a authorizationResultRedirect) IsInternalError() bool {
	return false
}

func (a authorizationResultError) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	if a.RedirectURI != nil {
		query := a.RedirectURI.Query()
		for k, v := range a.Response {
			query.Add(k, v)
		}
		a.RedirectURI.RawQuery = query.Encode()
		http.Redirect(rw, r, a.RedirectURI.String(), http.StatusFound)
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
	authzQuery := a.AuthorizeURI.Query()
	for k, v := range a.Request {
		authzQuery.Add(k, v)
	}
	a.AuthorizeURI.RawQuery = authzQuery.Encode()

	authnQuery := a.AuthenticateURI.Query()
	authnQuery.Add("redirect_uri", a.AuthorizeURI.String())
	a.AuthenticateURI.RawQuery = authnQuery.Encode()

	http.Redirect(rw, r, a.AuthenticateURI.String(), http.StatusFound)
}

func (a authorizationResultRequireAuthn) IsInternalError() bool {
	return false
}
