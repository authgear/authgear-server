package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIdentityOAuth{})
}

type NodeDoUseIdentityOAuth struct {
	OldIdentityInfo *identity.Info `json:"old_identity_info,omitempty"`
	NewIdentityInfo *identity.Info `json:"new_identity_info,omitempty"`
}

func NewNodeDoUseIdentityOAuth(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, oldIdentityInfo *identity.Info, spec *identity.Spec) (*NodeDoUseIdentityOAuth, error) {
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

	return &NodeDoUseIdentityOAuth{
		OldIdentityInfo: oldIdentityInfo,
		NewIdentityInfo: newIdentityInfo,
	}, nil
}

var _ authflow.NodeSimple = &NodeDoUseIdentityOAuth{}
var _ authflow.EffectGetter = &NodeDoUseIdentityOAuth{}
var _ authflow.Milestone = &NodeDoUseIdentityOAuth{}
var _ MilestoneDoUseUser = &NodeDoUseIdentityOAuth{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentityOAuth{}

func (*NodeDoUseIdentityOAuth) Kind() string {
	return "NodeDoUseIdentityOAuth"
}

func (*NodeDoUseIdentityOAuth) Milestone() {}
func (n *NodeDoUseIdentityOAuth) MilestoneDoUseUser() string {
	return n.NewIdentityInfo.UserID
}

func (n *NodeDoUseIdentityOAuth) MilestoneDoUseIdentity() *identity.Info { return n.NewIdentityInfo }

func (n *NodeDoUseIdentityOAuth) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Identities.Update(n.OldIdentityInfo, n.NewIdentityInfo)
		}),
	}, nil
}
