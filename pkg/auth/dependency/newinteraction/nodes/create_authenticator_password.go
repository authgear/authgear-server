package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateAuthenticatorPassword{})
}

type InputCreateAuthenticatorPassword interface {
	GetPassword() string
}

type EdgeCreateAuthenticatorPassword struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeCreateAuthenticatorPassword) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputCreateAuthenticatorPassword)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	spec := &authenticator.Spec{
		UserID: userID,
		Type:   authn.AuthenticatorTypePassword,
		Props:  map[string]interface{}{},
	}

	info, err := ctx.Authenticators.New(spec, input.GetPassword())
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorPassword{Stage: e.Stage, Authenticator: info}, nil
}

type NodeCreateAuthenticatorPassword struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorPassword) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorPassword) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
