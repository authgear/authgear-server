package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	blocking "github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIdentityPasskey{})
}

type NodeDoUseIdentityPasskey struct {
	AssertionResponse       []byte                `json:"assertion_response,omitempty"`
	Identity                *identity.Info        `json:"identity,omitempty"`
	IdentitySpec            *identity.Spec        `json:"identity_spec,omitempty"`
	Authenticator           *authenticator.Info   `json:"authenticator,omitempty"`
	RequireUpdate           bool                  `json:"require_update,omitempty"`
	IsPostIdentifiedInvoked bool                  `json:"is_post_identified_invoked"`
	Constraints             *eventapi.Constraints `json:"constraints,omitempty"`
}

func NewNodeDoUseIdentityPasskey(ctx context.Context, flows authflow.Flows, deps *authflow.Dependencies, n *NodeDoUseIdentityPasskey) (authenticationflow.ReactToResult, error) {
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

var _ authflow.NodeSimple = &NodeDoUseIdentityPasskey{}
var _ authflow.EffectGetter = &NodeDoUseIdentityPasskey{}
var _ authflow.Milestone = &NodeDoUseIdentityPasskey{}
var _ authflow.InputReactor = &NodeDoUseIdentityPasskey{}
var _ MilestoneDoUseUser = &NodeDoUseIdentityPasskey{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentityPasskey{}
var _ MilestoneDidSelectAuthenticator = &NodeDoUseIdentityPasskey{}
var _ MilestoneDidAuthenticate = &NodeDoUseIdentityPasskey{}
var _ MilestoneGetIdentitySpecs = &NodeDoUseIdentityPasskey{}

func (*NodeDoUseIdentityPasskey) Kind() string {
	return "NodeDoUseIdentityPasskey"
}

func (n *NodeDoUseIdentityPasskey) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	if n.IsPostIdentifiedInvoked {
		return nil, authflow.ErrEOF
	}
	return nil, authflow.ErrPauseAndRetryAccept
}

func (n *NodeDoUseIdentityPasskey) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return nil, authflow.ErrEOF
}

func (n *NodeDoUseIdentityPasskey) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.RequireUpdate {
				return deps.Authenticators.Update(ctx, n.Authenticator)
			}
			return nil
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.PasskeyService.ConsumeAssertionResponse(ctx, n.AssertionResponse)
		}),
	}, nil
}

func (*NodeDoUseIdentityPasskey) Milestone() {}
func (n *NodeDoUseIdentityPasskey) MilestoneDoUseUser() string {
	return n.Identity.UserID
}
func (n *NodeDoUseIdentityPasskey) MilestoneDoUseIdentity() *identity.Info { return n.Identity }
func (n *NodeDoUseIdentityPasskey) MilestoneDidSelectAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDoUseIdentityPasskey) MilestoneDidAuthenticate() (amr []string) {
	return n.Authenticator.AMR()
}

func (n *NodeDoUseIdentityPasskey) MilestoneGetIdentitySpecs() []*identity.Spec {
	return []*identity.Spec{n.IdentitySpec}
}
