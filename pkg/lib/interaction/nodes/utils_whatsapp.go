package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type SendWhatsappCodeResult struct {
	Target     string
	CodeLength int
	Kind       otp.Kind
}

type SendWhatsappCode struct {
	KindFactory func(config *config.AppConfig, channel model.AuthenticatorOOBChannel) otp.Kind
	Context     *interaction.Context
	Target      string
	IsResend    bool
}

func (s *SendWhatsappCode) Do(goCtx context.Context) (*SendWhatsappCodeResult, error) {
	channel := model.AuthenticatorOOBChannelWhatsapp
	form := otp.FormCode

	kind := s.KindFactory(s.Context.Config, channel)
	result := &SendWhatsappCodeResult{
		Target:     s.Target,
		CodeLength: form.CodeLength(),
		Kind:       kind,
	}

	code, err := s.Context.OTPCodeService.GenerateOTP(
		goCtx,
		kind,
		s.Target,
		form,
		&otp.GenerateOptions{WebSessionID: s.Context.WebSessionID},
	)
	if !s.IsResend && apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
		return nil, err
	}

	err = s.Context.OTPSender.Send(
		goCtx,
		otp.SendOptions{
			Channel: channel,
			Target:  s.Target,
			Form:    form,
			Type:    translation.MessageTypeWhatsappCode,
			OTP:     code,
		},
	)
	if !s.IsResend && apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
		return nil, err
	}

	return result, nil
}

func NewSendWhatsappCode(
	ctx *interaction.Context,
	kindFactory func(config *config.AppConfig, channel model.AuthenticatorOOBChannel) otp.Kind,
	target string,
	isResend bool) *SendWhatsappCode {
	return &SendWhatsappCode{
		Context:     ctx,
		KindFactory: kindFactory,
		Target:      target,
		IsResend:    isResend,
	}
}
