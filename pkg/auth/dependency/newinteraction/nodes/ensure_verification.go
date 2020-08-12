package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
)

func init() {
	newinteraction.RegisterNode(&NodeEnsureVerificationBegin{})
	newinteraction.RegisterNode(&NodeEnsureVerificationEnd{})
}

type EdgeEnsureVerificationBegin struct {
	Identity        *identity.Info
	RequestedByUser bool
}

func (e *EdgeEnsureVerificationBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
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

func (n *NodeEnsureVerificationBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	status, err := ctx.Verification.GetVerificationStatus(n.Identity)
	if err != nil {
		return err
	}

	n.VerificationStatus = status
	return nil
}

func (n *NodeEnsureVerificationBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeEnsureVerificationBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	switch n.VerificationStatus {
	case verification.StatusDisabled, verification.StatusVerified:
		break
	case verification.StatusPending:
		if n.RequestedByUser {
			return []newinteraction.Edge{&EdgeVerifyIdentity{Identity: n.Identity}}, nil
		}
	case verification.StatusRequired:
		return []newinteraction.Edge{&EdgeVerifyIdentity{Identity: n.Identity}}, nil
	}

	return []newinteraction.Edge{&EdgeEnsureVerificationEnd{Identity: n.Identity}}, nil
}

type EdgeEnsureVerificationEnd struct {
	Identity *identity.Info
}

func (e *EdgeEnsureVerificationEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeEnsureVerificationEnd{
		Identity: e.Identity,
	}, nil
}

type NodeEnsureVerificationEnd struct {
	Identity         *identity.Info      `json:"identity"`
	NewAuthenticator *authenticator.Info `json:"new_authenticator,omitempty"`
}

func (n *NodeEnsureVerificationEnd) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeEnsureVerificationEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeEnsureVerificationEnd) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
