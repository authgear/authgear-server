package deps

import "context"

type requestProviderContextKeyType struct{}

var requestProviderContextKey = requestProviderContextKeyType{}

func WithRequestProvider(ctx context.Context, p *RequestProvider) context.Context {
	return context.WithValue(ctx, requestProviderContextKey, p)
}

func GetRequestProvider(ctx context.Context) *RequestProvider {
	p := ctx.Value(requestProviderContextKey).(*RequestProvider)
	return p
}
