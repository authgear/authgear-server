package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateIdentityBegin{})
}

type EdgeCreateIdentityBegin struct {
	AllowAnonymousUser bool
}

func (e *EdgeCreateIdentityBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeCreateIdentityBegin{AllowAnonymousUser: e.AllowAnonymousUser}, nil
}

type NodeCreateIdentityBegin struct {
	AllowAnonymousUser bool                   `json:"allow_anonymous_user"`
	IdentityTypes      []authn.IdentityType   `json:"-"`
	IdentityConfig     *config.IdentityConfig `json:"-"`
}

func (n *NodeCreateIdentityBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	n.IdentityTypes = ctx.Config.Authentication.Identities
	n.IdentityConfig = ctx.Config.Identity
	return nil
}

func (n *NodeCreateIdentityBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateIdentityBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return n.deriveEdges(), nil
}

func (n *NodeCreateIdentityBegin) deriveEdges() []newinteraction.Edge {
	var edges []newinteraction.Edge
	for _, t := range n.IdentityTypes {
		switch t {
		case authn.IdentityTypeAnonymous:
			if n.AllowAnonymousUser {
				edges = append(edges, &EdgeUseIdentityAnonymous{
					IsCreating: true,
				})
			}

		case authn.IdentityTypeLoginID:
			edges = append(edges, &EdgeUseIdentityLoginID{
				Mode:    UseIdentityLoginIDModeCreate,
				Configs: n.IdentityConfig.LoginID.Keys,
			})

		case authn.IdentityTypeOAuth:
			edges = append(edges, &EdgeUseIdentityOAuthProvider{
				IsCreating: true,
				Configs:    n.IdentityConfig.OAuth.Providers,
			})

		default:
			panic("interaction: unknown identity type: " + t)
		}
	}

	return edges
}

func (n *NodeCreateIdentityBegin) GetIdentityCandidates() []identity.Candidate {
	var candidates []identity.Candidate
	for _, e := range n.deriveEdges() {
		if e, ok := e.(interface{ GetIdentityCandidates() []identity.Candidate }); ok {
			candidates = append(candidates, e.GetIdentityCandidates()...)
		}
	}
	return candidates
}
