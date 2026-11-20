package urlutil

import (
	"net/url"
)

func WithQueryParamsSetToFragment(u *url.URL, params map[string]string) *url.URL {
	q := url.Values{}
	for name, value := range params {
		q.Set(name, value)
	}
	newU := *u
	// Fragment holds the decoded form, so the encoded params go in RawFragment
	// to avoid percent-encoding them a second time.
	newU.RawFragment = q.Encode()
	newU.Fragment, _ = url.PathUnescape(newU.RawFragment)
	return &newU
}
