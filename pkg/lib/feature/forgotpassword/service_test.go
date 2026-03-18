package forgotpassword

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func newTestForgotPasswordAppConfig() *config.AppConfig {
	appConfig := &config.AppConfig{
		ForgotPassword: &config.ForgotPasswordConfig{},
	}
	config.SetFieldDefaults(appConfig)
	return appConfig
}

func newTestForgotPasswordFeatureConfig() *config.FeatureConfig {
	return config.NewEffectiveDefaultFeatureConfig()
}

func TestServicePassesAuthenticationFlowIDToGenerateOTP(t *testing.T) {
	Convey("SendCode passes authflow FlowID to OTP GenerateOTP", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		identities := NewMockIdentityService(ctrl)
		authenticators := NewMockAuthenticatorService(ctrl)
		otpCodes := NewMockOTPCodeService(ctrl)
		otpSender := NewMockOTPSender(ctrl)

		svc := &Service{
			Config:         newTestForgotPasswordAppConfig(),
			FeatureConfig:  newTestForgotPasswordFeatureConfig(),
			Identities:     identities,
			Authenticators: authenticators,
			OTPCodes:       otpCodes,
			OTPSender:      otpSender,
		}

		info := (&identity.LoginID{
			UserID:      "user-1",
			LoginIDType: model.LoginIDKeyTypeEmail,
			LoginID:     "user@example.com",
		}).ToInfo()

		identities.EXPECT().ListByClaim(gomock.Any(), string(model.ClaimEmail), "user@example.com").Return([]*identity.Info{info}, nil)
		identities.EXPECT().ListByClaim(gomock.Any(), string(model.ClaimPhoneNumber), "user@example.com").Return(nil, nil)
		authenticators.EXPECT().List(gomock.Any(), "user-1", gomock.Any(), gomock.Any()).Return(nil, nil)
		otpCodes.EXPECT().
			GenerateOTP(gomock.Any(), gomock.Any(), "user@example.com", otp.FormCode, gomock.AssignableToTypeOf(&otp.GenerateOptions{})).
			DoAndReturn(func(ctx context.Context, kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error) {
				So(opt.AuthenticationFlowID, ShouldEqual, "flow-1")
				return "111111", nil
			})
		otpSender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

		err := svc.SendCode(context.Background(), info.IdentityAwareStandardClaims()[model.ClaimEmail], &CodeOptions{
			AuthenticationFlowID: "flow-1",
			Channel:              CodeChannelEmail,
			Kind:                 CodeKindShortCode,
		})
		So(err, ShouldBeNil)
	})
}

func TestServicePassesAuthenticationFlowIDToInspectState(t *testing.T) {
	Convey("InspectState passes authflow FlowID to OTP InspectState", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		otpCodes := NewMockOTPCodeService(ctrl)
		svc := &Service{
			Config:        newTestForgotPasswordAppConfig(),
			FeatureConfig: newTestForgotPasswordFeatureConfig(),
			OTPCodes:      otpCodes,
		}

		otpCodes.EXPECT().
			InspectState(gomock.Any(), gomock.Any(), "user@example.com", gomock.AssignableToTypeOf(&otp.InspectStateOptions{})).
			DoAndReturn(func(ctx context.Context, kind otp.Kind, target string, opts *otp.InspectStateOptions) (*otp.State, error) {
				So(opts.AuthenticationFlowID, ShouldEqual, "flow-1")
				return &otp.State{CanResendAt: time.Unix(1700000000, 0).UTC()}, nil
			})

		_, err := svc.InspectState(context.Background(), "user@example.com", CodeChannelEmail, CodeKindShortCode, &InspectStateOptions{
			AuthenticationFlowID: "flow-1",
		})
		So(err, ShouldBeNil)
	})
}
