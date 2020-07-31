package webapp

import (
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/httputil"
)

func RemoveX(q url.Values) {
	for name := range q {
		if strings.HasPrefix(name, "x_") {
			delete(q, name)
		}
	}
}

func MakeURL(u *url.URL, path string, inQuery url.Values) *url.URL {
	uu := *u

	q := uu.Query()
	RemoveX(q)
	for name := range inQuery {
		q.Set(name, inQuery.Get(name))
	}
	uu.RawQuery = q.Encode()

	if path != "" {
		uu.Path = path
	}

	return httputil.HostRelative(&uu)
}
