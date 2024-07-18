package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

func init() {
	interaction.RegisterNode(&NodeSelectIdentityEnd{})
}

type EdgeSelectIdentityEnd struct {
	IdentitySpec     *identity.Spec
	IsAuthentication bool
}

func (e *EdgeSelectIdentityEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	bypassRateLimit := false
	var bypassInput interface{ BypassInteractionIPRateLimit() bool }
	if interaction.Input(input, &bypassInput) {
		bypassRateLimit = bypassInput.BypassInteractionIPRateLimit()
	}

	var reservation *ratelimit.Reservation
	if !bypassRateLimit {
		spec := interaction.AccountEnumerationPerIPRateLimitBucketSpec(ctx.Config.Authentication, string(ctx.RemoteIP))
		reservation = ctx.RateLimiter.Reserve(spec)
		if err := reservation.Error(); err != nil {
			return nil, err
		}
	}
	defer ctx.RateLimiter.Cancel(reservation)

	var otherMatch *identity.Info
	exactMatch, otherMatches, err := ctx.Identities.SearchBySpec(e.IdentitySpec)
	if err != nil {
		return nil, err
	}

	if exactMatch == nil {
		// Exact match not found; consume account enumeration rate limit.
		if reservation != nil {
			reservation.Consume()
		}

		// Take the first one as other match.
		if len(otherMatches) > 0 {
			otherMatch = otherMatches[0]
		}

		if e.IsAuthentication {
			switch e.IdentitySpec.Type {
			case model.IdentityTypeOAuth:
				// This branch should be unreachable.
				break
			case model.IdentityTypeAnonymous, model.IdentityTypeBiometric:
				// Anonymous and biometric are handled in their own node.
				break
			case model.IdentityTypeLoginID:
				loginIDValue := e.IdentitySpec.LoginID.Value
				err = ctx.Events.DispatchEventOnCommit(&nonblocking.AuthenticationFailedLoginIDEventPayload{
					LoginID: loginIDValue,
				})
				if err != nil {
					return nil, err
				}
			case model.IdentityTypePasskey:
				break
			case model.IdentityTypeSIWE:
				break
			default:
				panic(fmt.Errorf("interaction: unknown identity type: %v", e.IdentitySpec.Type))
			}
		}
	}

	// Ensure info is up-to-date.
	var oldIdentityInfo *identity.Info
	if exactMatch != nil && exactMatch.Type == model.IdentityTypeOAuth {
		oldIdentityInfo = exactMatch
		exactMatch, err = ctx.Identities.UpdateWithSpec(exactMatch, e.IdentitySpec, identity.NewIdentityOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &NodeSelectIdentityEnd{
		IdentitySpec:    e.IdentitySpec,
		IdentityInfo:    exactMatch,
		OldIdentityInfo: oldIdentityInfo,
		OtherMatch:      otherMatch,
	}, nil
}

type NodeSelectIdentityEnd struct {
	IdentitySpec    *identity.Spec `json:"identity_spec"`
	IdentityInfo    *identity.Info `json:"identity_info"`
	OldIdentityInfo *identity.Info `json:"old_identity_info"`
	OtherMatch      *identity.Info `json:"other_match"`
}

func (n *NodeSelectIdentityEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityEnd) GetEffects() ([]interaction.Effect, error) {
	// Update OAuth identity
	eff := func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
		if n.OldIdentityInfo != nil && n.IdentityInfo != nil && n.IdentityInfo.Type == model.IdentityTypeOAuth {
			_, err := ctx.Identities.CheckDuplicated(n.IdentityInfo)
			if err != nil {
				if identity.IsErrDuplicatedIdentity(err) {
					return n.FillDetails(identity.Deprecated_ErrDuplicatedIdentity)
				}
				return err
			}

			err = ctx.Identities.Update(n.OldIdentityInfo, n.IdentityInfo)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// We declare two effects here so that
	// 1. When the interaction is still ongoing, we will see the updated identity.
	// 2. When the interaction finishes, the identity will be updated.
	return []interaction.Effect{
		interaction.EffectRun(eff),
		interaction.EffectOnCommit(eff),
	}, nil
}

func (n *NodeSelectIdentityEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeSelectIdentityEnd) FillDetails(err error) error {
	spec := n.IdentitySpec
	var otherSpec *identity.Spec

	// The spec fetches an exact match and other match.
	// Normally it is the sign-in cases.
	if n.IdentityInfo != nil && n.OtherMatch != nil {
		s := n.OtherMatch.ToSpec()
		otherSpec = &s
	}

	// The spec fetches an exact match.
	// Normally it is the sign-up cases.
	if n.IdentityInfo != nil && n.OtherMatch == nil {
		s := n.IdentityInfo.ToSpec()
		otherSpec = &s
	}

	return identityFillDetails(err, spec, otherSpec)
}
