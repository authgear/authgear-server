package otelutil

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

// ContextWithClonedLabeler clones the current labeler in ctx (if any),
// appends attrs, and returns a context with the cloned labeler.
func ContextWithClonedLabeler(ctx context.Context, attrs ...attribute.KeyValue) context.Context {
	currentLabeler, _ := otelhttp.LabelerFromContext(ctx)
	clonedLabeler := &otelhttp.Labeler{}
	if currentLabeler != nil {
		for _, attr := range currentLabeler.Get() {
			clonedLabeler.Add(attr)
		}
	}
	for _, attr := range attrs {
		clonedLabeler.Add(attr)
	}
	return otelhttp.ContextWithLabeler(ctx, clonedLabeler)
}
