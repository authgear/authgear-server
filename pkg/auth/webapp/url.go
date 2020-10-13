package webapp

import (
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func RemoveX(q url.Values) {
	for name := range q {
		if strings.HasPrefix(name, "x_") {
			delete(q, name)
		}
	}
}

func MakeURL(u *url.URL, path string, inQuery url.Values) *url.URL {
	uu := httputil.HostRelative(u)
	uu.RawQuery = inQuery.Encode()
	if path != "" {
		uu.Path = path
	}
	return uu
}
