package intents

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
)

func init() {
	newinteraction.RegisterIntent(&IntentVerifyIdentityResume{})
}

type IntentVerifyIdentityResume struct {
	VerificationCodeID string `json:"verification_code_id"`
}

func NewIntentVerifyIdentityResume(codeID string) *IntentVerifyIdentityResume {
	return &IntentVerifyIdentityResume{
		VerificationCodeID: codeID,
	}
}

func (i *IntentVerifyIdentityResume) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	code, err := ctx.Verification.GetCode(i.VerificationCodeID)
	if err != nil {
		return nil, err
	}

	identityInfo, err := ctx.Identities.Get(
		code.UserID,
		authn.IdentityType(code.IdentityType),
		code.IdentityID)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		// Identity not found -> treat as if code is invalid
		return nil, verification.ErrCodeNotFound
	} else if err != nil {
		return nil, err
	}

	edge := &nodes.EdgeVerifyIdentityResume{
		Code:     code,
		Identity: identityInfo,
	}
	return edge.Instantiate(ctx, graph, nil)
}

func (i *IntentVerifyIdentityResume) DeriveEdgesForNode(graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeEnsureVerificationEnd:
		return []newinteraction.Edge{
			&nodes.EdgeDoVerifyIdentity{
				Identity:         node.Identity,
				NewAuthenticator: node.NewAuthenticator,
			},
		}, nil

	case *nodes.NodeDoVerifyIdentity:
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
