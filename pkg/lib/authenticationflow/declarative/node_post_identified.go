package declarative

import (
	"context"

	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	blocking "github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodePostIdentified{})
}

type NodePostIdentifiedOptions struct {
	Identification model.Identification
}

func NewNodePostIdentified(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, opts *NodePostIdentifiedOptions) (authflow.ReactToResult, error) {

	n := &NodePostIdentified{
		Identification: &opts.Identification,
	}

	authCtx, err := GetAuthenticationContext(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	payload := &blocking.AuthenticationPostIdentifiedBlockingEventPayload{
		Identification: *n.Identification,
		Authentication: *authCtx,

		Constraints:               nil,
		BotProtectionRequirements: nil,
	}
	e, err := deps.Events.PrepareBlockingEventWithTx(ctx, payload)
	if err != nil {
		return nil, err
	}

	var delayedFunction authflow.DelayedOneTimeFunction = func(ctx context.Context, deps *authenticationflow.Dependencies) error {
		err = deps.Events.DispatchEventWithoutTx(ctx, e)
		if err != nil {
			return err
		}
		n.IsPostIdentifiedInvoked = true
		n.Constraints = payload.Constraints
		n.BotProtectionRequirements = payload.BotProtectionRequirements
		return nil
	}

	return &authflow.NodeWithDelayedOneTimeFunction{
		Node:                   authflow.NewNodeSimple(n),
		DelayedOneTimeFunction: delayedFunction,
	}, nil
}

type NodePostIdentified struct {
	Identification *model.Identification `json:"identification"`

	IsPostIdentifiedInvoked   bool                                `json:"is_post_identified_invoked"`
	Constraints               *eventapi.Constraints               `json:"constraints,omitempty"`
	BotProtectionRequirements *eventapi.BotProtectionRequirements `json:"bot_protection_requirements,omitempty"`
}

var _ authflow.NodeSimple = &NodePostIdentified{}
var _ authflow.InputReactor = &NodePostIdentified{}
var _ authflow.Milestone = &NodePostIdentified{}
var _ MilestoneConstraintsProvider = &NodePostIdentified{}
var _ MilestoneBotProjectionRequirementsProvider = &NodePostIdentified{}

func (*NodePostIdentified) Kind() string {
	return "NodePostIdentified"
}

func (n *NodePostIdentified) Milestone() {}
func (n *NodePostIdentified) MilestoneConstraintsProvider() *eventapi.Constraints {
	return n.Constraints
}
func (n *NodePostIdentified) MilestoneBotProjectionRequirementsProvider() *eventapi.BotProtectionRequirements {
	return n.BotProtectionRequirements
}

func (n *NodePostIdentified) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	if n.IsPostIdentifiedInvoked {
		return nil, authflow.ErrEOF
	}
	return nil, authflow.ErrPauseAndRetryAccept
}

func (n *NodePostIdentified) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return nil, authflow.ErrEOF
}
