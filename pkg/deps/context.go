package deps

import "context"

type requestContainerContextKeyType struct{}

var requestContainerContextKey = requestContainerContextKeyType{}

func WithRequestContainer(ctx context.Context, container *RequestContainer) context.Context {
	return context.WithValue(ctx, requestContainerContextKey, container)
}

func GetRequestContainer(ctx context.Context) *RequestContainer {
	container := ctx.Value(requestContainerContextKey).(*RequestContainer)
	return container
}
