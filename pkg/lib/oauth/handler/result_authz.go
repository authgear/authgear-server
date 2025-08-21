package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type (
	authorizationResultCode struct {
		RedirectURI *url.URL

		ResponseMode string
		UseHTTP200   bool

		Response protocol.AuthorizationResponse
		Cookies  []*http.Cookie
	}
	AuthorizationResultError struct {
		RedirectURI *url.URL

		ResponseMode string
		UseHTTP200   bool

		InternalError bool
		Response      protocol.ErrorResponse
		Cookies       []*http.Cookie
	}
)

func (a authorizationResultCode) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	for _, cookie := range a.Cookies {
		httputil.UpdateCookie(rw, cookie)
	}
	writeResponseOptions := oauth.WriteResponseOptions{
		RedirectURI:  a.RedirectURI,
		ResponseMode: a.ResponseMode,
		UseHTTP200:   a.UseHTTP200,
		Response:     a.Response,
	}
	oauth.WriteResponse(rw, r, writeResponseOptions)
}

func (a authorizationResultCode) IsInternalError() bool {
	return false
}

func (a AuthorizationResultError) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	for _, cookie := range a.Cookies {
		httputil.UpdateCookie(rw, cookie)
	}
	if a.RedirectURI != nil {
		writeResponseOptions := oauth.WriteResponseOptions{
			RedirectURI:  a.RedirectURI,
			ResponseMode: a.ResponseMode,
			UseHTTP200:   a.UseHTTP200,
			Response:     a.Response,
		}

		oauth.WriteResponse(rw, r, writeResponseOptions)
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

func (a AuthorizationResultError) IsInternalError() bool {
	return a.InternalError
}

type AuthorizationParams struct {
	Context     context.Context
	Client      *config.OAuthClientConfig
	RedirectURI *url.URL
}
