package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIdentityWithUpdate{})
}

type NodeDoUseIdentityWithUpdate struct {
	*NodeDoUseIdentity
	OldIdentityInfo *identity.Info `json:"old_identity_info,omitempty"`
}

func NewNodeDoUseIdentityWithUpdate(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, oldIdentityInfo *identity.Info, spec *identity.Spec) (authflow.ReactToResult, error) {
	newIdentityInfo, err := deps.Identities.UpdateWithSpec(ctx, oldIdentityInfo, spec, identity.NewIdentityOptions{})
	if err != nil {
		return nil, err
	}

	nodeDoUseIden, err := NewNodeDoUseIdentity(ctx, deps, flows, &NodeDoUseIdentity{
		Identity:     newIdentityInfo,
		IdentitySpec: spec,
	})
	if err != nil {
		return nil, err
	}

	n := &NodeDoUseIdentityWithUpdate{
		NodeDoUseIdentity: nodeDoUseIden,
		OldIdentityInfo:   oldIdentityInfo,
	}

	return authflow.NewNodeSimple(n), nil
}

var _ authflow.NodeSimple = &NodeDoUseIdentityWithUpdate{}
var _ authflow.EffectGetter = &NodeDoUseIdentityWithUpdate{}
var _ authflow.Milestone = &NodeDoUseIdentityWithUpdate{}
var _ authflow.InputReactor = &NodeDoUseIdentityWithUpdate{}
var _ MilestoneDoUseUser = &NodeDoUseIdentityWithUpdate{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentityWithUpdate{}
var _ MilestoneGetIdentitySpecs = &NodeDoUseIdentityWithUpdate{}

func (*NodeDoUseIdentityWithUpdate) Kind() string {
	return "NodeDoUseIdentityWithUpdate"
}

func (n *NodeDoUseIdentityWithUpdate) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return n.NodeDoUseIdentity.CanReactTo(ctx, deps, flows)
}

func (n *NodeDoUseIdentityWithUpdate) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	return n.NodeDoUseIdentity.ReactTo(ctx, deps, flows, input)
}

func (*NodeDoUseIdentityWithUpdate) Milestone() {}
func (n *NodeDoUseIdentityWithUpdate) MilestoneDoUseUser() string {
	return n.NodeDoUseIdentity.Identity.UserID
}

func (n *NodeDoUseIdentityWithUpdate) MilestoneDoUseIdentity() *identity.Info {
	return n.NodeDoUseIdentity.MilestoneDoUseIdentity()
}
func (n *NodeDoUseIdentityWithUpdate) MilestoneDoUseIdentityIdentification() model.Identification {
	return n.NodeDoUseIdentity.MilestoneDoUseIdentityIdentification()
}

func (n *NodeDoUseIdentityWithUpdate) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.Identities.Update(ctx, n.OldIdentityInfo, n.NodeDoUseIdentity.Identity)
		}),
	}, nil
}

func (n *NodeDoUseIdentityWithUpdate) MilestoneGetIdentitySpecs() []*identity.Spec {
	return n.NodeDoUseIdentity.MilestoneGetIdentitySpecs()
}
