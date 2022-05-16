package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
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
	phone := targetInfo.Claims[authenticator.AuthenticatorClaimOOBOTPPhone].(string)
	code, err := ctx.WhatsappCodeProvider.CreateCode(phone, string(ctx.Config.ID), ctx.WebSessionID)
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationWhatsappTrigger{
		Stage:              e.Stage,
		Authenticator:      targetInfo,
		Authenticators:     e.Authenticators,
		AuthenticatorIndex: idx,
		WhatsappOTP:        code.Code,
		Phone:              phone,
		PhoneOTPMode:       ctx.Config.Authenticator.OOB.SMS.PhoneOTPMode,
	}, nil
}

type NodeAuthenticationWhatsappTrigger struct {
	Stage              authn.AuthenticationStage        `json:"stage"`
	Authenticator      *authenticator.Info              `json:"authenticator"`
	Authenticators     []*authenticator.Info            `json:"authenticators"`
	AuthenticatorIndex int                              `json:"authenticator_index"`
	WhatsappOTP        string                           `json:"whatsapp_otp"`
	Phone              string                           `json:"phone"`
	PhoneOTPMode       config.AuthenticatorPhoneOTPMode `json:"phone_otp_mode"`
}

// GetPhoneOTPMode implements WhatsappOTPNode.
func (n *NodeAuthenticationWhatsappTrigger) GetPhoneOTPMode() config.AuthenticatorPhoneOTPMode {
	return n.PhoneOTPMode
}

// GetWhatsappOTP implements WhatsappOTPNode.
func (n *NodeAuthenticationWhatsappTrigger) GetWhatsappOTP() string {
	return n.WhatsappOTP
}

// GetPhone implements WhatsappOTPNode.
func (n *NodeAuthenticationWhatsappTrigger) GetPhone() string {
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
		&EdgeAuthenticationWhatsapp{Stage: n.Stage, Authenticator: n.Authenticator},
	}
	if n.PhoneOTPMode == config.AuthenticatorPhoneOTPModeWhatsappSMS {
		// sms fallback
		edges = append(edges, &EdgeAuthenticationWhatsappFallbackSMS{
			Stage:          n.Stage,
			Authenticators: n.Authenticators,
		})
	}
	return edges, nil
}
