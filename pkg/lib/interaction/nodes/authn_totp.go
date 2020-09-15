package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationTOTP{})
}

type InputAuthenticationTOTP interface {
	GetTOTP() string
}

type EdgeAuthenticationTOTP struct {
	Stage          interaction.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeAuthenticationTOTP) AuthenticatorType() authn.AuthenticatorType {
	return authn.AuthenticatorTypeTOTP
}

func (e *EdgeAuthenticationTOTP) IsDefaultAuthenticator() bool {
	filtered := filterAuthenticators(e.Authenticators, authenticator.KeepDefault)
	return len(filtered) > 0
}

func (e *EdgeAuthenticationTOTP) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	input, ok := rawInput.(InputAuthenticationTOTP)
	if !ok {
		return nil, interaction.ErrIncompatibleInput
	}

	inputTOTP := input.GetTOTP()

	var info *authenticator.Info
	for _, a := range e.Authenticators {
		err := ctx.Authenticators.VerifySecret(a, nil, inputTOTP)
		if errors.Is(err, authenticator.ErrInvalidCredentials) {
			continue
		} else if err != nil {
			return nil, err
		} else {
			aa := a
			info = aa
		}
	}

	return &NodeAuthenticationTOTP{Stage: e.Stage, Authenticator: info}, nil
}

type NodeAuthenticationTOTP struct {
	Stage         interaction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info             `json:"authenticator"`
}

func (n *NodeAuthenticationTOTP) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationTOTP) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationTOTP) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}
