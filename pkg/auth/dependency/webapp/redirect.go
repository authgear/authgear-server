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
	http.Redirect(w, r, MakeURLWithPath(r.URL, path), http.StatusFound)
}

func RedirectToCurrentPath(w http.ResponseWriter, r *http.Request) {
	RedirectToPathWithQueryPreserved(w, r, r.URL.Path)
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

// MakeURLWithPath generates a relative URL with path and query only.
// The query is preserved.
// If the length of the common prefix is shorter than 2, then x_* query is removed.
// This behavior enables a automatic state cleanup mechanism
// For example, /login shares state with /login/password because the two paths
// together represent the login flow.
// When the user navigates from /login to /signup, all state is cleaned up.
// Query that does not start with x_ e.g. redirect_uri is always preserved.
func MakeURLWithPath(i *url.URL, path string) string {
	u := *i

	prefix := ""
	if strings.HasPrefix(path, u.Path) && len(u.Path) > len(prefix) {
		prefix = u.Path
	}
	if strings.HasPrefix(u.Path, path) && len(path) > len(prefix) {
		prefix = path
	}

	if len(prefix) < 2 {
		q := u.Query()
		for name := range q {
			if strings.HasPrefix(name, "x_") {
				delete(q, name)
			}
		}
		u.RawQuery = q.Encode()
	}

	u.Path = path
	u.Scheme = ""
	u.Opaque = ""
	u.Host = ""
	u.User = nil
	return u.String()
}

func MakeURLWithQuery(u *url.URL, query url.Values) string {
	q := u.Query()
	for name := range query {
		q.Set(name, query.Get(name))
	}
	return fmt.Sprintf("?%s", q.Encode())
}
