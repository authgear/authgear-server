package webapp

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func MakeURL(u *url.URL, path string, inQuery url.Values) *url.URL {
	uu := httputil.HostRelative(u)
	uu.RawQuery = inQuery.Encode()
	if path != "" {
		uu.Path = path
	}
	return uu
}
