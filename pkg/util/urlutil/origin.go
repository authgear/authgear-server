package urlutil

import (
	"net/url"
)

func ExtractOrigin(u *url.URL) *url.URL {
	return &url.URL{
		Scheme: u.Scheme,
		Opaque: u.Opaque,
		Host:   u.Host,
	}
}
