package otelauthgear

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

// meter is the global meter for metrics produced by Authgear.
// You use meter to define metrics in this package.
var meter = otel.Meter("github.com/authgear/authgear-server/pkg/lib/otelauthgear")

// AttributeKeyProjectID defines the attribute.
var AttributeKeyProjectID = attribute.Key("authgear.project_id")

// AttributeKeyClientID defines the attribute.
var AttributeKeyClientID = attribute.Key("authgear.client_id")

// AttributeKeyStatus defines the attribute.
var AttributeKeyStatus = attribute.Key("status")

// AttributeStatusOK is "status=ok".
var AttributeStatusOK = AttributeKeyStatus.String("ok")

// AttributeStatusError is "status=error".
var AttributeStatusError = AttributeKeyStatus.String("error")

func mustInt64Counter(name string, options ...metric.Int64CounterOption) metric.Int64Counter {
	counter, err := meter.Int64Counter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

// IntCounter is metric.Int64Counter or metric.Int64UpDownCounter
type IntCounter interface {
	Add(ctx context.Context, incr int64, options ...metric.AddOption)
}

// IntCounterAddOne prepares necessary attributes and calls Add with incr=1.
func IntCounterAddOne(ctx context.Context, counter IntCounter, inOptions ...metric.AddOption) {
	res := otelutil.GetResource(ctx)
	resAttrs := otelutil.ExtractAttributesFromResource(res)
	resAttrsOption := metric.WithAttributes(resAttrs...)

	finalOptions := []metric.AddOption{resAttrsOption}

	if kv, ok := ctx.Value(AttributeKeyProjectID).(attribute.KeyValue); ok {
		finalOptions = append(finalOptions, metric.WithAttributes(kv))
	}

	if kv, ok := ctx.Value(AttributeKeyClientID).(attribute.KeyValue); ok {
		finalOptions = append(finalOptions, metric.WithAttributes(kv))
	}

	finalOptions = append(finalOptions, inOptions...)

	counter.Add(ctx, 1, finalOptions...)
}
