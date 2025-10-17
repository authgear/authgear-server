package otelauthgear

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	httpconv "go.opentelemetry.io/otel/semconv/v1.34.0/httpconv"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
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

// attributeKeyProjectID defines the attribute.
// It is private because we expose another function to set it correctly.
var attributeKeyProjectID = attribute.Key("authgear.project_id")

// attributeKeyClientID defines the attribute.
// It is private because we expose another function to set it correctly.
var attributeKeyClientID = attribute.Key("authgear.client_id")

// AttributeKeyStatus defines the attribute.
var AttributeKeyStatus = attribute.Key("status")

// AttributeKeyWhatsappAPIType defines the attribute.
var AttributeKeyWhatsappAPIType = attribute.Key("whatsapp_api_type")

// AttributeKeyWhatsappAPIErrorCode defines the attribute.
var AttributeKeyWhatsappAPIErrorCode = attribute.Key("whatsapp_api_error_code")

// AttributeKeyWhatsappAPIErrorSubcode defines the attribute.
var AttributeKeyWhatsappAPIErrorSubcode = attribute.Key("whatsapp_api_error_subcode")

// AttributeKeyWhatsappAPIMessageStatus defines the attribute.
var AttributeKeyWhatsappAPIMessageStatus = attribute.Key("whatsapp_api_message_status")

// AttributeKeyWhatsappAPIIsTimeout defines the attribute.
var AttributeKeyWhatsappAPIIsTimeout = attribute.Key("whatsapp_api_is_timeout")

// AttributeKeyAPIErrorReason defines the attribute.
var AttributeKeyAPIErrorReason = attribute.Key("api_error_reason")

// AttributeStatusOK is "status=ok".
var AttributeStatusOK = AttributeKeyStatus.String("ok")

// AttributeStatusError is "status=error".
var AttributeStatusError = AttributeKeyStatus.String("error")

// AttributeKeyCSRFHasOmitCookie defines the attribute.
var AttributeKeyCSRFHasOmitCookie = attribute.Key("csrf.has_omit_cookie")

// AttributeKeyCSRFHasNoneCookie defines the attribute.
var AttributeKeyCSRFHasNoneCookie = attribute.Key("csrf.has_none_cookie")

// AttributeKeyCSRFHasLaxCookie defines the attribute.
var AttributeKeyCSRFHasLaxCookie = attribute.Key("csrf.has_lax_cookie")

// AttributeKeyCSRFHasStrictCookie defines the attribute.
var AttributeKeyCSRFHasStrictCookie = attribute.Key("csrf.has_strict_cookie")

// AttributeKeyGorillaCSRFFailureReason defines the attribute.
var AttributeKeyGorillaCSRFFailureReason = attribute.Key("gorilla_csrf.failure_reason")

var CounterOAuthSessionCreationCount = otelutil.MustInt64Counter(
	meter,
	"authgear.oauth_session.creation.count",
	metric.WithDescription("The number of creation of OAuth session"),
	metric.WithUnit("{session}"),
)

var CounterSAMLSessionCreationCount = otelutil.MustInt64Counter(
	meter,
	"authgear.saml_session.creation.count",
	metric.WithDescription("The number of creation of SAML session"),
	metric.WithUnit("{session}"),
)

var CounterAuthflowSessionCreationCount = otelutil.MustInt64Counter(
	meter,
	"authgear.authflow_session.creation.count",
	metric.WithDescription("The number of creation of Authflow session"),
	metric.WithUnit("{session}"),
)

var CounterWebSessionCreationCount = otelutil.MustInt64Counter(
	meter,
	"authgear.web_session.creation.count",
	metric.WithDescription("The number of creation of Web session"),
	metric.WithUnit("{session}"),
)

var CounterOAuthAuthorizationCodeCreationCount = otelutil.MustInt64Counter(
	meter,
	"authgear.oauth_authorization_code.creation.count",
	metric.WithDescription("The number of creation of OAuth authorization code"),
	metric.WithUnit("{code}"),
)
var CounterOAuthAuthorizationCodeConsumptionCount = otelutil.MustInt64Counter(
	meter,
	"authgear.oauth_authorization_code.consumption.count",
	metric.WithDescription("The number of consumption of OAuth authorization code"),
	metric.WithUnit("{code}"),
)

var CounterOAuthAccessTokenRefreshCount = otelutil.MustInt64Counter(
	meter,
	"authgear.oauth_access_token.refresh.count",
	metric.WithDescription("The number of access token obtained via a refresh token"),
	metric.WithUnit("{token}"),
)

// CounterEmailRequestCount has the following labels:
// - AttributeKeyStatus
var CounterEmailRequestCount = otelutil.MustInt64Counter(
	meter,
	"authgear.email.request.count",
	metric.WithDescription("The number of email request"),
	metric.WithUnit("{request}"),
)

