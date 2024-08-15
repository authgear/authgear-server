package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIdentityWithUpdate{})
}

type NodeDoUseIdentityWithUpdate struct {
	OldIdentityInfo *identity.Info `json:"old_identity_info,omitempty"`
	NewIdentityInfo *identity.Info `json:"new_identity_info,omitempty"`
}

func NewNodeDoUseIdentityWithUpdate(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, oldIdentityInfo *identity.Info, spec *identity.Spec) (*NodeDoUseIdentityWithUpdate, error) {
	userID, err := getUserID(flows)
	if errors.Is(err, ErrNoUserID) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	if userID != "" && userID != oldIdentityInfo.UserID {
		return nil, ErrDifferentUserID
	}

	if userIDHint := authflow.GetUserIDHint(ctx); userIDHint != "" {
		if userIDHint != oldIdentityInfo.UserID {
			return nil, api.ErrMismatchedUser
		}
	}

	newIdentityInfo, err := deps.Identities.UpdateWithSpec(oldIdentityInfo, spec, identity.NewIdentityOptions{})
	if err != nil {
		return nil, err
	}

	return &NodeDoUseIdentityWithUpdate{
		OldIdentityInfo: oldIdentityInfo,
		NewIdentityInfo: newIdentityInfo,
	}, nil
}

var _ authflow.NodeSimple = &NodeDoUseIdentityWithUpdate{}
var _ authflow.EffectGetter = &NodeDoUseIdentityWithUpdate{}
var _ authflow.Milestone = &NodeDoUseIdentityWithUpdate{}
var _ MilestoneDoUseUser = &NodeDoUseIdentityWithUpdate{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentityWithUpdate{}

func (*NodeDoUseIdentityWithUpdate) Kind() string {
	return "NodeDoUseIdentityWithUpdate"
}

func (*NodeDoUseIdentityWithUpdate) Milestone() {}
func (n *NodeDoUseIdentityWithUpdate) MilestoneDoUseUser() string {
	return n.NewIdentityInfo.UserID
}

func (n *NodeDoUseIdentityWithUpdate) MilestoneDoUseIdentity() *identity.Info {
	return n.NewIdentityInfo
}

func (n *NodeDoUseIdentityWithUpdate) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Identities.Update(n.OldIdentityInfo, n.NewIdentityInfo)
		}),
	}, nil
}
