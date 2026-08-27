package otp

import (
	"context"
	"errors"
	neturl "net/url"
	"path/filepath"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/messaging"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type AdditionalContext struct {
	HasPassword bool
}

type SendOptions struct {
	Channel                 model.AuthenticatorOOBChannel
	Target                  string
	Form                    Form
	Type                    translation.MessageType
	Kind                    Kind
	OTP                     string
	AdditionalContext       *AdditionalContext
	IsAdminAPIResetPassword bool
}

type EndpointsProvider interface {
	Origin() *neturl.URL
	LoginLinkVerificationEndpointURL() *neturl.URL
	ResetPasswordEndpointURL() *neturl.URL
}

type TranslationService interface {
	EmailMessageData(ctx context.Context, msg *translation.MessageSpec, variables *translation.PartialTemplateVariables) (*translation.EmailMessageData, error)
	SMSMessageData(ctx context.Context, msg *translation.MessageSpec, variables *translation.PartialTemplateVariables) (*translation.SMSMessageData, error)
	WhatsappMessageData(ctx context.Context, language string, msg *translation.MessageSpec, variables *translation.PartialTemplateVariables) (*translation.WhatsappMessageData, error)
}

type Sender interface {
	SendEmailInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *mail.SendOptions) error
	SendSMSImmediately(ctx context.Context, msgType translation.MessageType, opts *sms.SendOptions) error
	SendSMSInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *sms.SendOptions) error
	SendWhatsappInNewGoroutine(ctx context.Context, msgType translation.MessageType, opts *whatsapp.SendAuthenticationOTPOptions, resultCallback messaging.SendWhatsappResultCallback, errCalllback messaging.SendWhatsappErrorCallback) error
}

type SenderCodeStore interface {
	Get(ctx context.Context, purpose Purpose, target string) (*Code, error)
	Update(ctx context.Context, purpose Purpose, code *Code) error
}

type MessageSender struct {
	AppID       config.AppID
	Translation TranslationService
	Endpoints   EndpointsProvider
	Sender      Sender
	CodeStore   SenderCodeStore

	WhatsappConfig *config.WhatsappConfig
}

var SenderLogger = slogutil.NewLogger("otp-sender")

var FromAdminAPIQueryKey = "x_from_admin_api"

func (s *MessageSender) setupTemplateContext(msgType translation.MessageType, opts SendOptions) (*translation.PartialTemplateVariables, error) {
	url := ""
	if opts.Form == FormLink {
		var linkURL *neturl.URL
		switch msgType {
		case translation.MessageTypeSetupPrimaryOOB,
			translation.MessageTypeSetupSecondaryOOB,
			translation.MessageTypeAuthenticatePrimaryOOB,
			translation.MessageTypeAuthenticateSecondaryOOB:

			linkURL = s.Endpoints.LoginLinkVerificationEndpointURL()
			query := linkURL.Query()
			query.Set("code", opts.OTP)
			linkURL.RawQuery = query.Encode()

		case translation.MessageTypeForgotPassword:

			linkURL = s.Endpoints.ResetPasswordEndpointURL()
			query := linkURL.Query()
			query.Set("code", opts.OTP)
			if opts.IsAdminAPIResetPassword {
				query.Set(FromAdminAPIQueryKey, "true")
			}
			linkURL.RawQuery = query.Encode()

		default:
			panic("otp: unexpected message type for link: " + msgType)
		}

		url = linkURL.String()
	}

	ctx := &translation.PartialTemplateVariables{
		Code: opts.OTP,
		URL:  url,
		Link: url,
		Host: s.Endpoints.Origin().Host,
	}

	switch opts.Channel {
	case model.AuthenticatorOOBChannelEmail:
		ctx.Email = opts.Target
	case model.AuthenticatorOOBChannelSMS:
		ctx.Phone = opts.Target
	case model.AuthenticatorOOBChannelWhatsapp:
		ctx.Phone = opts.Target
	default:
		panic("otp: unknown channel: " + opts.Channel)
	}

	if opts.AdditionalContext != nil {
		ctx.HasPassword = opts.AdditionalContext.HasPassword
	}

	return ctx, nil
}

