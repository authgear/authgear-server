package httputil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHTTPPermissionsPolicy_String(t *testing.T) {
	Convey("Given a HTTPPermissionsPolicy with two directives", t, func() {
		policy := HTTPPermissionsPolicy{
			{PermissionsPolicyDirectiveAutoplay, PermissionsPolicyAllowlistAll},
			{PermissionsPolicyDirectiveCamera, PermissionsPolicyAllowlistNone},
		}

		actual := policy.String()
		expected := "autoplay=*, camera=()"
		So(actual, ShouldEqual, expected)
	})

	Convey("Given a HTTPPermissionsPolicy with no directives", t, func() {
		policy := HTTPPermissionsPolicy{}

		actual := policy.String()
		So(actual, ShouldEqual, "")
	})

	Convey("Using the default permissions policy", t, func() {
		actual := HTTPPermissionsPolicy(DefaultPermissionsPolicy).String()
		expected := "accelerometer=(), ambient-light-sensor=(), autoplay=*, battery=(), bluetooth=(), browsing-topics=(), camera=(), display-capture=(), document-domain=(), encrypted-media=(), execution-while-not-rendered=*, execution-while-out-of-viewport=*, fullscreen=*, gamepad=(), geolocation=(), gyroscope=(), hid=(), identity-credentials-get=(), idle-detection=(), local-fonts=(), magnetometer=(), microphone=(), midi=(), otp-credentials=(), payment=(), picture-in-picture=(), publickey-credentials-create=(self), publickey-credentials-get=(self), screen-wake-lock=(), serial=(), speaker-selection=(), storage-access=(), usb=(), web-share=(), window-management=(), xr-spatial-tracking=()"
		So(actual, ShouldEqual, expected)
	})
}
