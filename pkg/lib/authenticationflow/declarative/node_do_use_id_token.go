package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	blocking "github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIDToken{})
}

type NodeDoUseIDToken struct {
	IDToken                 string                `json:"id_token,omitempty"`
	IsPostIdentifiedInvoked bool                  `json:"is_post_identified_invoked"`
	Constraints             *eventapi.Constraints `json:"constraints,omitempty"`

	UserID string `json:"user_id,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUseIDToken{}
var _ authflow.Milestone = &NodeDoUseIDToken{}
var _ MilestoneDoUseUser = &NodeDoUseIDToken{}
var _ authflow.InputReactor = &NodeDoUseIDToken{}
var _ MilestoneConstraintsProvider = &NodeDoUseIDToken{}

func (n *NodeDoUseIDToken) MilestoneConstraintsProvider() *eventapi.Constraints {
	return n.Constraints
}

func (n *NodeDoUseIDToken) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	if n.IsPostIdentifiedInvoked {
		return nil, authflow.ErrEOF
	}
	return nil, authflow.ErrPauseAndRetryAccept
}

func (n *NodeDoUseIDToken) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return nil, authflow.ErrEOF
}

func NewNodeDoUseIDToken(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDoUseIDToken) (authflow.ReactToResult, error) {
	token, err := deps.IDTokens.VerifyIDToken(n.IDToken)
	if err != nil {
		return nil, apierrors.NewInvalid("invalid ID token")
	}

	userID := token.Subject()
	_, err = deps.Users.GetRaw(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, api.ErrUserNotFound
		}

		return nil, err
	}

	n.UserID = userID

	authCtx, err := GetAuthenticationContext(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	payload := &blocking.AuthenticationPostIdentifiedBlockingEventPayload{
		Identity:       nil,
		IDToken:        &n.IDToken,
		Constraints:    nil,
		Identification: config.AuthenticationFlowIdentificationIDToken,
		Authentication: *authCtx,
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
		return nil
	}

	return &authflow.NodeWithDelayedOneTimeFunction{
		Node:                   authflow.NewNodeSimple(n),
		DelayedOneTimeFunction: delayedFunction,
	}, nil
}

func (*NodeDoUseIDToken) Kind() string {
	return "NodeDoUseIDToken"
}

func (*NodeDoUseIDToken) Milestone() {}

func (n *NodeDoUseIDToken) MilestoneDoUseUser() string {
	return n.UserID
}
