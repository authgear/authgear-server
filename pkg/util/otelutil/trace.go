package otelutil

import (
	"context"
	"encoding/base64"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
)

const (
	queryNameTraceParent = "x_traceparent"
	queryNameBaggage     = "x_baggage"

	headerNameTraceParent = "traceparent"
	headerNameBaggage     = "baggage"

	baggageKeySDKUserID   = "authgear_sdk_user_id"
	baggageKeySDKDeviceID = "authgear_sdk_device_id"
)

type queryCarrier url.Values

var _ propagation.TextMapCarrier = queryCarrier{}

func (c queryCarrier) Get(key string) string {
	switch key {
	case headerNameTraceParent:
		return url.Values(c).Get(queryNameTraceParent)
	case headerNameBaggage:
		b64 := url.Values(c).Get(queryNameBaggage)
		if b64 != "" {
			if decoded, err := base64.RawURLEncoding.DecodeString(b64); err == nil {
				return string(decoded)
			}
		}
	}
	return url.Values(c).Get(key)
}

func (c queryCarrier) Set(key string, value string) {
	switch key {
	case headerNameTraceParent:
		url.Values(c).Set(queryNameTraceParent, value)
	case headerNameBaggage:
		url.Values(c).Set(queryNameBaggage, base64.RawURLEncoding.EncodeToString([]byte(value)))
	default:
		url.Values(c).Set(key, value)
	}
}

func (c queryCarrier) Keys() []string {
	keys := make([]string, 0, len(url.Values(c)))
	for k := range url.Values(c) {
		keys = append(keys, k)
	}
	return keys
}

// InjectTraceContextToURL injects the current trace context into the query parameters of the given URL.
func InjectTraceContextToURL(ctx context.Context, u *url.URL) *url.URL {
	q := u.Query()
	otel.GetTextMapPropagator().Inject(ctx, queryCarrier(q))
	newURL := *u
	newURL.RawQuery = q.Encode()
	return &newURL
}

// GetAuthgearBaggage returns the Authgear-specific baggage from the context.
// The baggage is expected to contain keys "authgear_sdk_user_id" and "authgear_sdk_device_id".
func GetAuthgearBaggage(ctx context.Context) map[string]string {
	b := baggage.FromContext(ctx)
	m := make(map[string]string)

	userID := b.Member(baggageKeySDKUserID)
	if userID.Key() != "" {
		val := userID.Value()
		// Safety check: ignore unexpectedly large values.
		if len(val) <= 512 {
			m[baggageKeySDKUserID] = val
		}
	}

	deviceID := b.Member(baggageKeySDKDeviceID)
	if deviceID.Key() != "" {
		val := deviceID.Value()
		// Safety check: ignore unexpectedly large values.
		if len(val) <= 512 {
			m[baggageKeySDKDeviceID] = val
		}
	}

	if len(m) == 0 {
		return nil
	}
	return m
}
