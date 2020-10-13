package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoCreateAuthenticator{})
}

type EdgeDoCreateAuthenticator struct {
	Stage          interaction.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeDoCreateAuthenticator) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoCreateAuthenticator{
		Stage:          e.Stage,
		Authenticators: e.Authenticators,
	}, nil
}

type NodeDoCreateAuthenticator struct {
	Stage          interaction.AuthenticationStage `json:"stage"`
	Authenticators []*authenticator.Info           `json:"authenticators"`
}

func (n *NodeDoCreateAuthenticator) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoCreateAuthenticator) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	err := perform(interaction.EffectRun(func(ctx *interaction.Context) error {
		for _, a := range n.Authenticators {
			err := ctx.Authenticators.Create(a)
			if errors.Is(err, authenticator.ErrAuthenticatorAlreadyExists) {
				return interaction.ErrDuplicatedAuthenticator
			} else if err != nil {
				return err
			}
		}

		return nil
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoCreateAuthenticator) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoCreateAuthenticator) UserAuthenticator(stage interaction.AuthenticationStage) (*authenticator.Info, bool) {
	if len(n.Authenticators) > 1 {
		panic("interaction: expect at most one primary/secondary authenticator")
	}
	if len(n.Authenticators) == 0 {
		return nil, false
	}
	if n.Stage == stage && n.Authenticators[0] != nil {
		return n.Authenticators[0], true
	}
	return nil, false
}

func (n *NodeDoCreateAuthenticator) UserNewAuthenticators() []*authenticator.Info {
	return n.Authenticators
}
