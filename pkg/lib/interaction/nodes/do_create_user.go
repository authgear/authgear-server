package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var ErrNoPublicSignup = apierrors.Forbidden.WithReason("NoPublicSignup").New("public signup is disabled")

func init() {
	interaction.RegisterNode(&NodeDoCreateUser{})
}

type EdgeDoCreateUser struct {
}

func (e *EdgeDoCreateUser) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	publicSignupDisabled := ctx.Config.Authentication.PublicSignupDisabled

	bypassPublicSignupDisabled := false
	var bypassPublicSignupDisabledInput interface{ BypassPublicSignupDisabled() bool }
	if interaction.Input(rawInput, &bypassPublicSignupDisabledInput) && bypassPublicSignupDisabledInput.BypassPublicSignupDisabled() {
		bypassPublicSignupDisabled = true
	}

	allowed := !publicSignupDisabled || bypassPublicSignupDisabled
	if !allowed {
		return nil, ErrNoPublicSignup
	}

	bypassRateLimit := false
	var bypassInput interface{ BypassInteractionIPRateLimit() bool }
	if interaction.Input(rawInput, &bypassInput) {
		bypassRateLimit = bypassInput.BypassInteractionIPRateLimit()
	}

	return &NodeDoCreateUser{
		CreateUserID:    uuid.New(),
		BypassRateLimit: bypassRateLimit,
		IsAdminAPI:      interaction.IsAdminAPI(rawInput),
	}, nil
}

type NodeDoCreateUser struct {
	CreateUserID    string `json:"create_user_id"`
	BypassRateLimit bool   `json:"bypass_rate_limit"`
	IsAdminAPI      bool   `json:"is_admin_api"`
}

func (n *NodeDoCreateUser) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoCreateUser) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			_, err := ctx.Users.Create(goCtx, n.CreateUserID)
			return err
		}),
		interaction.EffectOnCommit(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			u, err := ctx.Users.GetRaw(goCtx, n.CreateUserID)
			if err != nil {
				return err
			}

			newIdentities := graph.GetUserNewIdentities()
			ip := string(ctx.RemoteIP)
			isAnonymous := false
			for _, i := range newIdentities {
				if i.Type == model.IdentityTypeAnonymous {
					isAnonymous = true
				}
			}

			var reservation *ratelimit.Reservation
			if !n.BypassRateLimit {
				// `graph.GetUserNewIdentities` is used to identify if the
				// new user is anonymous user.
				// Therefore this checking need to be done in `EffectOnCommit`
				// to ensure all nodes are run.

				// check the rate limit only before running the effects
				bucket := interaction.SignupPerIPRateLimitBucketSpec(ctx.Config.Authentication, isAnonymous, ip)
				var failedReservation *ratelimit.FailedReservation
				reservation, failedReservation, err = ctx.RateLimiter.Reserve(goCtx, bucket)
				if err != nil {
					return err
				}
				if err := failedReservation.Error(); err != nil {
					return err
				}
			}
			defer ctx.RateLimiter.Cancel(goCtx, reservation)

			// run the effects
			err = ctx.Users.AfterCreate(goCtx,
				u,
				graph.GetUserNewIdentities(),
				graph.GetUserNewAuthenticators(),
				n.IsAdminAPI,
			)
			if err != nil {
				return err
			}
			if reservation != nil {
				reservation.PreventCancel()
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoCreateUser) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}

func (n *NodeDoCreateUser) UserID() string {
	return n.CreateUserID
}

func (n *NodeDoCreateUser) NewUserID() string {
	return n.CreateUserID
}
