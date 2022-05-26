package web

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type dynamicCSPContextKeyType struct{}

var dynamicCSPContextKey = dynamicCSPContextKeyType{}

func WithCSPNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, dynamicCSPContextKey, nonce)
}

func GetCSPNonce(ctx context.Context) string {
	nonce, _ := ctx.Value(dynamicCSPContextKey).(string)
	return nonce
}

func CSPJoin(directives []string) string {
	return strings.Join(directives, "; ")
}

func CSPDirectives(publicOrigin string, nonce string) ([]string, error) {
	u, err := url.Parse(publicOrigin)
	if err != nil {
		return nil, err
	}

	return []string{
		"default-src 'self'",
		fmt.Sprintf("script-src 'self' 'nonce-%s' www.googletagmanager.com", nonce),
		"frame-src 'self' www.googletagmanager.com",
		"font-src 'self' cdnjs.cloudflare.com static2.sharepointonline.com fonts.googleapis.com fonts.gstatic.com",
		"style-src 'self' 'unsafe-inline' cdnjs.cloudflare.com www.googletagmanager.com fonts.googleapis.com",
		// We use data URI to show QR image.
		// We can display external profile picture.
		"img-src 'self' http: https: data:",
		"object-src 'none'",
		"base-uri 'none'",
		// https://github.com/w3c/webappsec-csp/issues/7
		// 'self' does not include websocket in Safari :(
		fmt.Sprintf("connect-src 'self' https://www.google-analytics.com ws://%s wss://%s", u.Host, u.Host),
		"block-all-mixed-content",
		"frame-ancestors 'none'",
	}, nil
}
