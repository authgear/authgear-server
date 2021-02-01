package wechat

import (
	"context"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type wechatContext struct {
	WeChatRedirectURI string
	Platform          string
}

func WithWeChatRedirectURI(ctx context.Context, weChatRedirectURI string) context.Context {
	v, ok := ctx.Value(contextKey).(*wechatContext)
	if ok {
		v.WeChatRedirectURI = weChatRedirectURI
		return ctx
	}

	return context.WithValue(ctx, contextKey, &wechatContext{
		WeChatRedirectURI: weChatRedirectURI,
	})
}

func GetWeChatRedirectURI(ctx context.Context) string {
	v, ok := ctx.Value(contextKey).(*wechatContext)
	if !ok {
		return ""
	}
	return v.WeChatRedirectURI
}

func WithPlatform(ctx context.Context, platform string) context.Context {
	v, ok := ctx.Value(contextKey).(*wechatContext)
	if ok {
		v.Platform = platform
		return ctx
	}

	return context.WithValue(ctx, contextKey, &wechatContext{
		Platform: platform,
	})
}

func GetPlatform(ctx context.Context) string {
	v, ok := ctx.Value(contextKey).(*wechatContext)
	if !ok {
		return ""
	}
	return v.Platform
}
