package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIdentityPasskey{})
}

type NodeDoUseIdentityPasskey struct {
	AssertionResponse []byte              `json:"assertion_response,omitempty"`
	Identity          *identity.Info      `json:"identity,omitempty"`
	Authenticator     *authenticator.Info `json:"authenticator,omitempty"`
	RequireUpdate     bool                `json:"require_update,omitempty"`
}

func NewNodeDoUseIdentityPasskey(ctx context.Context, flows authflow.Flows, n *NodeDoUseIdentityPasskey) (*NodeDoUseIdentityPasskey, error) {
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

	return n, nil
}

var _ authflow.NodeSimple = &NodeDoUseIdentityPasskey{}
var _ authflow.EffectGetter = &NodeDoUseIdentityPasskey{}
var _ authflow.Milestone = &NodeDoUseIdentityPasskey{}
var _ MilestoneDoUseUser = &NodeDoUseIdentityPasskey{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentityPasskey{}
var _ MilestoneDidSelectAuthenticator = &NodeDoUseIdentityPasskey{}
var _ MilestoneDidAuthenticate = &NodeDoUseIdentityPasskey{}

func (*NodeDoUseIdentityPasskey) Kind() string {
	return "NodeDoUseIdentityPasskey"
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
