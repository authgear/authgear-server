package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorPassword{})
}

type InputCreateAuthenticatorPassword interface {
	GetPassword() string
}

type EdgeCreateAuthenticatorPassword struct {
	Stage     authn.AuthenticationStage
	IsDefault bool
}

func (e *EdgeCreateAuthenticatorPassword) AuthenticatorType() authn.AuthenticatorType {
	return authn.AuthenticatorTypePassword
}

func (e *EdgeCreateAuthenticatorPassword) IsDefaultAuthenticator() bool {
	return false
}

func (e *EdgeCreateAuthenticatorPassword) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var stageInput InputAuthenticationStage
	if !interaction.AsInput(rawInput, &stageInput) {
		return nil, interaction.ErrIncompatibleInput
	}
	stage := stageInput.GetAuthenticationStage()
	if stage != e.Stage {
		return nil, interaction.ErrIncompatibleInput
	}

	var input InputCreateAuthenticatorPassword
	if !interaction.AsInput(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	spec := &authenticator.Spec{
		UserID:    userID,
		IsDefault: e.IsDefault,
		Kind:      stageToAuthenticatorKind(e.Stage),
		Type:      authn.AuthenticatorTypePassword,
		Claims:    map[string]interface{}{},
	}

	info, err := ctx.Authenticators.New(spec, input.GetPassword())
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorPassword{Stage: e.Stage, Authenticator: info}, nil
}

type NodeCreateAuthenticatorPassword struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorPassword) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorPassword) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorPassword) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
