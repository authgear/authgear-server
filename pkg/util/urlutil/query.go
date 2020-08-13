package urlutil

import (
	"net/url"
)

func WithQueryParamsAdded(url *url.URL, params map[string]string) *url.URL {
	q := url.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	u := *url
	u.RawQuery = q.Encode()
	return &u
}
