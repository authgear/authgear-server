package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateIdentityBegin{})
}

type InputCreateIdentityBegin interface {
}

type EdgeCreateIdentityBegin struct {
	RequestedIdentity *identity.Spec
}

func (e *EdgeCreateIdentityBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeCreateIdentityBegin{
		RequestedIdentity: e.RequestedIdentity,
	}, nil
}

type NodeCreateIdentityBegin struct {
	RequestedIdentity *identity.Spec `json:"request_identity"`
}

func (n *NodeCreateIdentityBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateIdentityBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	for _, t := range ctx.Config.Authentication.Identities {
		if n.RequestedIdentity != nil && n.RequestedIdentity.Type != t {
			continue
		}

		switch t {
		case authn.IdentityTypeAnonymous:
			panic("TODO(interaction): handle anonymous signup")

		case authn.IdentityTypeLoginID:
			for _, c := range ctx.Config.Identity.LoginID.Keys {
				if n.RequestedIdentity != nil &&
					n.RequestedIdentity.Claims[identity.IdentityClaimLoginIDKey] != c.Key {
					continue
				}
				edges = append(edges, &EdgeCreateIdentityLoginID{Config: c, RequestedIdentity: n.RequestedIdentity})
			}

		case authn.IdentityTypeOAuth:
			id := config.ProviderID{}
			if n.RequestedIdentity != nil {
				id = config.NewProviderID(n.RequestedIdentity.Claims[identity.IdentityClaimOAuthProviderKeys].(map[string]interface{}))
			}
			for _, c := range ctx.Config.Identity.OAuth.Providers {
				if n.RequestedIdentity != nil && !c.ProviderID().Equal(&id) {
					continue
				}
				edges = append(edges, &EdgeCreateIdentityOAuth{Config: c, RequestedIdentity: n.RequestedIdentity})
			}

		default:
			panic("interaction: unknown identity type: " + t)
		}
	}

	return edges, nil
}
