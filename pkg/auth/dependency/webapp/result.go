package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/httputil"
)

type Result struct {
	state *State
	// redirectURI is a string because we may not attach x_sid to it sometimes.
	// For example, when it is the authorization URL to an OAuth provider.
	redirectURI      string
	errorRedirectURI *url.URL
	cookies          []*http.Cookie
}

func (r *Result) WriteResponse(w http.ResponseWriter, req *http.Request) {
	for _, cookie := range r.cookies {
		httputil.UpdateCookie(w, cookie)
	}

	if r.state.Error != nil {
		if r.errorRedirectURI != nil {
			http.Redirect(w, req, r.state.Attach(r.errorRedirectURI).String(), http.StatusFound)
			return
		}

		http.Redirect(w, req, r.state.Attach(req.URL).String(), http.StatusFound)
		return
	}

	http.Redirect(w, req, r.redirectURI, http.StatusFound)
	return
}