func (s *MessageSender) selectMessage(form Form, typ translation.MessageType) *translation.MessageSpec {
	var spec *translation.MessageSpec
	switch typ {
	case translation.MessageTypeVerification:
		spec = translation.MessageVerification
	case translation.MessageTypeSetupPrimaryOOB:
		if form == FormLink {
			spec = translation.MessageSetupPrimaryLoginLink
		} else {
			spec = translation.MessageSetupPrimaryOOB
		}
	case translation.MessageTypeSetupSecondaryOOB:
		if form == FormLink {
			spec = translation.MessageSetupSecondaryLoginLink
		} else {
			spec = translation.MessageSetupSecondaryOOB
		}
	case translation.MessageTypeAuthenticatePrimaryOOB:
		if form == FormLink {
			spec = translation.MessageAuthenticatePrimaryLoginLink
		} else {
			spec = translation.MessageAuthenticatePrimaryOOB
		}
	case translation.MessageTypeAuthenticateSecondaryOOB:
		if form == FormLink {
			spec = translation.MessageAuthenticateSecondaryLoginLink
		} else {
			spec = translation.MessageAuthenticateSecondaryOOB
		}
	case translation.MessageTypeForgotPassword:
		if form == FormLink {
			spec = translation.MessageForgotPasswordLink
		} else {
			spec = translation.MessageForgotPasswordOOB
		}
	case translation.MessageTypeWhatsappCode:
		spec = translation.MessageWhatsappCode
	default:
		panic("otp: unknown message type: " + typ)
	}

	return spec
}

func (s *MessageSender) sendEmail(ctx context.Context, opts SendOptions) error {
	spec := s.selectMessage(opts.Form, opts.Type)
	msgType := spec.MessageType

	variables, err := s.setupTemplateContext(msgType, opts)
	if err != nil {
		return err
	}

	data, err := s.Translation.EmailMessageData(ctx, spec, variables)
	if err != nil {
		return err
	}

	mailSendOptions := &mail.SendOptions{
		Sender:    data.Sender,
		ReplyTo:   data.ReplyTo,
		Subject:   data.Subject,
		Recipient: opts.Target,
		TextBody:  data.TextBody.String,
		HTMLBody:  data.HTMLBody.String,
	}

	err = s.Sender.SendEmailInNewGoroutine(ctx, msgType, mailSendOptions)
	return err
}

func (s *MessageSender) sendSMS(ctx context.Context, opts SendOptions, preferAsync bool) error {
	spec := s.selectMessage(opts.Form, opts.Type)
	msgType := spec.MessageType

	variables, err := s.setupTemplateContext(msgType, opts)
	if err != nil {
		return err
	}

	data, err := s.Translation.SMSMessageData(ctx, spec, variables)
	if err != nil {
		return err
	}

	smsSendOptions := &sms.SendOptions{
		Sender:            data.Sender,
		To:                opts.Target,
		Body:              data.Body.String,
		AppID:             string(s.AppID),
		TemplateName:      filepath.Base(spec.SMSTemplate.Name),
		LanguageTag:       data.Body.LanguageTag,
		TemplateVariables: sms.NewTemplateVariablesFromPreparedTemplateVariables(data.PreparedTemplateVariables),
	}
	if preferAsync {
		err = s.Sender.SendSMSInNewGoroutine(ctx, msgType, smsSendOptions)
	} else {
		err = s.Sender.SendSMSImmediately(ctx, msgType, smsSendOptions)
	}
	return err
}

