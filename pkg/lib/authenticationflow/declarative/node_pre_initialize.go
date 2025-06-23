package declarative

import (
	"context"

	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	blocking "github.com/authgear/authgear-server/pkg/api/event/blocking"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodePreInitialize{})
}

func NewNodePreInitialize(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.ReactToResult, error) {

	n := &NodePreInitialize{}

	authCtx, err := GetAuthenticationContext(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	payload := &blocking.AuthenticationPreInitializeBlockingEventPayload{
		AuthenticationContext: *authCtx,

		Constraints:               nil,
		BotProtectionRequirements: nil,
	}
	e, err := deps.Events.PrepareBlockingEventWithTx(ctx, payload)
	if err != nil {
		return nil, err
	}

	var delayedFunction authflow.DelayedOneTimeFunction = func(ctx context.Context, deps *authflow.Dependencies) error {
		err = deps.Events.DispatchEventWithoutTx(ctx, e)
		if err != nil {
			return err
		}
		n.IsPreInitializeInvoked = true
		n.Constraints = payload.Constraints
		n.BotProtectionRequirements = payload.BotProtectionRequirements
		return nil
	}

	return &authflow.NodeWithDelayedOneTimeFunction{
		Node:                   authflow.NewNodeSimple(n),
		DelayedOneTimeFunction: delayedFunction,
	}, nil
}

type NodePreInitialize struct {
	IsPreInitializeInvoked    bool                                `json:"is_pre_initialize_invoked"`
	Constraints               *eventapi.Constraints               `json:"constraints,omitempty"`
	BotProtectionRequirements *eventapi.BotProtectionRequirements `json:"bot_protection_requirements,omitempty"`
}

var _ authflow.NodeSimple = &NodePreInitialize{}
var _ authflow.InputReactor = &NodePreInitialize{}
var _ authflow.Milestone = &NodePreInitialize{}
var _ MilestoneConstraintsProvider = &NodePreInitialize{}
var _ MilestoneBotProjectionRequirementsProvider = &NodePreInitialize{}

func (*NodePreInitialize) Kind() string {
	return "NodePreInitialize"
}

func (n *NodePreInitialize) Milestone() {}
func (n *NodePreInitialize) MilestoneConstraintsProvider() *eventapi.Constraints {
	return n.Constraints
}
func (n *NodePreInitialize) MilestoneBotProjectionRequirementsProvider() *eventapi.BotProtectionRequirements {
	return n.BotProtectionRequirements
}

func (n *NodePreInitialize) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if n.IsPreInitializeInvoked {
		return nil, authflow.ErrEOF
	}
	return nil, authflow.ErrPauseAndRetryAccept
}

func (n *NodePreInitialize) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	return nil, authflow.ErrEOF
}
