package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIDToken{})
}

type NodeDoUseIDToken struct {
	IDToken string `json:"id_token,omitempty"`

	UserID string `json:"user_id,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUseIDToken{}
var _ authflow.Milestone = &NodeDoUseIDToken{}
var _ MilestoneDoUseUser = &NodeDoUseIDToken{}
var _ authflow.InputReactor = &NodeDoUseIDToken{}

func (n *NodeDoUseIDToken) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	return nil, nil
}

func (n *NodeDoUseIDToken) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	return NewNodePostIdentified(ctx, deps, flows, &NodePostIdentifiedOptions{
		// Identify with id_token does not trigger rate limit
		RateLimitReservation: nil,
		Identification: model.Identification{
			Identification: model.AuthenticationFlowIdentificationIDToken,
			Identity:       nil,
			IDToken:        &n.IDToken,
		},
	})
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

	return authflow.NewNodeSimple(n), nil
}

func (*NodeDoUseIDToken) Kind() string {
	return "NodeDoUseIDToken"
}

func (*NodeDoUseIDToken) Milestone() {}

func (n *NodeDoUseIDToken) MilestoneDoUseUser() string {
	return n.UserID
}
