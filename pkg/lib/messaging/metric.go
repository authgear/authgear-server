package messaging

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

func ApplySMSAPIErrorMetrics(smsapiErr *smsapi.SendError) []otelutil.MetricOption {
	var options []otelutil.MetricOption
	if smsapiErr.APIErrorKind != nil {
		options = append(options, otelauthgear.WithAPIErrorReason(smsapiErr.APIErrorKind.Reason))
	}
	if smsapiErr.ProviderType != "" {
		options = append(options, otelauthgear.WithProviderType(string(smsapiErr.ProviderType)))
	}
	if smsapiErr.ProviderErrorCode != "" {
		options = append(options, otelauthgear.WithProviderErrorCode(smsapiErr.ProviderErrorCode))
	}
	if smsapiErr.CustomProviderType != "" {
		options = append(options, otelauthgear.WithCustomProviderType(smsapiErr.CustomProviderType))
	}
	if smsapiErr.CustomProviderName != "" {
		options = append(options, otelauthgear.WithCustomProviderName(smsapiErr.CustomProviderName))
	}
	if smsapiErr.CustomProviderResponseCode != "" {
		options = append(options, otelauthgear.WithCustomProviderResponseCode(smsapiErr.CustomProviderResponseCode))
	}
	return options
}
