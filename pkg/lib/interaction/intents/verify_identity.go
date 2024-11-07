package intents

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentVerifyIdentity{})
}

type IntentVerifyIdentity struct {
	UserID       string             `json:"user_id"`
	IdentityType model.IdentityType `json:"identity_type"`
	IdentityID   string             `json:"identity_id"`
}

func NewIntentVerifyIdentity(userID string, identityType model.IdentityType, identityID string) *IntentVerifyIdentity {
	return &IntentVerifyIdentity{
		UserID:       userID,
		IdentityType: identityType,
		IdentityID:   identityID,
	}
}

func (i *IntentVerifyIdentity) InstantiateRootNode(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	identityInfo, err := ctx.Identities.Get(goCtx, i.IdentityID)
	if err != nil {
		return nil, err
	}

	if identityInfo.UserID != i.UserID {
		return nil, api.NewInvariantViolated(
			"IdentityNotBelongToUser",
			"identity does not belong to the user",
			nil,
		)
	}

	return &nodes.NodeEnsureVerificationBegin{
		Identity:        identityInfo,
		RequestedByUser: true,
		PhoneOTPMode:    ctx.Config.Authenticator.OOB.SMS.PhoneOTPMode,
	}, nil
}

func (i *IntentVerifyIdentity) DeriveEdgesForNode(goCtx context.Context, graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeEnsureVerificationEnd:
		return []interaction.Edge{
			&nodes.EdgeDoVerifyIdentity{
				Identity:         node.Identity,
				NewVerifiedClaim: node.NewVerifiedClaim,
				RequestedByUser:  true,
			},
		}, nil

	case *nodes.NodeDoVerifyIdentity:
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
