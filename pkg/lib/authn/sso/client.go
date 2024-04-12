package sso

import "net/http"

type OAuthHTTPClient struct {
	*http.Client
}
