package intents

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentVerifyIdentityResume{})
}

type IntentVerifyIdentityResume struct {
	VerificationCodeID string `json:"verification_code_id"`
}

func NewIntentVerifyIdentityResume(codeID string) *IntentVerifyIdentityResume {
	return &IntentVerifyIdentityResume{
		VerificationCodeID: codeID,
	}
}

func (i *IntentVerifyIdentityResume) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
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

func (i *IntentVerifyIdentityResume) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeEnsureVerificationEnd:
		return []interaction.Edge{
			&nodes.EdgeDoVerifyIdentity{
				Identity:         node.Identity,
				NewVerifiedClaim: node.NewVerifiedClaim,
			},
		}, nil

	case *nodes.NodeDoVerifyIdentity:
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
