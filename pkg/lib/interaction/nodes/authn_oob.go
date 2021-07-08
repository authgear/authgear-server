package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationOOB{})
}

type InputAuthenticationOOB interface {
	GetOOBOTP() string
}

type EdgeAuthenticationOOB struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
	Secret        string
}

func (e *EdgeAuthenticationOOB) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationOOB
	if !interaction.AsInput(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	info := e.Authenticator
	err := ctx.Authenticators.VerifySecret(info, nil, input.GetOOBOTP())
	if err != nil {
		info = nil
	}

	return &NodeAuthenticationOOB{Stage: e.Stage, Authenticator: info, AuthenticatorType: e.Authenticator.Type}, nil
}

type NodeAuthenticationOOB struct {
	Stage             authn.AuthenticationStage `json:"stage"`
	AuthenticatorType authn.AuthenticatorType   `json:"authenticator_type"`
	Authenticator     *authenticator.Info       `json:"authenticator"`
}

func (n *NodeAuthenticationOOB) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationOOB) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationOOB) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	var typ authn.AuthenticationType
	switch n.AuthenticatorType {
	case authn.AuthenticatorTypeOOBEmail:
		typ = authn.AuthenticationTypeOOBOTPEmail
	case authn.AuthenticatorTypeOOBSMS:
		typ = authn.AuthenticationTypeOOBOTPSMS
	default:
		panic(fmt.Errorf("interaction: unexpected authenticator type: %v", n.AuthenticatorType))
	}

	return []interaction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			AuthenticationType:    typ,
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}
