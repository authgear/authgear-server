package otp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOTPDeliveryStatusInternalToAPIStatus(t *testing.T) {
	Convey("ToAPIStatus reports whether a poll can advance the status", t, func() {
		Convey("a status that something else can advance is reported as sending", func() {
			So(OTPDeliveryStatusInternalWaitingToSend.ToAPIStatus(), ShouldEqual, model.OTPDeliveryStatusSending)
			So(OTPDeliveryStatusInternalWaitingForConfirmation.ToAPIStatus(), ShouldEqual, model.OTPDeliveryStatusSending)
		})

		Convey("wont_send is indistinguishable from sent, so we do not reveal that no message was sent", func() {
			So(OTPDeliveryStatusInternalWontSend.ToAPIStatus(), ShouldEqual, model.OTPDeliveryStatusSent)
			So(OTPDeliveryStatusInternalSent.ToAPIStatus(), ShouldEqual, model.OTPDeliveryStatusSent)
		})

		Convey("error is reported as failed", func() {
			So(OTPDeliveryStatusInternalFailed.ToAPIStatus(), ShouldEqual, model.OTPDeliveryStatusFailed)
		})
	})
}

func TestDeriveLegacyDeliveryStatus(t *testing.T) {
	Convey("deriveLegacyDeliveryStatus reproduces the behaviour of a code stored without InternalDeliveryStatus", t, func() {
		Convey("no channel means no delivery attempt was recorded", func() {
			So(deriveLegacyDeliveryStatus(&Code{}), ShouldEqual, OTPDeliveryStatusInternalWaitingToSend)
		})

		Convey("a send error is terminal", func() {
			So(deriveLegacyDeliveryStatus(&Code{
				OOBChannel:       model.AuthenticatorOOBChannelEmail,
				SendMessageError: &apierrors.APIError{},
			}), ShouldEqual, OTPDeliveryStatusInternalFailed)
		})

		Convey("whatsapp awaits an asynchronous confirmation", func() {
			So(deriveLegacyDeliveryStatus(&Code{
				OOBChannel: model.AuthenticatorOOBChannelWhatsapp,
			}), ShouldEqual, OTPDeliveryStatusInternalWaitingForConfirmation)
		})

		Convey("email and sms are terminal once dispatched", func() {
			So(deriveLegacyDeliveryStatus(&Code{
				OOBChannel: model.AuthenticatorOOBChannelEmail,
			}), ShouldEqual, OTPDeliveryStatusInternalSent)
			So(deriveLegacyDeliveryStatus(&Code{
				OOBChannel: model.AuthenticatorOOBChannelSMS,
			}), ShouldEqual, OTPDeliveryStatusInternalSent)
		})
	})
}

func TestNewCodeDeliveryStatus(t *testing.T) {
	Convey("newCode records the initial delivery status", t, func() {
		cfg := loadTestAppConfig()
		svc := newTestService(newTestRateLimiter())
		kind := KindVerification(cfg, model.AuthenticatorOOBChannelEmail)

		Convey("a code that is going to be sent is waiting to send", func() {
			code := svc.newCode(kind, "user@example.com", FormCode, &GenerateOptions{})
			So(code.InternalDeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalWaitingToSend)
		})

		Convey("a code that is never going to be sent wont send", func() {
			code := svc.newCode(kind, "user@example.com", FormCode, &GenerateOptions{SkipSending: true})
			So(code.InternalDeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalWontSend)
		})
	})
}

