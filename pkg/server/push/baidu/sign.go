package baidu

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/url"
	"sort"
)

// mockable in test
var signRequest = signRequestFunc

// signRequestFunc derives the md5 signature used in making requests.
// See http://push.baidu.com/doc/restapi/sdk_developer#%E7%AD%BE%E5%90%8D%E7%AE%97%E6%B3%95
// for details of the algorithm.
func signRequestFunc(method string, urlString string, values url.Values, secretKey string) string {
	baseString := method + urlString + makeQueryString(values) + secretKey
	encoded := url.QueryEscape(baseString)

	h := md5.New()
	if _, err := io.WriteString(h, encoded); err != nil {
		// md5's Hash.Write never returns error
		panic(err)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func makeQueryString(v url.Values) string {
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		prefix := k + "="
		for _, v := range vs {
			buf.WriteString(prefix)
			buf.WriteString(v)
		}
	}

	return buf.String()
}
