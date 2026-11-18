package otp

import (
	"context"
	neturl "net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type testSenderEndpoints struct{}

func (testSenderEndpoints) Origin() *neturl.URL {
	return &neturl.URL{Scheme: "https", Host: "auth.example.com"}
}

func (testSenderEndpoints) LoginLinkVerificationEndpointURL() *neturl.URL {
	return &neturl.URL{Scheme: "https", Host: "auth.example.com", Path: "/authflow/v2/verify_login_link"}
}

func (testSenderEndpoints) ResetPasswordEndpointURL() *neturl.URL {
	return &neturl.URL{Scheme: "https", Host: "auth.example.com", Path: "/authflow/v2/reset_password"}
}

func TestMessageSenderBuildLinkURL(t *testing.T) {
	Convey("MessageSender.buildLinkURL", t, func() {
		ctx := context.Background()
		newSender := func(ui *config.UIConfig) *MessageSender {
			return &MessageSender{
				Endpoints: testSenderEndpoints{},
				UIConfig:  ui,
			}
		}
		project := &config.UIConfig{
			CustomUIEndpoints: &config.CustomUIEndpointsConfig{
				LoginLink:       "https://custom.example.com/login-link?a=b",
				AccountRecovery: "https://custom.example.com/reset",
			},
		}
		loginLink := SendOptions{OTP: "CODE"}
		adminReset := SendOptions{OTP: "CODE", IsAdminAPIResetPassword: true}

		Convey("uses Auth UI when no Custom UI page is set", func() {
			s := newSender(&config.UIConfig{})
			u := s.buildLinkURL(ctx, translation.MessageTypeAuthenticatePrimaryOOB, loginLink)
			So(u.String(), ShouldEqual, "https://auth.example.com/authflow/v2/verify_login_link?code=CODE")

			u = s.buildLinkURL(ctx, translation.MessageTypeForgotPassword, adminReset)
			So(u.String(), ShouldEqual, "https://auth.example.com/authflow/v2/reset_password?code=CODE&x_from_admin_api=true")
		})

		Convey("uses the project page and keeps its query parameters", func() {
			s := newSender(project)
			u := s.buildLinkURL(ctx, translation.MessageTypeSetupPrimaryOOB, loginLink)
			So(u.String(), ShouldEqual, "https://custom.example.com/login-link?a=b&code=CODE")

			u = s.buildLinkURL(ctx, translation.MessageTypeForgotPassword, adminReset)
			So(u.String(), ShouldEqual, "https://custom.example.com/reset?code=CODE")
		})

		Convey("sets URL and Link to the same link", func() {
			s := newSender(project)
			opts := SendOptions{
				OTP:     "CODE",
				Form:    FormLink,
				Channel: "email",
				Target:  "user@example.com",
			}
			variables, err := s.setupTemplateContext(ctx, translation.MessageTypeAuthenticatePrimaryOOB, opts)
			So(err, ShouldBeNil)
			So(variables.Link, ShouldEqual, "https://custom.example.com/login-link?a=b&code=CODE")
			So(variables.URL, ShouldEqual, variables.Link)
		})
	})
}
