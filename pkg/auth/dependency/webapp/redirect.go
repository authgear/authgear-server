package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/httputil"
)

const DefaultRedirectURI = "/settings"

func GetRedirectURI(r *http.Request, trustProxy bool) string {
	redirectURI, err := httputil.GetRedirectURI(r, trustProxy)
	if err != nil {
		return DefaultRedirectURI
	}
	return redirectURI
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
