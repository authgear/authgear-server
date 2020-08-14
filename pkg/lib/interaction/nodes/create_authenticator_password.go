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
	Stage interaction.AuthenticationStage
	Tag   []string
}

func (e *EdgeCreateAuthenticatorPassword) AuthenticatorType() authn.AuthenticatorType {
	return authn.AuthenticatorTypePassword
}

func (e *EdgeCreateAuthenticatorPassword) HasDefaultTag() bool {
	return false
}

func (e *EdgeCreateAuthenticatorPassword) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	input, ok := rawInput.(InputCreateAuthenticatorPassword)
	if !ok {
		return nil, interaction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	spec := &authenticator.Spec{
		UserID: userID,
		Tag:    stageToAuthenticatorTag(e.Stage),
		Type:   authn.AuthenticatorTypePassword,
		Props:  map[string]interface{}{},
	}
	spec.Tag = append(spec.Tag, e.Tag...)

	info, err := ctx.Authenticators.New(spec, input.GetPassword())
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorPassword{Stage: e.Stage, Authenticator: info}, nil
}

type NodeCreateAuthenticatorPassword struct {
	Stage         interaction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info             `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorPassword) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorPassword) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorPassword) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
