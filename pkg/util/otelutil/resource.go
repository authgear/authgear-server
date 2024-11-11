package otelutil

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

const envvar_X_OTEL_METRICS_RESOURCE_ATTRIBUTES = "X_OTEL_METRICS_RESOURCE_ATTRIBUTES"

type contextKeyResourceType struct{}

var contextKeyResource = contextKeyResourceType{}

func WithResource(ctx context.Context, res *sdkresource.Resource) context.Context {
	return context.WithValue(ctx, contextKeyResource, res)
}

// GetResource retrieves *sdkresource.Resource if it is present in ctx.
// Otherwise, an empty *sdkresource.Resource is returned.
func GetResource(ctx context.Context) *sdkresource.Resource {
	res, ok := ctx.Value(contextKeyResource).(*sdkresource.Resource)
	if ok {
		return res
	}
	return sdkresource.Empty()
}

// ExtractAttributesFromResource looks at the environment variable
// X_OTEL_METRICS_RESOURCE_ATTRIBUTES and copy the attributes into metrics.
// X_OTEL_METRICS_RESOURCE_ATTRIBUTES is a comma-separated list of attribute key.
func ExtractAttributesFromResource(res *sdkresource.Resource) []attribute.KeyValue {
	X_OTEL_METRICS_RESOURCE_ATTRIBUTES := strings.TrimSpace(os.Getenv(envvar_X_OTEL_METRICS_RESOURCE_ATTRIBUTES))

	var keyValues []attribute.KeyValue
	keys := strings.Split(X_OTEL_METRICS_RESOURCE_ATTRIBUTES, ",")
	for _, key := range keys {
		iter := res.Iter()
		for iter.Next() {
			attr := iter.Attribute()
			if string(attr.Key) == key {
				keyValues = append(keyValues, attr)
			}
		}
	}

	return keyValues
}
