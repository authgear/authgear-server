package urlutil

import (
	"net/url"
)

func ExtractOrigin(u *url.URL) *url.URL {
	newU := *u
	newU.Path = ""
	newU.RawQuery = ""
	newU.Fragment = ""
	return &newU
}
