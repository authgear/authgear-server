package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	blocking "github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIdentity{})
}

type NodeDoUseIdentity struct {
	Identity                *identity.Info        `json:"identity,omitempty"`
	IdentitySpec            *identity.Spec        `json:"identity_spec,omitempty"`
	IsPostIdentifiedInvoked bool                  `json:"is_post_identified_invoked"`
	Constraints             *eventapi.Constraints `json:"constraints,omitempty"`
}

func NewNodeDoUseIdentity(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDoUseIdentity) (*NodeDoUseIdentity, authflow.DelayedOneTimeFunction, error) {
	userID, err := getUserID(flows)
	if errors.Is(err, ErrNoUserID) {
		err = nil
	}
	if err != nil {
		return nil, nil, err
	}

	if userID != "" && userID != n.Identity.UserID {
		return nil, nil, ErrDifferentUserID
	}

	if userIDHint := authflow.GetUserIDHint(ctx); userIDHint != "" {
		if userIDHint != n.Identity.UserID {
			return nil, nil, api.ErrMismatchedUser
		}
	}

	authCtx, err := GetAuthenticationContext(ctx, deps, flows)
	if err != nil {
		return nil, nil, err
	}

	idenModel := n.Identity.ToModel()
	// Include the identity of this node
	authCtx.AddAssertedIdentity(idenModel)

	payload := &blocking.AuthenticationPostIdentifiedBlockingEventPayload{
		Identity:       &idenModel,
		Constraints:    nil,
		Identification: n.Identity.ToIdentification(),
		Authentication: *authCtx,
	}
	e, err := deps.Events.PrepareBlockingEventWithTx(ctx, payload)
	if err != nil {
		return nil, nil, err
	}

	var delayedFunction authflow.DelayedOneTimeFunction = func(ctx context.Context, deps *authenticationflow.Dependencies) error {
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

func NewNodeDoUseIdentityReactToResult(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDoUseIdentity) (authenticationflow.ReactToResult, error) {
	_, delayedFunction, err := NewNodeDoUseIdentity(ctx, deps, flows, n)
	if err != nil {
		return nil, err
	}

	return &authenticationflow.NodeWithDelayedOneTimeFunction{
		Node:                   authenticationflow.NewNodeSimple(n),
		DelayedOneTimeFunction: delayedFunction,
	}, nil
}

var _ authflow.NodeSimple = &NodeDoUseIdentity{}
var _ authflow.Milestone = &NodeDoUseIdentity{}
var _ authflow.InputReactor = &NodeDoUseIdentity{}
var _ MilestoneDoUseUser = &NodeDoUseIdentity{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentity{}
var _ MilestoneGetIdentitySpecs = &NodeDoUseIdentity{}
var _ MilestoneConstraintsProvider = &NodeDoUseIdentity{}

func (*NodeDoUseIdentity) Kind() string {
	return "NodeDoUseIdentity"
}

func (n *NodeDoUseIdentity) MilestoneConstraintsProvider() *eventapi.Constraints {
	return n.Constraints
}

func (n *NodeDoUseIdentity) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	if n.IsPostIdentifiedInvoked {
		return nil, authflow.ErrEOF
	}
	return nil, authflow.ErrPauseAndRetryAccept
}

func (n *NodeDoUseIdentity) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return nil, authflow.ErrEOF
}

func (*NodeDoUseIdentity) Milestone() {}
func (n *NodeDoUseIdentity) MilestoneDoUseUser() string {
	return n.Identity.UserID
}
func (n *NodeDoUseIdentity) MilestoneDoUseIdentity() *identity.Info { return n.Identity }

func (n *NodeDoUseIdentity) MilestoneGetIdentitySpecs() []*identity.Spec {
	return []*identity.Spec{n.IdentitySpec}
}
