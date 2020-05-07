package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

func RedirectToPathWithQueryPreserved(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, MakeURLWithPathWithQueryPreserved(r.URL, path), http.StatusFound)
}

func RedirectToCurrentPath(w http.ResponseWriter, r *http.Request) {
	RedirectToPathWithQueryPreserved(w, r, r.URL.Path)
}

func RedirectToPathWithQuery(w http.ResponseWriter, r *http.Request, path string, query url.Values) {
	http.Redirect(w, r, NewURLWithPathAndQuery(path, query), http.StatusFound)
}

func getRedirectURI(r *http.Request) (out string, err error) {
	out = r.URL.Query().Get("redirect_uri")
	if out == "" {
		err = errors.New("not found")
		return
	}

	out, err = parseRedirectURI(r, out)
	return
}

func parseRedirectURI(r *http.Request, redirectURL string) (out string, err error) {
	u, err := r.URL.Parse(redirectURL)
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

func MakeURLWithPathWithQueryPreserved(i *url.URL, path string) string {
	u := *i
	u.Path = path
	u.Scheme = ""
	u.Opaque = ""
	u.Host = ""
	u.User = nil
	return u.String()
}

func MakeURLWithPathWithoutX(i *url.URL, path string) string {
	u := *i

	u.Path = path
	u.Scheme = ""
	u.Opaque = ""
	u.Host = ""
	u.User = nil

	q := u.Query()
	for name := range q {
		if strings.HasPrefix(name, "x_") {
			delete(q, name)
		}
	}
	u.RawQuery = q.Encode()

	return u.String()
}

func MakeURLWithQuery(u *url.URL, query url.Values) string {
	q := u.Query()
	for name := range query {
		q.Set(name, query.Get(name))
	}
	return fmt.Sprintf("?%s", q.Encode())
}

func NewURLWithPathAndQuery(path string, query url.Values) string {
	u := url.URL{}
	u.Path = path
	q := u.Query()
	for name := range query {
		q.Set(name, query.Get(name))
	}
	u.RawQuery = q.Encode()
	return u.String()
}
