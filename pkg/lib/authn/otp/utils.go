package otp

import (
	"context"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var UtilsLogger = slogutil.NewLogger("otp-utils")

func selectByChannel[T any](channel model.AuthenticatorOOBChannel, email T, sms T, whatsapp T) T {
	switch channel {
	case model.AuthenticatorOOBChannelEmail:
		return email
	case model.AuthenticatorOOBChannelSMS:
		return sms
	case model.AuthenticatorOOBChannelWhatsapp:
		return whatsapp
	}
	panic("invalid channel: " + channel)
}

func whatsappMessageStatusToOTPDeliveryStatus(ctx context.Context, messageStatus whatsapp.WhatsappMessageStatus) (model.OTPDeliveryStatus, *apierrors.APIError) {
	var err *apierrors.APIError
	var deliveryStatus model.OTPDeliveryStatus
	switch messageStatus {
	case whatsapp.WhatsappMessageStatusAccepted:
		deliveryStatus = model.OTPDeliveryStatusSending
	case whatsapp.WhatsappMessageStatusSent,
		whatsapp.WhatsappMessageStatusDelivered,
		whatsapp.WhatsappMessageStatusRead:
		deliveryStatus = model.OTPDeliveryStatusSent
	case whatsapp.WhatsappMessageStatusFailed:
		deliveryStatus = model.OTPDeliveryStatusFailed
		// TODO(tung): Check if we can have other errors
		err = apierrors.AsAPIError(whatsapp.ErrInvalidWhatsappUser)
	default:
		UtilsLogger.GetLogger(ctx).With(
			slog.String("status", string(messageStatus)),
		).Error(ctx, "unexpected whatsapp message status")
		deliveryStatus = model.OTPDeliveryStatusFailed
	}
	return deliveryStatus, err
}
