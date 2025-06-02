package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	blocking "github.com/authgear/authgear-server/pkg/api/event/blocking"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIdentityWithUpdate{})
}

type NodeDoUseIdentityWithUpdate struct {
	OldIdentityInfo         *identity.Info        `json:"old_identity_info,omitempty"`
	NewIdentityInfo         *identity.Info        `json:"new_identity_info,omitempty"`
	NewIdentitySpec         *identity.Spec        `json:"new_identity_spec,omitempty"`
	IsPostIdentifiedInvoked bool                  `json:"is_post_identified_invoked"`
	Constraints             *eventapi.Constraints `json:"constraints,omitempty"`
}

func NewNodeDoUseIdentityWithUpdate(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, oldIdentityInfo *identity.Info, spec *identity.Spec) (authflow.ReactToResult, error) {
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

	newIdentityInfo, err := deps.Identities.UpdateWithSpec(ctx, oldIdentityInfo, spec, identity.NewIdentityOptions{})
	if err != nil {
		return nil, err
	}

	n := &NodeDoUseIdentityWithUpdate{
		OldIdentityInfo: oldIdentityInfo,
		NewIdentityInfo: newIdentityInfo,
		NewIdentitySpec: spec,
	}

	payload := &blocking.AuthenticationPostIdentifiedBlockingEventPayload{
		Identity:    n.NewIdentityInfo.ToModel(),
		Constraints: nil,
	}
	e, err := deps.Events.PrepareBlockingEventWithTx(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &authflow.NodeWithDelayedOneTimeFunction{
		Node: authflow.NewNodeSimple(n),
		DelayedOneTimeFunction: func(ctx context.Context, deps *authflow.Dependencies) error {
			err = deps.Events.DispatchEventWithoutTx(ctx, e)
			if err != nil {
				return err
			}
			n.IsPostIdentifiedInvoked = true
			n.Constraints = payload.Constraints
			return nil
		},
	}, nil
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
	if n.IsPostIdentifiedInvoked {
		return nil, authflow.ErrEOF
	}
	return nil, authflow.ErrPauseAndRetryAccept
}

func (n *NodeDoUseIdentityWithUpdate) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	return nil, authflow.ErrEOF
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
			return deps.Identities.Update(ctx, n.OldIdentityInfo, n.NewIdentityInfo)
		}),
	}, nil
}

func (n *NodeDoUseIdentityWithUpdate) MilestoneGetIdentitySpecs() []*identity.Spec {
	return []*identity.Spec{n.NewIdentitySpec}
}