func (s *MessageSender) sendWhatsapp(ctx context.Context, opts SendOptions) (err error) {

	spec := s.selectMessage(opts.Form, opts.Type)
	msgType := spec.MessageType

	whatsappSendAuthenticationOTPOptions := &whatsapp.SendAuthenticationOTPOptions{
		To:  opts.Target,
		OTP: opts.OTP,
	}

	resultCallback := func(ctx context.Context, result *messaging.SendWhatsappResult) {
		_ = s.updateCodeAfterSent(ctx, opts, afterSentResult{
			WhatsappMessageID: result.MessageID,
			AwaitConfirmation: true,
		})
	}

	errorCallback := func(ctx context.Context, err error) {
		_ = s.updateCodeAfterSent(ctx, opts, afterSentResult{
			SendError: err,
		})
	}

	err = s.Sender.SendWhatsappInNewGoroutine(ctx, msgType, whatsappSendAuthenticationOTPOptions, resultCallback, errorCallback)
	return err
}

type afterSentResult struct {
	SendError         error
	WhatsappMessageID string
	// AwaitConfirmation indicates the provider reports the outcome asynchronously.
	AwaitConfirmation bool
}

func (s *MessageSender) updateCodeAfterSent(ctx context.Context, opts SendOptions, result afterSentResult) error {
	logger := SenderLogger.GetLogger(ctx)
	// Detach the deadline so that the context is not canceled along with the request.
	// Nothing other than this update ever records a delivery attempt, so losing it
	// leaves the code waiting to send until it expires.
	ctx = context.WithoutCancel(ctx)
	code, err := s.CodeStore.Get(ctx, opts.Kind.Purpose(), opts.Target)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to get code in result callback")
		return err
	}
	if code.SendMessageError != nil || code.InternalDeliveryStatus == OTPDeliveryStatusInternalFailed {
		// If it was error, ignore any later updates
		return nil
	}

	switch {
	case result.SendError != nil:
		code.SendMessageError = apierrors.AsAPIErrorWithContext(ctx, result.SendError)
		code.InternalDeliveryStatus = OTPDeliveryStatusInternalFailed
	case result.AwaitConfirmation:
		code.InternalDeliveryStatus = OTPDeliveryStatusInternalWaitingForConfirmation
	default:
		code.InternalDeliveryStatus = OTPDeliveryStatusInternalSent
	}

	if result.WhatsappMessageID != "" {
		code.WhatsappMessageID = result.WhatsappMessageID
	}
	// Still read by consumeCode and by deriveLegacyDeliveryStatus.
	code.OOBChannel = opts.Channel
	err = s.CodeStore.Update(ctx, opts.Kind.Purpose(), code)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to update code in result callback")
		return err
	}
	return nil
}

func (s *MessageSender) Send(ctx context.Context, opts SendOptions) error {
	return s.send(ctx, opts, false)
}

func (s *MessageSender) SendAsync(ctx context.Context, opts SendOptions) error {
	return s.send(ctx, opts, true)
}

func (s *MessageSender) send(ctx context.Context, opts SendOptions, preferAsync bool) (err error) {
	// A channel whose provider gains asynchronous reporting only has to set this;
	// nothing downstream looks at the channel.
	var awaitConfirmation bool

	// This records the delivery attempt synchronously, before the request returns,
	// even though the delivery itself may be asynchronous. Readers depend on that.
	defer func() {
		updateErr := s.updateCodeAfterSent(ctx, opts, afterSentResult{
			SendError:         err,
			AwaitConfirmation: awaitConfirmation,
		})
		if updateErr != nil {
			SenderLogger.GetLogger(ctx).WithError(updateErr).Error(ctx, "failed to update code after sent")
			err = errors.Join(err, updateErr)
		}
	}()

	switch opts.Channel {
	case model.AuthenticatorOOBChannelEmail:
		err = s.sendEmail(ctx, opts)
		return
	case model.AuthenticatorOOBChannelSMS:
		err = s.sendSMS(ctx, opts, preferAsync)
		return
	case model.AuthenticatorOOBChannelWhatsapp:
		// The outcome arrives later through the message status callback.
		awaitConfirmation = true
		err = s.sendWhatsapp(ctx, opts)
		return
	default:
		panic("otp: unknown channel: " + opts.Channel)
	}
}
