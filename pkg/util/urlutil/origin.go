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

func ApplyOriginToURL(origin *url.URL, u *url.URL) *url.URL {
	out := *u
	out.Scheme = origin.Scheme
	out.Opaque = origin.Opaque
	out.Host = origin.Host
	return &out
}
