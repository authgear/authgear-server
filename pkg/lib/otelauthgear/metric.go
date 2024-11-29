package otelauthgear

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Suppose you have a task to add a new metric, what should you do?
//
// The first step is to learn what metric instruments you can use.
// Read this.
// https://opentelemetry.io/docs/concepts/signals/metrics/#metric-instruments
//
// Now that you know what metric instruments are available.
// You need to understand the metric you need to track, and decide the instrument.
// For example, suppose you are tasked to track the number of signups.
// Since the number of signups only go up, you should use a Counter.
// It is trivial to locate the location where a signup occurs in the codebase,
// you should use the non-async version of a Counter.
// The number of signups is an integer, so you should use a integer version of a Counter.
// Therefore, you should use a Int64Counter.
//
// The next step is to define the instrument.
// If you have read https://opentelemetry.io/docs/concepts/signals/metrics/#metric-instruments
// then you already know an instrument have 4 properties, namely, name, kind, unit, and description.
// We have figured out the kind of the instrument, so we have the remaining 3 to deal with.
// Read https://opentelemetry.io/docs/specs/semconv/general/metrics/#general-guidelines
// to understand the general naming guidelines.
// In short, do this
// - Use lowercase characters.
// - Use dot (.) as separator.
// - Use underscore (_) to separate words.
// - Use "count" instead pluralization, https://opentelemetry.io/docs/specs/semconv/general/metrics/#use-count-instead-of-pluralization-for-updowncounters
// The instrument name should be "authgear.signup.count"
// For the description, use a short sentence that can describe the instrument.
// For the unit, read this https://opentelemetry.io/docs/specs/semconv/general/metrics/#instrument-units
// The unit in this case should be "{signup}".
//
// Then you need to decide whether the instrument has any meaningful attributes.
// For signup, there is no.
// You may be tempted to use attribute "status" with value "ok" and "error"
// But that does not make much sense.
// This is because a signup can fail for many reasons along the journey.
// You cannot really pinpoint a location where you can say it is status=error.
// What metric can have "status" then?
// A good example is the metric for tracking the number of email delivery.
// We know where the email delivery happens.
// It happens when we call the SMTP server.
// So that metric can have "status".
//
// In the last step, you locate the locations where you need to insert the instrumentation code.
// You should use the helper functions defined in this package, like IntCounterAddOne.
// These helper functions ensure the metric has necessary attributes attached with them.
//
// See https://github.com/authgear/authgear-server/pull/4906 for an example.

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

var CounterOAuthSessionCreationCount = mustInt64Counter(
	"authgear.oauth_session.creation.count",
	metric.WithDescription("The number of creation of OAuth session"),
	metric.WithUnit("{session}"),
)

var CounterSAMLSessionCreationCount = mustInt64Counter(
	"authgear.saml_session.creation.count",
	metric.WithDescription("The number of creation of SAML session"),
	metric.WithUnit("{session}"),
)

var CounterAuthflowSessionCreationCount = mustInt64Counter(
	"authgear.authflow_session.creation.count",
	metric.WithDescription("The number of creation of Authflow session"),
	metric.WithUnit("{session}"),
)

var CounterWebSessionCreationCount = mustInt64Counter(
	"authgear.web_session.creation.count",
	metric.WithDescription("The number of creation of Web session"),
	metric.WithUnit("{session}"),
)

var CounterOAuthAuthorizationCodeCreationCount = mustInt64Counter(
	"authgear.oauth_authorization_code.creation.count",
	metric.WithDescription("The number of creation of OAuth authorization code"),
	metric.WithUnit("{code}"),
)
var CounterOAuthAuthorizationCodeConsumptionCount = mustInt64Counter(
	"authgear.oauth_authorization_code.consumption.count",
	metric.WithDescription("The number of consumption of OAuth authorization code"),
	metric.WithUnit("{code}"),
)

var CounterOAuthAccessTokenRefreshCount = mustInt64Counter(
	"authgear.oauth_access_token.refresh.count",
	metric.WithDescription("The number of access token obtained via a refresh token"),
	metric.WithUnit("{token}"),
)

// CounterEmailRequestCount has the following labels:
// - AttributeKeyStatus
var CounterEmailRequestCount = mustInt64Counter(
	"authgear.email.request.count",
	metric.WithDescription("The number of email request"),
	metric.WithUnit("{request}"),
)

// CounterSMSRequestCount has the following labels:
// - AttributeKeyStatus
var CounterSMSRequestCount = mustInt64Counter(
	"authgear.sms.request.count",
	metric.WithDescription("The number of SMS request"),
	metric.WithUnit("{request}"),
)

// CounterWhatsappRequestCount has the following labels:
// - AttributeKeyStatus
var CounterWhatsappRequestCount = mustInt64Counter(
	"authgear.whatsapp.request.count",
	metric.WithDescription("The number of Whatsapp request"),
	metric.WithUnit("{request}"),
)

// CounterCSRFRequestCount has the following labels:
// - AttributeKeyStatus
var CounterCSRFRequestCount = mustInt64Counter(
	"authgear.csrf.request.count",
	metric.WithDescription("The number of HTTP request with CSRF protection"),
	metric.WithUnit("{request}"),
)

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
	var finalOptions []metric.AddOption

	if kv, ok := ctx.Value(AttributeKeyProjectID).(attribute.KeyValue); ok {
		finalOptions = append(finalOptions, metric.WithAttributes(kv))
	}

	if kv, ok := ctx.Value(AttributeKeyClientID).(attribute.KeyValue); ok {
		finalOptions = append(finalOptions, metric.WithAttributes(kv))
	}

	finalOptions = append(finalOptions, inOptions...)

	counter.Add(ctx, 1, finalOptions...)
}