// CounterSMSRequestCount has the following labels:
// - AttributeKeyStatus
var CounterSMSRequestCount = otelutil.MustInt64Counter(
	meter,
	"authgear.sms.request.count",
	metric.WithDescription("The number of SMS request"),
	metric.WithUnit("{request}"),
)

// CounterWhatsappRequestCount has the following labels:
// - AttributeKeyStatus
var CounterWhatsappRequestCount = otelutil.MustInt64Counter(
	meter,
	"authgear.whatsapp.request.count",
	metric.WithDescription("The number of Whatsapp request"),
	metric.WithUnit("{request}"),
)

// CounterCSRFRequestCount has the following labels:
// - AttributeKeyStatus
// - AttributeKeyCSRFHasOmitCookie
// - AttributeKeyCSRFHasNoneCookie
// - AttributeKeyCSRFHasLaxCookie
// - AttributeKeyCSRFHasStrictCookie
var CounterCSRFRequestCount = otelutil.MustInt64Counter(
	meter,
	"authgear.csrf.request.count",
	metric.WithDescription("The number of HTTP request with CSRF protection"),
	metric.WithUnit("{request}"),
)

// CounterNonBlockingWebhookCount has the following labels:
// - AttributeKeyStatus
var CounterNonBlockingWebhookCount = otelutil.MustInt64Counter(
	meter,
	"authgear.webhook.non_blocking.count",
	metric.WithDescription("The number of non blocking webhook"),
	metric.WithUnit("{request}"),
)

// CounterBlockingWebhookCount has the following labels:
// - AttributeKeyStatus
var CounterBlockingWebhookCount = otelutil.MustInt64Counter(
	meter,
	"authgear.webhook.blocking.count",
	metric.WithDescription("The number of blocking webhook"),
	metric.WithUnit("{request}"),
)

// HTTPServerRequestDurationHistogram is https://opentelemetry.io/docs/specs/semconv/http/http-metrics/#metric-httpserverrequestduration
var HTTPServerRequestDurationHistogram, _ = httpconv.NewServerRequestDuration(
	meter,
	// The spec says we SHOULD define explicit boundaries.
	// https://opentelemetry.io/docs/specs/semconv/http/http-metrics/#metric-httpserverrequestduration
	// In fact, if we do not, the default boundary is not suitable for request duration scale.
	metric.WithExplicitBucketBoundaries(
		0.005,
		0.01,
		0.025,
		0.05,
		0.075,
		0.1,
		0.25,
		0.5,
		0.75,
		1,
		2.5,
		5,
		7.5,
		10,
	),
)

type metricOptionAttributeKeyValue struct {
	attribute.KeyValue
}

func (o metricOptionAttributeKeyValue) ToOtelMetricOption() metric.MeasurementOption {
	return metric.WithAttributes(o.KeyValue)
}

func WithStatusOk() otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeStatusOK}
}

func WithStatusError() otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeStatusError}
}

func WithWhatsappAPIType(apiType config.WhatsappAPIType) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeKeyWhatsappAPIType.String(string(apiType))}
}

func WithWhatsappAPIErrorCode(code int) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeKeyWhatsappAPIErrorCode.String(fmt.Sprint(code))}
}

func WithWhatsappAPIErrorSubcode(subcode int) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeKeyWhatsappAPIErrorSubcode.String(fmt.Sprint(subcode))}
}

func WithWhatsappAPIMessageStatusAndTimeout(status string, isTimeout bool) []otelutil.MetricOption {
	return []otelutil.MetricOption{
		metricOptionAttributeKeyValue{AttributeKeyWhatsappAPIMessageStatus.String(status)},
		metricOptionAttributeKeyValue{AttributeKeyWhatsappAPIIsTimeout.Bool(isTimeout)},
	}
}

func WithAPIErrorReason(kind string) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeKeyAPIErrorReason.String(kind)}
}

func WithHTTPStatusCode(code int) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{semconv.HTTPResponseStatusCodeKey.Int(code)}
}

func WithCSRFHasOmitCookie(b bool) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeKeyCSRFHasOmitCookie.Bool(b)}
}

func WithCSRFHasNoneCookie(b bool) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeKeyCSRFHasNoneCookie.Bool(b)}
}

func WithCSRFHasLaxCookie(b bool) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeKeyCSRFHasLaxCookie.Bool(b)}
}

func WithCSRFHasStrictCookie(b bool) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeKeyCSRFHasStrictCookie.Bool(b)}
}

func WithGorillaCSRFFailureReason(reason string) otelutil.MetricOption {
	return metricOptionAttributeKeyValue{AttributeKeyGorillaCSRFFailureReason.String(reason)}
}

func SetProjectID(ctx context.Context, projectID string) {
	labeler, _ := otelhttp.LabelerFromContext(ctx)
	labeler.Add(attributeKeyProjectID.String(projectID))
}

func SetClientID(ctx context.Context, clientID string) {
	labeler, _ := otelhttp.LabelerFromContext(ctx)
	labeler.Add(attributeKeyClientID.String(clientID))
}
