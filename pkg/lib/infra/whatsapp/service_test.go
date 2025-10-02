package whatsapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func TestServiceSendAuthenticationOTP(t *testing.T) {
	Convey("TestServiceSendAuthenticationOTP", t, func() {

		ctx := context.Background()

		localizationConfig := &config.LocalizationConfig{
			SupportedLanguages: []string{"en"},
			FallbackLanguage:   func() *string { s := "en"; return &s }(),
		}

		Convey("should send authentication OTP via CloudAPI", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockCloudAPIClient := NewMockServiceCloudAPIClient(ctrl)
			mockMessageStore := NewMockServiceMessageStore(ctrl)

			cfg := &config.WhatsappConfig{
				APIType_NoDefault:          config.WhatsappAPITypeCloudAPI,
				MessageSentCallbackTimeout: "1ms",
			}
			credentials := &config.WhatsappCloudAPICredentials{
				Webhook: &config.WhatsappCloudAPIWebhook{
					VerifyToken: "some-token",
				},
			}

			s := &whatsapp.Service{
				Clock:                 clock.NewSystemClock(),
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
			sendResult := &whatsapp.CloudAPISendAuthenticationOTPResult{
				MessageID:     messageID,
				MessageStatus: whatsapp.WhatsappMessageStatusAccepted,
			}

			mockCloudAPIClient.EXPECT().GetLanguages().Return([]string{"en"}).Times(1)
			mockCloudAPIClient.EXPECT().SendAuthenticationOTP(
				ctx,
				opts,
				"en",
			).Return(sendResult, nil).Times(1)

			mockMessageStore.EXPECT().SetMessageStatusIfNotExist(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).Return(nil).AnyTimes()

			result, err := s.SendAuthenticationOTP(ctx, opts)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, &whatsapp.SendAuthenticationOTPResult{
				MessageID:     messageID,
				MessageStatus: whatsapp.WhatsappMessageStatusAccepted,
			})
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

			_, err := s.SendAuthenticationOTP(ctx, opts)
			So(err, ShouldBeError, whatsapp.ErrNoAvailableWhatsappClient)
		})

		Convey("should return error if SendAuthenticationOTP fails", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
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
			).Return(nil, expectedError).Times(1)

			_, err := s.SendAuthenticationOTP(ctx, opts)
			So(err, ShouldBeError, expectedError)
		})

		Convey("should call SetMessageStatusIfNotExist after MessageSentCallbackTimeout to set status to WhatsappMessageStatusFailed if webhook is configured", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockCloudAPIClient := NewMockServiceCloudAPIClient(ctrl)
			mockMessageStore := NewMockServiceMessageStore(ctrl)

			cfg := &config.WhatsappConfig{
				APIType_NoDefault:          config.WhatsappAPITypeCloudAPI,
				MessageSentCallbackTimeout: "1s",
			}
			credentials := &config.WhatsappCloudAPICredentials{
				Webhook: &config.WhatsappCloudAPIWebhook{
					VerifyToken: "some-token",
				},
			}

			s := &whatsapp.Service{
				Clock:                 clock.NewSystemClock(),
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
			sendResult := &whatsapp.CloudAPISendAuthenticationOTPResult{
				MessageID:     messageID,
				MessageStatus: whatsapp.WhatsappMessageStatusAccepted,
			}

			mockCloudAPIClient.EXPECT().GetLanguages().Return([]string{"en"}).Times(1)
			mockCloudAPIClient.EXPECT().SendAuthenticationOTP(
				ctx,
				opts,
				"en",
			).Return(sendResult, nil).Times(1)

			mockMessageStore.EXPECT().SetMessageStatusIfNotExist(
				gomock.Any(), // context.WithoutCancel(ctx)
				messageID,
				&whatsapp.WhatsappMessageStatusData{
					Status:    whatsapp.WhatsappMessageStatusFailed,
					IsTimeout: true,
				},
			).Return(nil).Times(1)

			result, err := s.SendAuthenticationOTP(ctx, opts)
			// Wait for the goroutine to finish
			time.Sleep(1500 * time.Millisecond)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, &whatsapp.SendAuthenticationOTPResult{
				MessageID:     messageID,
				MessageStatus: whatsapp.WhatsappMessageStatusAccepted,
			})
		})

		Convey("should call SetMessageStatusIfNotExist to set status to WhatsappMessageStatusDelivered immediately if webhook is not configured", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockCloudAPIClient := NewMockServiceCloudAPIClient(ctrl)
			mockMessageStore := NewMockServiceMessageStore(ctrl)

			cfg := &config.WhatsappConfig{
				APIType_NoDefault:          config.WhatsappAPITypeCloudAPI,
				MessageSentCallbackTimeout: "5s",
			}
			credentials := &config.WhatsappCloudAPICredentials{
				Webhook: nil,
			}

			s := &whatsapp.Service{
				Clock:                 clock.NewSystemClock(),
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
			sendResult := &whatsapp.CloudAPISendAuthenticationOTPResult{
				MessageID:     messageID,
				MessageStatus: whatsapp.WhatsappMessageStatusAccepted,
			}

			mockCloudAPIClient.EXPECT().GetLanguages().Return([]string{"en"}).Times(1)
			mockCloudAPIClient.EXPECT().SendAuthenticationOTP(
				ctx,
				opts,
				"en",
			).Return(sendResult, nil).Times(1)

			mockMessageStore.EXPECT().SetMessageStatusIfNotExist(
				ctx,
				messageID,
				&whatsapp.WhatsappMessageStatusData{
					Status:    whatsapp.WhatsappMessageStatusDelivered,
					IsTimeout: false, // Should not be timeout if no webhook
				},
			).Return(nil).Times(1)

			result, err := s.SendAuthenticationOTP(ctx, opts)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, &whatsapp.SendAuthenticationOTPResult{
				MessageID:     messageID,
				MessageStatus: whatsapp.WhatsappMessageStatusAccepted,
			})
		})
	})
}
