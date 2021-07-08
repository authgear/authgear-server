package nodes

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httputil"
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
	if interaction.AsInput(input, &bypassInput) {
		bypassRateLimit = bypassInput.BypassInteractionIPRateLimit()
	}

	if !bypassRateLimit {
		ip := httputil.GetIP(ctx.Request, bool(ctx.TrustProxy))
		err := ctx.RateLimiter.TakeToken(interaction.AccountEnumerationRateLimitBucket(ip))
		if err != nil {
			return nil, err
		}
	}

	var info *identity.Info
	info, err := ctx.Identities.GetBySpec(e.IdentitySpec)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		// nolint: ineffassign
		err = nil

		if e.IsAuthentication {
			switch e.IdentitySpec.Type {
			case authn.IdentityTypeOAuth:
				// This branch should be unreachable.
				break
			case authn.IdentityTypeAnonymous, authn.IdentityTypeBiometric:
				// Anonymous and biometric are handled in their own node.
				break
			case authn.IdentityTypeLoginID:
				loginIDValue := e.IdentitySpec.Claims[identity.IdentityClaimLoginIDValue].(string)
				err = ctx.Events.DispatchEvent(&nonblocking.AuthenticationFailedLoginIDEventPayload{
					LoginID: loginIDValue,
				})
				if err != nil {
					return nil, err
				}
			default:
				panic(fmt.Errorf("interaction: unknown identity type: %v", e.IdentitySpec.Type))
			}
		}
	} else if err != nil {
		return nil, err
	}

	return &NodeSelectIdentityEnd{
		IdentitySpec: e.IdentitySpec,
		IdentityInfo: info,
	}, nil
}

type NodeSelectIdentityEnd struct {
	IdentitySpec *identity.Spec `json:"identity_spec"`
	IdentityInfo *identity.Info `json:"identity_info"`
}

func (n *NodeSelectIdentityEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeSelectIdentityEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
