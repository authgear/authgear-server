package loginlink

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestService(t *testing.T) {
	Convey("Service", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		otpCodes := NewMockOTPCodeService(ctrl)
		events := NewMockEventStore(ctrl)
		s := &Service{
			Config:   &config.AppConfig{},
			OTPCodes: otpCodes,
			Events:   events,
		}

		expectValid := func() {
			otpCodes.EXPECT().LookupCode(gomock.Any(), otp.PurposeOOBOTP, "CODE").Return("user@example.com", nil)
			otpCodes.EXPECT().VerifyOTP(gomock.Any(), gomock.Any(), "user@example.com", "CODE", &otp.VerifyOptions{SkipConsume: true}).Return(nil)
		}

		Convey("Inspect", func() {
			Convey("accepts a valid code without approving it", func() {
				expectValid()

				So(s.Inspect(ctx, "CODE"), ShouldBeNil)
			})

			Convey("rejects an unknown code", func() {
				otpCodes.EXPECT().LookupCode(gomock.Any(), otp.PurposeOOBOTP, "UNKNOWN").Return("", otp.ErrCodeNotFound)

				So(s.Inspect(ctx, "UNKNOWN"), ShouldBeError, otp.ErrCodeNotFound)
			})

			Convey("returns the verification error", func() {
				otpCodes.EXPECT().LookupCode(gomock.Any(), otp.PurposeOOBOTP, "CODE").Return("user@example.com", nil)
				otpCodes.EXPECT().VerifyOTP(gomock.Any(), gomock.Any(), "user@example.com", "CODE", gomock.Any()).Return(otp.ErrConsumedCode)

				So(s.Inspect(ctx, "CODE"), ShouldBeError, otp.ErrConsumedCode)
			})
		})

		Convey("Approve", func() {
			Convey("approves the code and refreshes the flow", func() {
				expectValid()
				otpCodes.EXPECT().SetSubmittedCode(gomock.Any(), gomock.Any(), "user@example.com", "CODE").Return(&otp.State{
					AuthenticationFlowWebsocketChannelName: "channel",
				}, nil)
				events.EXPECT().Publish(gomock.Any(), "channel", authflow.NewEventRefresh()).Return(nil)

				So(s.Approve(ctx, "CODE"), ShouldBeNil)
			})

			Convey("does not publish without a websocket channel", func() {
				expectValid()
				otpCodes.EXPECT().SetSubmittedCode(gomock.Any(), gomock.Any(), "user@example.com", "CODE").Return(&otp.State{}, nil)

				So(s.Approve(ctx, "CODE"), ShouldBeNil)
			})

			Convey("rejects an unknown code", func() {
				otpCodes.EXPECT().LookupCode(gomock.Any(), otp.PurposeOOBOTP, "UNKNOWN").Return("", otp.ErrCodeNotFound)

				So(s.Approve(ctx, "UNKNOWN"), ShouldBeError, otp.ErrCodeNotFound)
			})

			Convey("does not approve when verification fails", func() {
				otpCodes.EXPECT().LookupCode(gomock.Any(), otp.PurposeOOBOTP, "CODE").Return("user@example.com", nil)
				otpCodes.EXPECT().VerifyOTP(gomock.Any(), gomock.Any(), "user@example.com", "CODE", gomock.Any()).Return(otp.ErrTooManyAttempts)

				So(s.Approve(ctx, "CODE"), ShouldBeError, otp.ErrTooManyAttempts)
			})
		})
	})
}
