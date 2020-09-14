package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeEnsureVerificationBegin{})
	interaction.RegisterNode(&NodeEnsureVerificationEnd{})
}

type EdgeEnsureVerificationBegin struct {
	Identity        *identity.Info
	RequestedByUser bool
}

func (e *EdgeEnsureVerificationBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeEnsureVerificationBegin{
		Identity:        e.Identity,
		RequestedByUser: e.RequestedByUser,
	}, nil
}

type NodeEnsureVerificationBegin struct {
	Identity           *identity.Info      `json:"identity"`
	RequestedByUser    bool                `json:"requested_by_user"`
	VerificationStatus verification.Status `json:"-"`
}

func (n *NodeEnsureVerificationBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	claims, err := ctx.Verification.GetIdentityVerificationStatus(n.Identity)
	if err != nil {
		return err
	}

	// TODO(verification): handle multiple verifiable claims per identity
	if len(claims) > 0 {
		n.VerificationStatus = claims[0].Status
	} else {
		n.VerificationStatus = verification.StatusDisabled
	}
	return nil
}

func (n *NodeEnsureVerificationBegin) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeEnsureVerificationBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	switch n.VerificationStatus {
	case verification.StatusDisabled, verification.StatusVerified:
		break
	case verification.StatusPending:
		if n.RequestedByUser {
			return []interaction.Edge{&EdgeVerifyIdentity{Identity: n.Identity}}, nil
		}
	case verification.StatusRequired:
		return []interaction.Edge{&EdgeVerifyIdentity{Identity: n.Identity}}, nil
	}

	return []interaction.Edge{&EdgeEnsureVerificationEnd{Identity: n.Identity}}, nil
}

type EdgeEnsureVerificationEnd struct {
	Identity *identity.Info
}

func (e *EdgeEnsureVerificationEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeEnsureVerificationEnd{
		Identity: e.Identity,
	}, nil
}

type NodeEnsureVerificationEnd struct {
	Identity         *identity.Info      `json:"identity"`
	NewVerifiedClaim *verification.Claim `json:"new_verified_claim,omitempty"`
}

func (n *NodeEnsureVerificationEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeEnsureVerificationEnd) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeEnsureVerificationEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
