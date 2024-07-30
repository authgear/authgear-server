package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type InputWhatsappOTPResendCode interface {
	DoResend()
}

type EdgeWhatsappOTPResendCode struct {
	Target         string
	OTPKindFactory otp.DeprecatedKindFactory
}

func (e *EdgeWhatsappOTPResendCode) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputWhatsappOTPResendCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	_, err := NewSendWhatsappCode(ctx, e.OTPKindFactory, e.Target, true).Do()
	if err != nil {
		return nil, err
	}

	return nil, interaction.ErrSameNode
}
