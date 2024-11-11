package otelutil

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func SetupOTelHTTPLabeler(ctx context.Context) context.Context {
	labeler := &otelhttp.Labeler{}
	res := GetResource(ctx)
	attrs := ExtractAttributesFromResource(res)
	labeler.Add(attrs...)
	ctx = otelhttp.ContextWithLabeler(ctx, labeler)
	return ctx
}
