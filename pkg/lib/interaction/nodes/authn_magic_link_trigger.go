package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationMagicLinkTrigger{})
}

type InputAuthenticationMagicLinkTrigger interface {
	GetMagicLinkAuthenticatorIndex() int
}

type EdgeAuthenticationMagicLinkTrigger struct {
	Stage          authn.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeAuthenticationMagicLinkTrigger) getAuthenticator(idx int) (*authenticator.Info, error) {
	if idx < 0 || idx >= len(e.Authenticators) {
		return nil, authenticator.ErrAuthenticatorNotFound
	}

	return e.Authenticators[idx], nil
}

func (e *EdgeAuthenticationMagicLinkTrigger) GetTarget(idx int) string {
	info, err := e.getAuthenticator(idx)
	if err != nil {
		return ""
	}
	return info.OOBOTP.ToTarget()
}

func (e *EdgeAuthenticationMagicLinkTrigger) AuthenticatorType() model.AuthenticatorType {
	return model.AuthenticatorTypeOOBSMS
}

func (e *EdgeAuthenticationMagicLinkTrigger) IsDefaultAuthenticator() bool {
	filtered := authenticator.ApplyFilters(e.Authenticators, authenticator.KeepDefault)
	return len(filtered) > 0
}

func (e *EdgeAuthenticationMagicLinkTrigger) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationMagicLinkTrigger
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	idx := input.GetMagicLinkAuthenticatorIndex()
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
		OTPMode:              otp.OTPModeMagicLink,
	}).Do()
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationMagicLinkTrigger{
		Stage:              e.Stage,
		Authenticator:      targetInfo,
		Authenticators:     e.Authenticators,
		AuthenticatorIndex: idx,
		MagicLinkOTP:       result.Code,
		Target:             result.Target,
	}, nil
}

type NodeAuthenticationMagicLinkTrigger struct {
	Stage              authn.AuthenticationStage `json:"stage"`
	Authenticator      *authenticator.Info       `json:"authenticator"`
	Authenticators     []*authenticator.Info     `json:"authenticators"`
	AuthenticatorIndex int                       `json:"authenticator_index"`
	MagicLinkOTP       string                    `json:"magic_link_otp"`
	Target             string                    `json:"target"`
}

// GetMagicLinkOTP implements MagicLinkOTPNode.
func (n *NodeAuthenticationMagicLinkTrigger) GetMagicLinkOTP() string {
	return n.MagicLinkOTP
}

// GetPhone implements MagicLinkOTPNode.
func (n *NodeAuthenticationMagicLinkTrigger) GetMagicLinkOTPTarget() string {
	return n.Target
}

// GetAuthenticatorIndex implements MagicLinkOTPAuthnNode.
func (n *NodeAuthenticationMagicLinkTrigger) GetAuthenticatorIndex() int {
	return n.AuthenticatorIndex
}

func (n *NodeAuthenticationMagicLinkTrigger) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationMagicLinkTrigger) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationMagicLinkTrigger) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	edges := []interaction.Edge{
		&EdgeAuthenticationMagicLink{Stage: n.Stage, Authenticator: n.Authenticator},
	}
	return edges, nil
}
