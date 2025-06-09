package declarative

import (
	"context"

	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	blocking "github.com/authgear/authgear-server/pkg/api/event/blocking"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoCreateIdentity{})
}

type NodeDoCreateIdentityOptions struct {
	SkipCreate   bool
	Identity     *identity.Info
	IdentitySpec *identity.Spec
}

func NewNodeDoCreateIdentity(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, opts NodeDoCreateIdentityOptions) (*NodeDoCreateIdentity, authflow.DelayedOneTimeFunction, error) {
	n := &NodeDoCreateIdentity{
		SkipCreate:   opts.SkipCreate,
		Identity:     opts.Identity,
		IdentitySpec: opts.IdentitySpec,
	}

	authCtx, err := GetAuthenticationContext(ctx, deps, flows)
	if err != nil {
		return nil, nil, err
	}

	// Include the identity of this node
	authCtx.AddAssertedIdentity(n.Identity.ToModel())

	payload := &blocking.AuthenticationPostIdentifiedBlockingEventPayload{
		Identity:       n.Identity.ToModel(),
		Constraints:    nil,
		Authentication: *authCtx,
	}
	e, err := deps.Events.PrepareBlockingEventWithTx(ctx, payload)
	if err != nil {
		return nil, nil, err
	}

	var delayedFunction authflow.DelayedOneTimeFunction = func(ctx context.Context, deps *authflow.Dependencies) error {
		err = deps.Events.DispatchEventWithoutTx(ctx, e)
		if err != nil {
			return err
		}
		n.IsPostIdentifiedInvoked = true
		n.Constraints = payload.Constraints
		return nil
	}

	return n, delayedFunction, nil
}

func NewNodeDoCreateIdentityReactToResult(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, opts NodeDoCreateIdentityOptions) (authflow.ReactToResult, error) {
	node, delayedFunction, err := NewNodeDoCreateIdentity(ctx, deps, flows, opts)
	if err != nil {
		return nil, err
	}

	return &authflow.NodeWithDelayedOneTimeFunction{
		Node:                   authflow.NewNodeSimple(node),
		DelayedOneTimeFunction: delayedFunction,
	}, nil
}

type NodeDoCreateIdentity struct {
	SkipCreate              bool                  `json:"skip_create,omitempty"`
	Identity                *identity.Info        `json:"identity,omitempty"`
	IdentitySpec            *identity.Spec        `json:"identity_spec,omitempty"`
	IsPostIdentifiedInvoked bool                  `json:"is_post_identified_invoked"`
	Constraints             *eventapi.Constraints `json:"constraints,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoCreateIdentity{}
var _ authflow.Milestone = &NodeDoCreateIdentity{}
var _ MilestoneDoCreateIdentity = &NodeDoCreateIdentity{}
var _ MilestoneGetIdentitySpecs = &NodeDoCreateIdentity{}
var _ authflow.EffectGetter = &NodeDoCreateIdentity{}
var _ authflow.InputReactor = &NodeDoCreateIdentity{}

func (n *NodeDoCreateIdentity) Kind() string {
	return "NodeDoCreateIdentity"
}

func (*NodeDoCreateIdentity) Milestone() {}
func (n *NodeDoCreateIdentity) MilestoneDoCreateIdentity() *identity.Info {
	return n.Identity
}
func (n *NodeDoCreateIdentity) MilestoneGetIdentitySpecs() []*identity.Spec {
	return []*identity.Spec{n.IdentitySpec}
}
func (n *NodeDoCreateIdentity) MilestoneDoCreateIdentitySkipCreate() {
	n.SkipCreate = true
}
func (n *NodeDoCreateIdentity) MilestoneDoCreateIdentityUpdate(newInfo *identity.Info) {
	n.Identity = newInfo
}

func (n *NodeDoCreateIdentity) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if n.IsPostIdentifiedInvoked {
		return nil, authflow.ErrEOF
	}
	return nil, authflow.ErrPauseAndRetryAccept
}

func (n *NodeDoCreateIdentity) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	return nil, authflow.ErrEOF
}

func (n *NodeDoCreateIdentity) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	if n.SkipCreate {
		return nil, nil
	}
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			err := deps.Identities.Create(ctx, n.Identity)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}
