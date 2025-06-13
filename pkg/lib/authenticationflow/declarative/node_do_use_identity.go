package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIdentity{})
}

type NodeDoUseIdentity struct {
	Identity     *identity.Info `json:"identity,omitempty"`
	IdentitySpec *identity.Spec `json:"identity_spec,omitempty"`
}

func NewNodeDoUseIdentity(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDoUseIdentity) (*NodeDoUseIdentity, error) {
	userID, err := getUserID(flows)
	if errors.Is(err, ErrNoUserID) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	if userID != "" && userID != n.Identity.UserID {
		return nil, ErrDifferentUserID
	}

	if userIDHint := authflow.GetUserIDHint(ctx); userIDHint != "" {
		if userIDHint != n.Identity.UserID {
			return nil, api.ErrMismatchedUser
		}
	}

	return n, nil

}

func NewNodeDoUseIdentityReactToResult(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDoUseIdentity) (authenticationflow.ReactToResult, error) {
	n, err := NewNodeDoUseIdentity(ctx, deps, flows, n)
	if err != nil {
		return nil, err
	}

	return authenticationflow.NewNodeSimple(n), nil
}

var _ authflow.NodeSimple = &NodeDoUseIdentity{}
var _ authflow.Milestone = &NodeDoUseIdentity{}
var _ authflow.InputReactor = &NodeDoUseIdentity{}
var _ MilestoneDoUseUser = &NodeDoUseIdentity{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentity{}
var _ MilestoneGetIdentitySpecs = &NodeDoUseIdentity{}

func (*NodeDoUseIdentity) Kind() string {
	return "NodeDoUseIdentity"
}

func (n *NodeDoUseIdentity) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	return nil, nil
}

func (n *NodeDoUseIdentity) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	idmodel := n.Identity.ToModel()
	return NewNodePostIdentified(ctx, deps, flows, &NodePostIdentifiedOptions{
		Identity:       &idmodel,
		IDToken:        nil,
		Identification: n.Identity.ToIdentification(),
	})
}

func (*NodeDoUseIdentity) Milestone() {}
func (n *NodeDoUseIdentity) MilestoneDoUseUser() string {
	return n.Identity.UserID
}
func (n *NodeDoUseIdentity) MilestoneDoUseIdentity() *identity.Info { return n.Identity }

func (n *NodeDoUseIdentity) MilestoneGetIdentitySpecs() []*identity.Spec {
	return []*identity.Spec{n.IdentitySpec}
}
