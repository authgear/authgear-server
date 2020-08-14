package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateAuthenticatorPassword{})
}

type InputCreateAuthenticatorPassword interface {
	GetPassword() string
}

type EdgeCreateAuthenticatorPassword struct {
	Stage newinteraction.AuthenticationStage
	Tag   []string
}

func (e *EdgeCreateAuthenticatorPassword) AuthenticatorType() authn.AuthenticatorType {
	return authn.AuthenticatorTypePassword
}

func (e *EdgeCreateAuthenticatorPassword) HasDefaultTag() bool {
	return false
}

func (e *EdgeCreateAuthenticatorPassword) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputCreateAuthenticatorPassword)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
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
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorPassword) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorPassword) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorPassword) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
