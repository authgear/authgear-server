package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationOOBTrigger{})
}

type InputAuthenticationOOBTrigger interface {
	GetOOBAuthenticatorType() string
	GetOOBAuthenticatorIndex() int
}

type EdgeAuthenticationOOBTrigger struct {
	Stage                authn.AuthenticationStage
	OOBAuthenticatorType model.AuthenticatorType
	Authenticators       []*authenticator.Info
}

func (e *EdgeAuthenticationOOBTrigger) getAuthenticator(idx int) (*authenticator.Info, error) {
	if idx < 0 || idx >= len(e.Authenticators) {
		return nil, authenticator.ErrAuthenticatorNotFound
	}

	return e.Authenticators[idx], nil
}

func (e *EdgeAuthenticationOOBTrigger) AuthenticatorType() model.AuthenticatorType {
	return e.OOBAuthenticatorType
}

func (e *EdgeAuthenticationOOBTrigger) IsDefaultAuthenticator() bool {
	filtered := authenticator.ApplyFilters(e.Authenticators, authenticator.KeepDefault)
	return len(filtered) > 0
}

func (e *EdgeAuthenticationOOBTrigger) GetOOBOTPTarget(idx int) string {
	info, err := e.getAuthenticator(idx)
	if err != nil {
		return ""
	}

	var target string
	switch info.Type {
	case model.AuthenticatorTypeOOBSMS:
		target = info.OOBOTP.Phone
	case model.AuthenticatorTypeOOBEmail:
		target = info.OOBOTP.Email
	default:
		panic("interaction: incompatible authenticator type for oob: " + info.Type)
	}
	return target
}

func (e *EdgeAuthenticationOOBTrigger) GetOOBOTPChannel(idx int) model.AuthenticatorOOBChannel {
	info, err := e.getAuthenticator(idx)
	if err != nil {
		return ""
	}
	switch info.Type {
	case model.AuthenticatorTypeOOBSMS:
		return model.AuthenticatorOOBChannelSMS
	case model.AuthenticatorTypeOOBEmail:
		return model.AuthenticatorOOBChannelEmail
	default:
		panic("interaction: incompatible authenticator type for oob: " + info.Type)
	}
}

func (e *EdgeAuthenticationOOBTrigger) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationOOBTrigger
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	// It is possible that to have multiple EdgeAuthenticationOOBTrigger at the
	// same time (e.g. email or sms), check the OOBAuthenticatorType in input
	// to determine which authenticator we want to trigger
	oobAuthenticatorType := input.GetOOBAuthenticatorType()
	if model.AuthenticatorType(oobAuthenticatorType) != e.OOBAuthenticatorType {
		return nil, interaction.ErrIncompatibleInput
	}

	idx := input.GetOOBAuthenticatorIndex()
	if idx < 0 || idx >= len(e.Authenticators) {
		return nil, authenticator.ErrAuthenticatorNotFound
	}
	targetInfo := e.Authenticators[idx]

	result, err := (&SendOOBCode{
		Context:              ctx,
		Stage:                e.Stage,
		IsAuthenticating:     true,
		AuthenticatorInfo:    targetInfo,
		IgnoreRatelimitError: true,
		OTPForm:              otp.FormCode,
	}).Do(goCtx)
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationOOBTrigger{
		Stage:         e.Stage,
		Authenticator: targetInfo,
		Target:        result.Target,
		Channel:       result.Channel,
		CodeLength:    result.CodeLength,
	}, nil
}

type NodeAuthenticationOOBTrigger struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
	Target        string                    `json:"target"`
	Channel       string                    `json:"channel"`
	CodeLength    int                       `json:"code_length"`
}

// GetOOBOTPTarget implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPTarget() string {
	return n.Target
}

// GetOOBOTPChannel implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPChannel() string {
	return n.Channel
}

// GetSelectedPhoneNumberForPhoneOTPAuthentication implements AuthenticationPhoneOTPTriggerNode
func (n *NodeAuthenticationOOBTrigger) GetSelectedPhoneNumberForPhoneOTPAuthentication() string {
	if n.Channel == string(model.AuthenticatorOOBChannelSMS) {
		return n.Target
	}
	return ""
}

// GetOOBOTPOOBType implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPOOBType() interaction.OOBType {
	switch n.Stage {
	case authn.AuthenticationStagePrimary:
		return interaction.OOBTypeAuthenticatePrimary
	case authn.AuthenticationStageSecondary:
		return interaction.OOBTypeAuthenticateSecondary
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}
}

// GetOOBOTPCodeLength implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPCodeLength() int {
	return n.CodeLength
}

func (n *NodeAuthenticationOOBTrigger) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationOOBTrigger) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationOOBTrigger) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeOOBResendCode{
			Stage:            n.Stage,
			IsAuthenticating: true,
			Authenticator:    n.Authenticator,
			OTPForm:          otp.FormCode,
		},
		&EdgeAuthenticationOOB{Stage: n.Stage, Authenticator: n.Authenticator},
	}, nil
}
