package declarative

import (
	"context"

	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	blocking "github.com/authgear/authgear-server/pkg/api/event/blocking"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

func init() {
	authflow.RegisterNode(&NodePreAuthenticate{})
}

type NodePreAuthenticate struct {
	IsPreAuthenticatedInvoked bool                  `json:"is_pre_authenticated_invoked"`
	Constraints               *eventapi.Constraints `json:"constraints,omitempty"`
	RateLimits                eventapi.RateLimits   `json:"rate_limits,omitempty"`
}

func newNodePreAuthenticate(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (*NodePreAuthenticate, authflow.DelayedOneTimeFunction, error) {
	authCtx, err := GetAuthenticationContext(ctx, deps, flows)
	if err != nil {
		return nil, nil, err
	}

	payload := &blocking.AuthenticationPreAuthenticatedBlockingEventPayload{
		Constraints:           nil,
		AuthenticationContext: *authCtx,
	}
	e, err := deps.Events.PrepareBlockingEventWithTx(ctx, payload)
	if err != nil {
		return nil, nil, err
	}

	n := &NodePreAuthenticate{
		IsPreAuthenticatedInvoked: false,
		Constraints:               nil,
	}

	var delayedFunction authflow.DelayedOneTimeFunction = func(ctx context.Context, deps *authflow.Dependencies) error {
		err = deps.Events.DispatchEventWithoutTx(ctx, e)
		if err != nil {
			return err
		}
		n.IsPreAuthenticatedInvoked = true
		n.Constraints = payload.Constraints
		n.RateLimits = payload.RateLimits
		return nil
	}

	return n, delayedFunction, nil

}

func NewNodePreAuthenticateNodeSimple(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.ReactToResult, error) {
	n, delayedFunction, err := newNodePreAuthenticate(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	return &authflow.NodeWithDelayedOneTimeFunction{
		Node:                   authflow.NewNodeSimple(n),
		DelayedOneTimeFunction: delayedFunction,
	}, nil
}

var _ authflow.NodeSimple = &NodePreAuthenticate{}
var _ authflow.Milestone = &NodePreAuthenticate{}
var _ authflow.InputReactor = &NodePreAuthenticate{}
var _ authflow.EffectGetter = &NodePreAuthenticate{}
var _ MilestoneConstraintsProvider = &NodePreAuthenticate{}
var _ MilestonePreAuthenticated = &NodePreAuthenticate{}

func (*NodePreAuthenticate) Kind() string {
	return "NodePreAuthenticate"
}

func (*NodePreAuthenticate) Milestone() {}
func (n *NodePreAuthenticate) MilestoneConstraintsProvider() *eventapi.Constraints {
	return n.Constraints
}
func (*NodePreAuthenticate) MilestonePreAuthenticated() {}

func (n *NodePreAuthenticate) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if n.IsPreAuthenticatedInvoked {
		return nil, authflow.ErrEOF
	}
	return nil, authflow.ErrPauseAndRetryAccept
}

func (n *NodePreAuthenticate) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	return nil, authflow.ErrEOF
}

func (n *NodePreAuthenticate) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			// If rate_limits is not returned in hook response, do not modify the weights
			if n.RateLimits == nil {
				return nil
			}
			ratelimit.SetRateLimitWeights(ctx, toRateLimitWeights(n.RateLimits))
			return nil
		}),
	}, nil
}
