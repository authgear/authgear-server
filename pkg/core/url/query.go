package url

import (
	gourl "net/url"
)

func WithQueryParamsAdded(url *gourl.URL, params map[string]string) *gourl.URL {
	q := url.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	u := *url
	u.RawQuery = q.Encode()
	return &u
}
