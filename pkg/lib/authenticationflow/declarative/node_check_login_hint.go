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
	if hint.Type != oauth.LoginHintTypeLoginID {
		panic("This method should only be called if login_hint have type login_id")
	}

	checkForAtLeastOneIdentityWithClaim := func(claimName string, value string) error {
		identities, err := deps.Identities.ListByClaim(claimName, value)
		if err != nil {
			return err
		}
		for _, iden := range identities {
			if iden.UserID == n.UserID {
				return nil
			}
		}
		return ErrDifferentUserID
	}

	switch {
	case hint.LoginIDEmail != "":
		return checkForAtLeastOneIdentityWithClaim("email", hint.LoginIDEmail)
	case hint.LoginIDPhone != "":
		return checkForAtLeastOneIdentityWithClaim("phone_number", hint.LoginIDPhone)
	case hint.LoginIDUsername != "":
		return checkForAtLeastOneIdentityWithClaim("preferred_username", hint.LoginIDUsername)
	default:
		return fmt.Errorf("unable to enforce login_hint as no login_id provided")
	}
}
