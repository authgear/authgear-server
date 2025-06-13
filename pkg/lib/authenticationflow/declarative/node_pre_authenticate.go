package declarative

import (
	"context"

	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	blocking "github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authenticationflow.RegisterNode(&NodePreAuthenticate{})
}

type NodePreAuthenticate struct {
	IsPreAuthenticatedInvoked bool                  `json:"is_pre_authenticated_invoked"`
	Constraints               *eventapi.Constraints `json:"constraints,omitempty"`
}

func newNodePreAuthenticate(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (*NodePreAuthenticate, authenticationflow.DelayedOneTimeFunction, error) {
	authCtx, err := GetAuthenticationContext(ctx, deps, flows)
	if err != nil {
		return nil, nil, err
	}

	payload := &blocking.AuthenticationPreAuthenticatedBlockingEventPayload{
		Constraints:    nil,
		Authentication: *authCtx,
	}
	e, err := deps.Events.PrepareBlockingEventWithTx(ctx, payload)
	if err != nil {
		return nil, nil, err
	}

	n := &NodePreAuthenticate{
		IsPreAuthenticatedInvoked: false,
		Constraints:               nil,
	}

	var delayedFunction authenticationflow.DelayedOneTimeFunction = func(ctx context.Context, deps *authenticationflow.Dependencies) error {
		err = deps.Events.DispatchEventWithoutTx(ctx, e)
		if err != nil {
			return err
		}
		n.IsPreAuthenticatedInvoked = true
		n.Constraints = payload.Constraints
		return nil
	}

	return n, delayedFunction, nil

}

func NewNodePreAuthenticateNodeSimple(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.ReactToResult, error) {
	n, delayedFunction, err := newNodePreAuthenticate(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	return &authenticationflow.NodeWithDelayedOneTimeFunction{
		Node:                   authenticationflow.NewNodeSimple(n),
		DelayedOneTimeFunction: delayedFunction,
	}, nil
}

var _ authenticationflow.NodeSimple = &NodePreAuthenticate{}
var _ authenticationflow.Milestone = &NodePreAuthenticate{}
var _ authenticationflow.InputReactor = &NodePreAuthenticate{}
var _ MilestoneConstraintsProvider = &NodePreAuthenticate{}

func (*NodePreAuthenticate) Kind() string {
	return "NodePreAuthenticate"
}

func (*NodePreAuthenticate) Milestone() {}
func (n *NodePreAuthenticate) MilestoneConstraintsProvider() *eventapi.Constraints {
	return n.Constraints
}

func (n *NodePreAuthenticate) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	if n.IsPreAuthenticatedInvoked {
		return nil, authenticationflow.ErrEOF
	}
	return nil, authenticationflow.ErrPauseAndRetryAccept
}

func (n *NodePreAuthenticate) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return nil, authenticationflow.ErrEOF
}
