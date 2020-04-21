package flows

import "net/http"

type TokenResult struct {
	Token string
}

type WebAppResult struct {
	Cookies []*http.Cookie
}
