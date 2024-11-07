package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoVerifyIdentity{})
}

type EdgeDoVerifyIdentity struct {
	Identity         *identity.Info
	NewVerifiedClaim *verification.Claim
	RequestedByUser  bool
}

func (e *EdgeDoVerifyIdentity) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	isAdminAPI := interaction.IsAdminAPI(rawInput)
	skipVerificationEvent := !e.RequestedByUser

	return &NodeDoVerifyIdentity{
		Identity:              e.Identity,
		NewVerifiedClaim:      e.NewVerifiedClaim,
		IsAdminAPI:            isAdminAPI,
		SkipVerificationEvent: skipVerificationEvent,
	}, nil
}

type NodeDoVerifyIdentity struct {
	Identity              *identity.Info      `json:"identity"`
	NewVerifiedClaim      *verification.Claim `json:"new_verified_claim"`
	IsAdminAPI            bool                `json:"is_admin_api"`
	SkipVerificationEvent bool                `json:"skip_verification_event"`
}

func (n *NodeDoVerifyIdentity) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoVerifyIdentity) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.NewVerifiedClaim != nil {
				if err := ctx.Verification.MarkClaimVerified(goCtx, n.NewVerifiedClaim); err != nil {
					return err
				}
			}

			return nil
		}),
		interaction.EffectOnCommit(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.SkipVerificationEvent {
				return nil
			}

			var e event.Payload
			if n.NewVerifiedClaim != nil {
				if payload, ok := nonblocking.NewIdentityVerifiedEventPayload(
					model.UserRef{
						Meta: model.Meta{
							ID: n.Identity.UserID,
						},
					},
					n.Identity.ToModel(),
					string(n.NewVerifiedClaim.Name),
					n.IsAdminAPI,
				); ok {
					e = payload
				}
			}

			if e != nil {
				err := ctx.Events.DispatchEventOnCommit(goCtx, e)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoVerifyIdentity) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
