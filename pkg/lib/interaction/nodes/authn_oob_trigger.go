package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationOOBTrigger{})
}

type InputAuthenticationOOBTrigger interface {
	GetOOBAuthenticatorIndex() int
}

type EdgeAuthenticationOOBTrigger struct {
	Stage          interaction.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeAuthenticationOOBTrigger) getAuthenticator(idx int) (*authenticator.Info, error) {
	if idx < 0 || idx >= len(e.Authenticators) {
		return nil, authenticator.ErrAuthenticatorNotFound
	}

	return e.Authenticators[idx], nil
}

func (e *EdgeAuthenticationOOBTrigger) AuthenticatorType() authn.AuthenticatorType {
	return authn.AuthenticatorTypeOOB
}

func (e *EdgeAuthenticationOOBTrigger) IsDefaultAuthenticator() bool {
	filtered := filterAuthenticators(e.Authenticators, authenticator.KeepDefault)
	return len(filtered) > 0
}

func (e *EdgeAuthenticationOOBTrigger) GetOOBOTPTarget(idx int) string {
	info, err := e.getAuthenticator(idx)
	if err != nil {
		return ""
	}

	channel := authn.AuthenticatorOOBChannel(info.Claims[authenticator.AuthenticatorClaimOOBOTPChannelType].(string))
	var target string
	switch channel {
	case authn.AuthenticatorOOBChannelSMS:
		target = info.Claims[authenticator.AuthenticatorClaimOOBOTPPhone].(string)
	case authn.AuthenticatorOOBChannelEmail:
		target = info.Claims[authenticator.AuthenticatorClaimOOBOTPEmail].(string)
	}
	return target
}

func (e *EdgeAuthenticationOOBTrigger) GetOOBOTPChannel(idx int) string {
	info, err := e.getAuthenticator(idx)
	if err != nil {
		return ""
	}
	return info.Claims[authenticator.AuthenticatorClaimOOBOTPChannelType].(string)
}

func (e *EdgeAuthenticationOOBTrigger) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationOOBTrigger
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	idx := input.GetOOBAuthenticatorIndex()
	if idx < 0 || idx >= len(e.Authenticators) {
		return nil, authenticator.ErrAuthenticatorNotFound
	}
	targetInfo := e.Authenticators[idx]

	result, err := sendOOBCode(ctx, e.Stage, true, targetInfo)
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationOOBTrigger{
		Stage:         e.Stage,
		Authenticator: targetInfo,
		Target:        result.Target,
		Channel:       result.Channel,
		CodeLength:    result.CodeLength,
		SendCooldown:  result.SendCooldown,
	}, nil
}

type NodeAuthenticationOOBTrigger struct {
	Stage         interaction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info             `json:"authenticator"`
	Target        string                          `json:"target"`
	Channel       string                          `json:"channel"`
	CodeLength    int                             `json:"code_length"`
	SendCooldown  int                             `json:"send_cooldown"`
}

// GetOOBOTPTarget implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPTarget() string {
	return n.Target
}

// GetOOBOTPChannel implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPChannel() string {
	return n.Channel
}

// GetOOBOTPCodeSendCooldown implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPCodeSendCooldown() int {
	return n.SendCooldown
}

// GetOOBOTPCodeLength implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPCodeLength() int {
	return n.CodeLength
}

func (n *NodeAuthenticationOOBTrigger) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationOOBTrigger) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationOOBTrigger) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeOOBResendCode{
			Stage:            n.Stage,
			IsAuthenticating: true,
			Authenticator:    n.Authenticator,
		},
		&EdgeAuthenticationOOB{Stage: n.Stage, Authenticator: n.Authenticator},
	}, nil
}
