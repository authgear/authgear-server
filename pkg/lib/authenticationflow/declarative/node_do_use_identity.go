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
	IsPostIdentifiedInvoked bool                  `json:"is_post_identified_invoked"`
	Constraints             *eventapi.Constraints `json:"constraints,omitempty"`
}

func NewNodeDoUseIdentity(ctx context.Context, flows authflow.Flows, deps *authflow.Dependencies, n *NodeDoUseIdentity) (authenticationflow.ReactToResult, error) {
	userID, err := getUserID(flows)
	if errors.Is(err, ErrNoUserID) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	if userID != "" && userID != n.Identity.UserID {
		return nil, ErrDifferentUserID
	}

	if userIDHint := authflow.GetUserIDHint(ctx); userIDHint != "" {
		if userIDHint != n.Identity.UserID {
			return nil, api.ErrMismatchedUser
		}
	}

	payload := &blocking.AuthenticationPostIdentifiedBlockingEventPayload{
		Identity:    n.Identity.ToModel(),
		Constraints: nil,
	}
	e, err := deps.Events.PrepareBlockingEventWithTx(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &authenticationflow.NodeWithDelayedOneTimeFunction{
		Node: authenticationflow.NewNodeSimple(n),
		DelayedOneTimeFunction: func(ctx context.Context, deps *authenticationflow.Dependencies) error {
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

var _ authflow.NodeSimple = &NodeDoUseIdentity{}
var _ authflow.Milestone = &NodeDoUseIdentity{}
var _ authflow.InputReactor = &NodeDoUseIdentity{}
var _ MilestoneDoUseUser = &NodeDoUseIdentity{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentity{}

func (*NodeDoUseIdentity) Kind() string {
	return "NodeDoUseIdentity"
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
