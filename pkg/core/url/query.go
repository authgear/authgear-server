package url

import (
	gourl "net/url"
	"strings"
)

// QueryParam is an entry of application/x-www-form-urlencoded.
type QueryParam struct {
	Key   string
	Value string
}

// Query is application/x-www-form-urlencoded.
// Unlike url.Values, this preserves insertion order of the pairs.
type Query struct {
	Params []QueryParam
}

// Add adds the given key and value.
func (q *Query) Add(key, value string) {
	q.Params = append(q.Params, QueryParam{
		Key:   key,
		Value: value,
	})
}

// Encode returns the encoded form.
func (q *Query) Encode() string {
	if len(q.Params) <= 0 {
		return ""
	}
	var buf strings.Builder
	for _, param := range q.Params {
		key := gourl.QueryEscape(param.Key)
		value := gourl.QueryEscape(param.Value)
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(value)
	}
	return buf.String()
}

func WithQueryParamsAdded(url *gourl.URL, params map[string]string) *gourl.URL {
	q := url.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	u := *url
	u.RawQuery = q.Encode()
	return &u
}
