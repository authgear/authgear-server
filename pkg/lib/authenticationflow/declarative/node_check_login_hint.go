package declarative

import (
	"context"
	"fmt"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

func init() {
	authflow.RegisterNode(&NodeCheckLoginHint{})
}

type NodeCheckLoginHint struct {
	UserID    string `json:"user_id,omitempty"`
	LoginHint string `json:"login_hint,omitempty"`
}

func NewNodeCheckLoginHint(ctx context.Context, deps *authflow.Dependencies, userID string) (*NodeCheckLoginHint, error) {
	loginHintStr := authflow.GetLoginHint(ctx)
	node := &NodeCheckLoginHint{
		UserID:    userID,
		LoginHint: loginHintStr,
	}
	err := node.check(deps)
	if err != nil {
		return nil, err
	}

	return node, nil
}

var _ authflow.NodeSimple = &NodeCheckLoginHint{}
var _ MilestoneCheckLoginHint = &NodeCheckLoginHint{}

func (n *NodeCheckLoginHint) Kind() string {
	return "NodeCheckLoginHint"
}

func (n *NodeCheckLoginHint) Milestone()               {}
func (n *NodeCheckLoginHint) MilestoneCheckLoginHint() {}

func (n *NodeCheckLoginHint) check(deps *authflow.Dependencies) error {
	loginHint, err := oauth.ParseLoginHint(n.LoginHint)
	if err != nil {
		// Not a valid login_hint, skip the check
		return nil
	}
	if !loginHint.Enforce {
		// Not enforced, skip the check
		return nil
	}
	switch loginHint.Type {
	case oauth.LoginHintTypeLoginID:
		return n.checkEnforcedLoginID(deps, loginHint)
	default:
		panic(fmt.Errorf("enforcing login_hint of type %s unsupported", loginHint.Type))
	}
}

func (n *NodeCheckLoginHint) checkEnforcedLoginID(deps *authflow.Dependencies, hint *oauth.LoginHint) error {
	userIDs, err := deps.UserFacade.GetUserIDsByLoginHint(hint)
	if err != nil {
		return err
	}
	for _, userID := range userIDs {
		if userID == n.UserID {
			return nil
		}
	}
	return ErrDifferentUserID
}
