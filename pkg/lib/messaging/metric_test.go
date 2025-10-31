package messaging

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	smsapi "github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestApplySMSAPIErrorMetrics(t *testing.T) {
	Convey("TestApplySMSAPIErrorMetrics", t, func() {

		Convey("should return correct options for APIErrorKind", func() {
			apiErrorKind := apierrors.Kind{Reason: "invalid_number"}
			smsapiErr := &smsapi.SendError{APIErrorKind: &apiErrorKind}
			options := ApplySMSAPIErrorMetrics(smsapiErr)
			So(options, ShouldResemble, []otelutil.MetricOption{
				otelauthgear.WithAPIErrorReason("invalid_number"),
			})
		})

		Convey("should return correct options for ProviderType", func() {
			smsapiErr := &smsapi.SendError{ProviderType: config.SMSProviderCustom}
			options := ApplySMSAPIErrorMetrics(smsapiErr)
			So(options, ShouldResemble, []otelutil.MetricOption{
				otelauthgear.WithProviderType("custom"),
			})
		})

		Convey("should return correct options for ProviderErrorCode", func() {
			smsapiErr := &smsapi.SendError{ProviderErrorCode: "100"}
			options := ApplySMSAPIErrorMetrics(smsapiErr)
			So(options, ShouldResemble, []otelutil.MetricOption{
				otelauthgear.WithProviderErrorCode("100"),
			})
		})

		Convey("should return correct options for CustomProviderType", func() {
			smsapiErr := &smsapi.SendError{CustomProviderType: "custom_sms"}
			options := ApplySMSAPIErrorMetrics(smsapiErr)
			So(options, ShouldResemble, []otelutil.MetricOption{
				otelauthgear.WithCustomProviderType("custom_sms"),
			})
		})

		Convey("should return correct options for CustomProviderName", func() {
			smsapiErr := &smsapi.SendError{CustomProviderName: "MySMSProvider"}
			options := ApplySMSAPIErrorMetrics(smsapiErr)
			So(options, ShouldResemble, []otelutil.MetricOption{
				otelauthgear.WithCustomProviderName("MySMSProvider"),
			})
		})

		Convey("should return correct options for CustomProviderResponseCode", func() {
			smsapiErr := &smsapi.SendError{CustomProviderResponseCode: "200"}
			options := ApplySMSAPIErrorMetrics(smsapiErr)
			So(options, ShouldResemble, []otelutil.MetricOption{
				otelauthgear.WithCustomProviderResponseCode("200"),
			})
		})

		Convey("should return all options if all fields are present", func() {
			apiErrorKind := apierrors.Kind{Reason: "invalid_number"}
			smsapiErr := &smsapi.SendError{
				APIErrorKind:               &apiErrorKind,
				ProviderType:               config.SMSProviderNexmo,
				ProviderErrorCode:          "100",
				CustomProviderType:         "custom_sms",
				CustomProviderName:         "MySMSProvider",
				CustomProviderResponseCode: "200",
			}
			options := ApplySMSAPIErrorMetrics(smsapiErr)
			So(options, ShouldResemble, []otelutil.MetricOption{
				otelauthgear.WithAPIErrorReason("invalid_number"),
				otelauthgear.WithProviderType("nexmo"),
				otelauthgear.WithProviderErrorCode("100"),
				otelauthgear.WithCustomProviderType("custom_sms"),
				otelauthgear.WithCustomProviderName("MySMSProvider"),
				otelauthgear.WithCustomProviderResponseCode("200"),
			})
		})
	})
}