func TestInspectStateDeliveryStatus(t *testing.T) {
	Convey("InspectState reports the delivery status without revealing whether a message was sent", t, func() {
		cfg := loadTestAppConfig()
		svc := newTestService(newTestRateLimiter())
		kind := KindVerification(cfg, model.AuthenticatorOOBChannelEmail)
		target := "user@example.com"

		store := svc.CodeStore.(*testCodeStore)
		seed := func(internalDeliveryStatus OTPDeliveryStatusInternal) {
			store.codes[store.key(kind.Purpose(), target)] = &Code{
				Target:                 target,
				Purpose:                kind.Purpose(),
				Form:                   FormCode,
				Code:                   "123456",
				ExpireAt:               time.Unix(1700000300, 0).UTC(),
				InternalDeliveryStatus: internalDeliveryStatus,
			}
		}

		Convey("a code generated for an administrator to pass on out of band is reported as sent", func() {
			seed(OTPDeliveryStatusInternalWontSend)

			state, err := svc.InspectState(context.Background(), kind, target, nil)

			So(err, ShouldBeNil)
			So(state.DeliveryStatus, ShouldEqual, model.OTPDeliveryStatusSent)
			So(state.DeliveryError, ShouldBeNil)
		})

		Convey("a code whose delivery attempt has not been recorded is reported as sending", func() {
			seed(OTPDeliveryStatusInternalWaitingToSend)

			state, err := svc.InspectState(context.Background(), kind, target, nil)

			So(err, ShouldBeNil)
			So(state.DeliveryStatus, ShouldEqual, model.OTPDeliveryStatusSending)
		})

		Convey("a dispatched code is reported as sent", func() {
			seed(OTPDeliveryStatusInternalSent)

			state, err := svc.InspectState(context.Background(), kind, target, nil)

			So(err, ShouldBeNil)
			So(state.DeliveryStatus, ShouldEqual, model.OTPDeliveryStatusSent)
		})

		Convey("a code that does not exist is pretended to be sent", func() {
			state, err := svc.InspectState(context.Background(), kind, target, nil)

			So(err, ShouldBeNil)
			So(state.DeliveryStatus, ShouldEqual, model.OTPDeliveryStatusSent)
		})
	})
}

type testWhatsappService struct {
	result *whatsapp.GetMessageStatusResult
	err    error
}

func (s *testWhatsappService) GetMessageStatus(ctx context.Context, messageID string) (*whatsapp.GetMessageStatusResult, error) {
	return s.result, s.err
}

func TestGetOTPMessageDeliverStatusWaitingForConfirmation(t *testing.T) {
	Convey("a code awaiting an asynchronous confirmation is resolved by asking the provider", t, func() {
		svc := newTestService(newTestRateLimiter())
		code := &Code{
			InternalDeliveryStatus: OTPDeliveryStatusInternalWaitingForConfirmation,
			OOBChannel:             model.AuthenticatorOOBChannelWhatsapp,
		}

		Convey("the message id is not known yet, so the confirmation is still outstanding", func() {
			svc.WhatsappService = &testWhatsappService{}

			status, err := svc.getOTPMessageDeliverStatus(context.Background(), code)

			So(err, ShouldBeNil)
			So(status.DeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalWaitingForConfirmation)
		})

		Convey("the provider has no status yet, so the confirmation is still outstanding", func() {
			code.WhatsappMessageID = "message-id"
			svc.WhatsappService = &testWhatsappService{result: nil}

			status, err := svc.getOTPMessageDeliverStatus(context.Background(), code)

			So(err, ShouldBeNil)
			So(status.DeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalWaitingForConfirmation)
		})

		Convey("the provider accepted the message but has not confirmed delivery", func() {
			code.WhatsappMessageID = "message-id"
			svc.WhatsappService = &testWhatsappService{result: &whatsapp.GetMessageStatusResult{
				Status: whatsapp.WhatsappMessageStatusAccepted,
			}}

			status, err := svc.getOTPMessageDeliverStatus(context.Background(), code)

			So(err, ShouldBeNil)
			So(status.DeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalWaitingForConfirmation)
		})

		Convey("the provider confirmed delivery", func() {
			code.WhatsappMessageID = "message-id"
			svc.WhatsappService = &testWhatsappService{result: &whatsapp.GetMessageStatusResult{
				Status: whatsapp.WhatsappMessageStatusDelivered,
			}}

			status, err := svc.getOTPMessageDeliverStatus(context.Background(), code)

			So(err, ShouldBeNil)
			So(status.DeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalSent)
		})

		Convey("the provider reported a failure, which is surfaced with its error", func() {
			code.WhatsappMessageID = "message-id"
			apiError := &apierrors.APIError{}
			svc.WhatsappService = &testWhatsappService{result: &whatsapp.GetMessageStatusResult{
				Status:   whatsapp.WhatsappMessageStatusFailed,
				APIError: apiError,
			}}

			status, err := svc.getOTPMessageDeliverStatus(context.Background(), code)

			So(err, ShouldBeNil)
			So(status.DeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalFailed)
			So(status.DeliveryError, ShouldEqual, apiError)
		})

		Convey("a legacy whatsapp code with no persisted status takes the same path", func() {
			legacy := &Code{
				OOBChannel:        model.AuthenticatorOOBChannelWhatsapp,
				WhatsappMessageID: "message-id",
			}
			svc.WhatsappService = &testWhatsappService{result: &whatsapp.GetMessageStatusResult{
				Status: whatsapp.WhatsappMessageStatusDelivered,
			}}

			status, err := svc.getOTPMessageDeliverStatus(context.Background(), legacy)

			So(err, ShouldBeNil)
			So(status.DeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalSent)
		})
	})
}

