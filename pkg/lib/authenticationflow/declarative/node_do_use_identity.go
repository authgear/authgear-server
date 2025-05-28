package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
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

func NewNodeDoUseIdentity(ctx context.Context, flows authflow.Flows, n *NodeDoUseIdentity) (*NodeDoUseIdentity, error) {
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

var _ authflow.NodeSimple = &NodeDoUseIdentity{}
var _ authflow.Milestone = &NodeDoUseIdentity{}
var _ MilestoneDoUseUser = &NodeDoUseIdentity{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentity{}

func (*NodeDoUseIdentity) Kind() string {
	return "NodeDoUseIdentity"
}

func (*NodeDoUseIdentity) Milestone() {}
func (n *NodeDoUseIdentity) MilestoneDoUseUser() string {
	return n.Identity.UserID
}
func (n *NodeDoUseIdentity) MilestoneDoUseIdentity() *identity.Info { return n.Identity }
