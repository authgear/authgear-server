package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/httputil"
)

const DefaultRedirectURI = "/settings"

// RedirectToRedirectURI looks at `redirect_uri`.
// If it is absent, defaults to "/settings".
// redirect_uri is then resolved against r.URL
// redirect_uri must have the same origin.
// Finally a 302 response is written.
func RedirectToRedirectURI(w http.ResponseWriter, r *http.Request, trustProxy bool) {
	redirectURI, err := httputil.GetRedirectURI(r, trustProxy)
	if err != nil {
		http.Redirect(w, r, DefaultRedirectURI, http.StatusFound)
	} else {
		http.Redirect(w, r, redirectURI, http.StatusFound)
	}
}

func RedirectToPathWithX(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, MakeURLWithPathWithX(r.URL, path), http.StatusFound)
}

func RedirectToPathWithoutX(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, MakeURLWithPathWithoutX(r.URL, path), http.StatusFound)
}

func RedirectToCurrentPath(w http.ResponseWriter, r *http.Request) {
	RedirectToPathWithX(w, r, r.URL.Path)
}

func RedirectToPathWithQuery(w http.ResponseWriter, r *http.Request, path string, query url.Values) {
	http.Redirect(w, r, NewURLWithPathAndQuery(path, query), http.StatusFound)
}

func MakeURLWithPathWithX(i *url.URL, path string) string {
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