func TestUpdateCodeAfterSent(t *testing.T) {
	Convey("updateCodeAfterSent records the outcome of the delivery attempt", t, func() {
		cfg := loadTestAppConfig()
		kind := KindVerification(cfg, model.AuthenticatorOOBChannelEmail)
		target := "user@example.com"
		opts := SendOptions{
			Channel: model.AuthenticatorOOBChannelEmail,
			Target:  target,
			Form:    FormCode,
			Kind:    kind,
		}

		newSenderWithCode := func(seed *Code) (*MessageSender, func() *Code) {
			store := newTestCodeStore()
			seed.Target = target
			seed.Purpose = kind.Purpose()
			store.codes[store.key(kind.Purpose(), target)] = seed
			sender := &MessageSender{CodeStore: store}
			return sender, func() *Code {
				return store.codes[store.key(kind.Purpose(), target)]
			}
		}

		waitingToSend := func() *Code {
			return &Code{InternalDeliveryStatus: OTPDeliveryStatusInternalWaitingToSend}
		}

		Convey("a dispatch with no asynchronous confirmation is terminal", func() {
			sender, get := newSenderWithCode(waitingToSend())

			err := sender.updateCodeAfterSent(context.Background(), opts, afterSentResult{})

			So(err, ShouldBeNil)
			So(get().InternalDeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalSent)
			So(get().SendMessageError, ShouldBeNil)
			So(get().OOBChannel, ShouldEqual, model.AuthenticatorOOBChannelEmail)
		})

		Convey("a dispatch whose outcome is confirmed later awaits confirmation", func() {
			sender, get := newSenderWithCode(waitingToSend())

			err := sender.updateCodeAfterSent(context.Background(), opts, afterSentResult{
				AwaitConfirmation: true,
				WhatsappMessageID: "message-id",
			})

			So(err, ShouldBeNil)
			So(get().InternalDeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalWaitingForConfirmation)
			So(get().WhatsappMessageID, ShouldEqual, "message-id")
		})

		Convey("a send error is recorded so a later reader can report it", func() {
			sender, get := newSenderWithCode(waitingToSend())

			err := sender.updateCodeAfterSent(context.Background(), opts, afterSentResult{
				SendError: errors.New("smtp is down"),
			})

			So(err, ShouldBeNil)
			So(get().InternalDeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalFailed)
			So(get().SendMessageError, ShouldNotBeNil)
		})

		Convey("a canceled request is still recorded, because the update outlives the request", func() {
			sender, get := newSenderWithCode(waitingToSend())
			canceledCtx, cancel := context.WithCancel(context.Background())
			cancel()

			err := sender.updateCodeAfterSent(canceledCtx, opts, afterSentResult{})

			So(err, ShouldBeNil)
			So(get().InternalDeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalSent)
		})

		Convey("a recorded failure is terminal and is not overwritten by a later success", func() {
			sender, get := newSenderWithCode(&Code{
				InternalDeliveryStatus: OTPDeliveryStatusInternalFailed,
				SendMessageError:       &apierrors.APIError{},
			})

			err := sender.updateCodeAfterSent(context.Background(), opts, afterSentResult{})

			So(err, ShouldBeNil)
			So(get().InternalDeliveryStatus, ShouldEqual, OTPDeliveryStatusInternalFailed)
		})

		Convey("a legacy code carrying only a send error is also treated as terminal", func() {
			sender, get := newSenderWithCode(&Code{
				SendMessageError: &apierrors.APIError{},
			})

			err := sender.updateCodeAfterSent(context.Background(), opts, afterSentResult{})

			So(err, ShouldBeNil)
			So(get().InternalDeliveryStatus, ShouldBeEmpty)
		})
	})
}
