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
	interaction.RegisterNode(&NodeAuthenticationLoginLinkTrigger{})
}

type InputAuthenticationLoginLinkTrigger interface {
	GetLoginLinkAuthenticatorIndex() int
}

type EdgeAuthenticationLoginLinkTrigger struct {
	Stage          authn.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeAuthenticationLoginLinkTrigger) getAuthenticator(idx int) (*authenticator.Info, error) {
	if idx < 0 || idx >= len(e.Authenticators) {
		return nil, authenticator.ErrAuthenticatorNotFound
	}

	return e.Authenticators[idx], nil
}

func (e *EdgeAuthenticationLoginLinkTrigger) GetTarget(idx int) string {
	info, err := e.getAuthenticator(idx)
	if err != nil {
		return ""
	}
	return info.OOBOTP.ToTarget()
}

func (e *EdgeAuthenticationLoginLinkTrigger) AuthenticatorType() model.AuthenticatorType {
	return model.AuthenticatorTypeOOBSMS
}

func (e *EdgeAuthenticationLoginLinkTrigger) IsDefaultAuthenticator() bool {
	filtered := authenticator.ApplyFilters(e.Authenticators, authenticator.KeepDefault)
	return len(filtered) > 0
}

func (e *EdgeAuthenticationLoginLinkTrigger) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationLoginLinkTrigger
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	idx := input.GetLoginLinkAuthenticatorIndex()
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
		OTPForm:              otp.FormLink,
	}).Do(goCtx)
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationLoginLinkTrigger{
		Stage:              e.Stage,
		Authenticator:      targetInfo,
		Authenticators:     e.Authenticators,
		AuthenticatorIndex: idx,
		Target:             result.Target,
		Channel:            result.Channel,
	}, nil
}

type NodeAuthenticationLoginLinkTrigger struct {
	Stage              authn.AuthenticationStage `json:"stage"`
	Authenticator      *authenticator.Info       `json:"authenticator"`
	Authenticators     []*authenticator.Info     `json:"authenticators"`
	AuthenticatorIndex int                       `json:"authenticator_index"`
	Channel            string                    `json:"channel"`
	Target             string                    `json:"target"`
}

// GetLoginLinkOTPTarget implements LoginLinkOTPNode.
func (n *NodeAuthenticationLoginLinkTrigger) GetLoginLinkOTPTarget() string {
	return n.Target
}

// GetLoginLinkOTPChannel implements LoginLinkOTPNode.
func (n *NodeAuthenticationLoginLinkTrigger) GetLoginLinkOTPChannel() string {
	return n.Channel
}

// GetLoginLinkOTPOOBType implements LoginLinkOTPNode.
func (n *NodeAuthenticationLoginLinkTrigger) GetLoginLinkOTPOOBType() interaction.OOBType {
	switch n.Stage {
	case authn.AuthenticationStagePrimary:
		return interaction.OOBTypeAuthenticatePrimary
	case authn.AuthenticationStageSecondary:
		return interaction.OOBTypeAuthenticateSecondary
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}
}

// GetAuthenticatorIndex implements LoginLinkOTPAuthnNode.
func (n *NodeAuthenticationLoginLinkTrigger) GetAuthenticatorIndex() int {
	return n.AuthenticatorIndex
}

func (n *NodeAuthenticationLoginLinkTrigger) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationLoginLinkTrigger) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationLoginLinkTrigger) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	edges := []interaction.Edge{
		&EdgeOOBResendCode{
			Stage:            n.Stage,
			IsAuthenticating: true,
			Authenticator:    n.Authenticator,
			OTPForm:          otp.FormLink,
		},
		&EdgeAuthenticationLoginLink{Stage: n.Stage, Authenticator: n.Authenticator},
	}
	return edges, nil
}
