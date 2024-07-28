package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationWhatsappTrigger{})
}

type InputAuthenticationWhatsappTrigger interface {
	GetWhatsappAuthenticatorIndex() int
}

type EdgeAuthenticationWhatsappTrigger struct {
	Stage          authn.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeAuthenticationWhatsappTrigger) getAuthenticator(idx int) (*authenticator.Info, error) {
	if idx < 0 || idx >= len(e.Authenticators) {
		return nil, authenticator.ErrAuthenticatorNotFound
	}

	return e.Authenticators[idx], nil
}

func (e *EdgeAuthenticationWhatsappTrigger) GetPhone(idx int) string {
	info, err := e.getAuthenticator(idx)
	if err != nil {
		return ""
	}
	return info.OOBOTP.Phone
}

func (e *EdgeAuthenticationWhatsappTrigger) AuthenticatorType() model.AuthenticatorType {
	return model.AuthenticatorTypeOOBSMS
}

func (e *EdgeAuthenticationWhatsappTrigger) IsDefaultAuthenticator() bool {
	filtered := authenticator.ApplyFilters(e.Authenticators, authenticator.KeepDefault)
	return len(filtered) > 0
}

func (e *EdgeAuthenticationWhatsappTrigger) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationWhatsappTrigger
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	idx := input.GetWhatsappAuthenticatorIndex()
	if idx < 0 || idx >= len(e.Authenticators) {
		return nil, authenticator.ErrAuthenticatorNotFound
	}
	targetInfo := e.Authenticators[idx]
	phone := targetInfo.OOBOTP.Phone
	result, err := NewSendWhatsappCode(ctx, otp.KindOOBOTPCode, phone, false).Do()
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationWhatsappTrigger{
		Stage:              e.Stage,
		Authenticator:      targetInfo,
		Authenticators:     e.Authenticators,
		AuthenticatorIndex: idx,
		WhatsappOTPLength:  result.CodeLength,
		Phone:              phone,
	}, nil
}

type NodeAuthenticationWhatsappTrigger struct {
	Stage              authn.AuthenticationStage `json:"stage"`
	Authenticator      *authenticator.Info       `json:"authenticator"`
	Authenticators     []*authenticator.Info     `json:"authenticators"`
	AuthenticatorIndex int                       `json:"authenticator_index"`
	WhatsappOTPLength  int                       `json:"whatsapp_otp_length"`
	Phone              string                    `json:"phone"`
}

// GetWhatsappOTP implements WhatsappOTPNode.
func (n *NodeAuthenticationWhatsappTrigger) GetWhatsappOTPLength() int {
	return n.WhatsappOTPLength
}

// GetPhone implements WhatsappOTPNode.
func (n *NodeAuthenticationWhatsappTrigger) GetPhone() string {
	return n.Phone
}

// GetOTPKindFactory implements WhatsappOTPNode.
func (n *NodeAuthenticationWhatsappTrigger) GetOTPKindFactory() otp.DeprecatedKindFactory {
	return otp.KindOOBOTPCode
}

// GetSelectedPhoneNumberForPhoneOTPAuthentication implements AuthenticationPhoneOTPTriggerNode
func (n *NodeAuthenticationWhatsappTrigger) GetSelectedPhoneNumberForPhoneOTPAuthentication() string {
	return n.Phone
}

// GetAuthenticatorIndex implements WhatsappOTPAuthnNode.
func (n *NodeAuthenticationWhatsappTrigger) GetAuthenticatorIndex() int {
	return n.AuthenticatorIndex
}

func (n *NodeAuthenticationWhatsappTrigger) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationWhatsappTrigger) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationWhatsappTrigger) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	edges := []interaction.Edge{
		&EdgeWhatsappOTPResendCode{
			Target:         n.Phone,
			OTPKindFactory: n.GetOTPKindFactory(),
		},
		&EdgeAuthenticationWhatsapp{Stage: n.Stage, Authenticator: n.Authenticator},
	}
	return edges, nil
}
