package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
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

func NewNodeDoUseIDToken(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDoUseIDToken) (*NodeDoUseIDToken, error) {
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

	return n, nil
}

func (*NodeDoUseIDToken) Kind() string {
	return "NodeDoUseIDToken"
}

func (*NodeDoUseIDToken) Milestone() {}

func (n *NodeDoUseIDToken) MilestoneDoUseUser() string {
	return n.UserID
}
