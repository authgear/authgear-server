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
	newU.Fragment = q.Encode()
	return &newU
}
