package webapp

import (
	"errors"
	"net/http"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

const DefaultRedirectURI = "/settings"

// RedirectToRedirectURI looks at `redirect_uri`.
// If it is absent, defaults to "/settings".
// redirect_uri is then resolved against r.URL
// redirect_uri must have the same origin.
// Finally a 302 response is written.
func RedirectToRedirectURI(w http.ResponseWriter, r *http.Request) {
	redirectURI, err := getRedirectURI(r)
	if err != nil {
		http.Redirect(w, r, DefaultRedirectURI, http.StatusFound)
	} else {
		http.Redirect(w, r, redirectURI, http.StatusFound)
	}
}

func getRedirectURI(r *http.Request) (out string, err error) {
	out = r.URL.Query().Get("redirect_uri")
	if out == "" {
		err = errors.New("not found")
		return
	}

	u, err := r.URL.Parse(out)
	if err != nil {
		return
	}

	recursive := u.Path == r.URL.Path || (u.RawPath != "" && u.RawPath == r.URL.RawPath)
	sameOrigin := (u.Scheme == "" && u.Host == "") || u.Scheme == corehttp.GetProto(r) && u.Host == corehttp.GetHost(r)

	if !sameOrigin {
		err = errors.New("not the same origin")
		return
	}

	if recursive {
		err = errors.New("recursive")
		return
	}

	out = u.String()
	return
}
