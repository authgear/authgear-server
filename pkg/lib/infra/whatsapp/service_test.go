package whatsapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func TestServiceSendAuthenticationOTP(t *testing.T) {
	Convey("TestServiceSendAuthenticationOTP", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClock := clock.NewMockClockAt("2020-02-01T00:00:00Z")
		ctx := context.Background()

		// Common LocalizationConfig for tests
		localizationConfig := &config.LocalizationConfig{
			SupportedLanguages: []string{"en"},
			FallbackLanguage:   func() *string { s := "en"; return &s }(),
		}

		Convey("should send authentication OTP via CloudAPI", func() {
			mockCloudAPIClient := NewMockServiceCloudAPIClient(ctrl)
			mockMessageStore := NewMockServiceMessageStore(ctrl)

			cfg := &config.WhatsappConfig{
				APIType_NoDefault:          config.WhatsappAPITypeCloudAPI,
				MessageSentCallbackTimeout: "5s",
			}
			credentials := &config.WhatsappCloudAPICredentials{
				Webhook: &config.WhatsappCloudAPIWebhook{
					VerifyToken: "some-token",
				},
			}

			s := &whatsapp.Service{
				Clock:                 mockClock,
				WhatsappConfig:        cfg,
				LocalizationConfig:    localizationConfig,
				GlobalWhatsappAPIType: config.GlobalWhatsappAPIType(config.WhatsappAPITypeCloudAPI),
				CloudAPIClient:        mockCloudAPIClient,
				MessageStore:          mockMessageStore,
				Credentials:           credentials,
			}

			opts := &whatsapp.SendAuthenticationOTPOptions{
				OTP: "123456",
				To:  "+1234567890",
			}

			messageID := uuid.New()

			mockCloudAPIClient.EXPECT().GetLanguages().Return([]string{"en"}).Times(1)
			mockCloudAPIClient.EXPECT().SendAuthenticationOTP(
				ctx,
				opts,
				"en",
			).Return(messageID, nil).Times(1)

			// Expect message status update check
			mockMessageStore.EXPECT().GetMessageStatus(
				ctx,
				messageID,
			).Return(whatsapp.WhatsappMessageStatusSent, nil).Times(1)

			err := s.SendAuthenticationOTP(ctx, opts)
			So(err, ShouldBeNil)
		})

		Convey("should return error if CloudAPIClient is nil", func() {
			cfg := &config.WhatsappConfig{
				APIType_NoDefault:          config.WhatsappAPITypeCloudAPI,
				MessageSentCallbackTimeout: "5s",
			}

			s := &whatsapp.Service{
				WhatsappConfig:        cfg,
				LocalizationConfig:    localizationConfig,
				GlobalWhatsappAPIType: config.GlobalWhatsappAPIType(config.WhatsappAPITypeCloudAPI),
				CloudAPIClient:        nil,
			}

			opts := &whatsapp.SendAuthenticationOTPOptions{
				OTP: "123456",
				To:  "+1234567890",
			}

			err := s.SendAuthenticationOTP(ctx, opts)
			So(err, ShouldBeError, whatsapp.ErrNoAvailableWhatsappClient)
		})

		Convey("should return error if SendAuthenticationOTP fails", func() {
			mockCloudAPIClient := NewMockServiceCloudAPIClient(ctrl)

			cfg := &config.WhatsappConfig{
				APIType_NoDefault:          config.WhatsappAPITypeCloudAPI,
				MessageSentCallbackTimeout: "5s",
			}

			s := &whatsapp.Service{
				WhatsappConfig:        cfg,
				LocalizationConfig:    localizationConfig,
				GlobalWhatsappAPIType: config.GlobalWhatsappAPIType(config.WhatsappAPITypeCloudAPI),
				CloudAPIClient:        mockCloudAPIClient,
			}

			opts := &whatsapp.SendAuthenticationOTPOptions{
				OTP: "123456",
				To:  "+1234567890",
			}

			expectedError := errors.New("failed to send OTP")
			mockCloudAPIClient.EXPECT().GetLanguages().Return([]string{"en"}).Times(1)
			mockCloudAPIClient.EXPECT().SendAuthenticationOTP(
				ctx,
				opts,
				"en",
			).Return("", expectedError).Times(1)

			err := s.SendAuthenticationOTP(ctx, opts)
			So(err, ShouldBeError, expectedError)
		})

		Convey("should return error if message status update times out", func() {
			mockCloudAPIClient := NewMockServiceCloudAPIClient(ctrl)
			mockMessageStore := NewMockServiceMessageStore(ctrl)

			cfg := &config.WhatsappConfig{
				APIType_NoDefault:          config.WhatsappAPITypeCloudAPI,
				MessageSentCallbackTimeout: "5s",
			}
			credentials := &config.WhatsappCloudAPICredentials{
				Webhook: &config.WhatsappCloudAPIWebhook{
					VerifyToken: "some-token",
				},
			}

			s := &whatsapp.Service{
				Clock:                 mockClock,
				WhatsappConfig:        cfg,
				LocalizationConfig:    localizationConfig,
				GlobalWhatsappAPIType: config.GlobalWhatsappAPIType(config.WhatsappAPITypeCloudAPI),
				CloudAPIClient:        mockCloudAPIClient,
				MessageStore:          mockMessageStore,
				Credentials:           credentials,
			}

			opts := &whatsapp.SendAuthenticationOTPOptions{
				OTP: "123456",
				To:  "+1234567890",
			}

			messageID := uuid.New()

			mockCloudAPIClient.EXPECT().GetLanguages().Return([]string{"en"}).Times(1)
			mockCloudAPIClient.EXPECT().SendAuthenticationOTP(
				ctx,
				opts,
				"en",
			).Return(messageID, nil).Times(1)

			mockMessageStore.EXPECT().GetMessageStatus(
				ctx,
				messageID,
			).DoAndReturn(func(ctx context.Context, messageID string) (whatsapp.WhatsappMessageStatus, error) {
				// Simulate timeout by advancing clock
				mockClock.AdvanceSeconds(6)
				return "", errors.New("timeout")
			}).Times(1)

			err := s.SendAuthenticationOTP(ctx, opts)
			So(err, ShouldBeError, whatsapp.ErrInvalidWhatsappUser)
		})

		Convey("should return error if message status is Failed", func() {
			mockCloudAPIClient := NewMockServiceCloudAPIClient(ctrl)
			mockMessageStore := NewMockServiceMessageStore(ctrl)

			cfg := &config.WhatsappConfig{
				APIType_NoDefault:          config.WhatsappAPITypeCloudAPI,
				MessageSentCallbackTimeout: "5s",
			}
			credentials := &config.WhatsappCloudAPICredentials{
				Webhook: &config.WhatsappCloudAPIWebhook{
					VerifyToken: "some-token",
				},
			}

			s := &whatsapp.Service{
				Clock:                 mockClock,
				WhatsappConfig:        cfg,
				LocalizationConfig:    localizationConfig,
				GlobalWhatsappAPIType: config.GlobalWhatsappAPIType(config.WhatsappAPITypeCloudAPI),
				CloudAPIClient:        mockCloudAPIClient,
				MessageStore:          mockMessageStore,
				Credentials:           credentials,
			}

			opts := &whatsapp.SendAuthenticationOTPOptions{
				OTP: "123456",
				To:  "+1234567890",
			}

			messageID := uuid.New()

			mockCloudAPIClient.EXPECT().GetLanguages().Return([]string{"en"}).Times(1)
			mockCloudAPIClient.EXPECT().SendAuthenticationOTP(
				ctx,
				opts,
				"en",
			).Return(messageID, nil).Times(1)

			mockMessageStore.EXPECT().GetMessageStatus(
				ctx,
				messageID,
			).Return(whatsapp.WhatsappMessageStatusFailed, nil).Times(1)

			err := s.SendAuthenticationOTP(ctx, opts)
			So(err, ShouldBeError, whatsapp.ErrInvalidWhatsappUser)
		})
	})
}
