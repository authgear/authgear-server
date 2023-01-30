package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var ErrNoPublicSignup = apierrors.Forbidden.WithReason("NoPublicSignup").New("public signup is disabled")

func init() {
	interaction.RegisterNode(&NodeDoCreateUser{})
}

type EdgeDoCreateUser struct {
}

func (e *EdgeDoCreateUser) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
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

	if !bypassRateLimit {
		// check the rate limit only to ensure that we have token to signup
		// the token will be token after running the effects successfully
		bucket := interaction.AntiSpamSignupBucket(string(ctx.RemoteIP))
		pass, _, err := ctx.RateLimiter.CheckToken(bucket)
		if err != nil {
			return nil, err
		}
		if !pass {
			return nil, bucket.BucketError()
		}
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

func (n *NodeDoCreateUser) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoCreateUser) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			_, err := ctx.Users.Create(n.CreateUserID)
			return err
		}),
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			u, err := ctx.Users.GetRaw(n.CreateUserID)
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

			if !n.BypassRateLimit && isAnonymous {
				// `graph.GetUserNewIdentities` is used to identify if the
				// new user is anonymous user.
				// Therefore this checking need to be done in `EffectOnCommit`
				// to ensure all nodes are run.
				// check the rate limit only before running the effects
				bucket := interaction.AntiSpamSignupAnonymousBucket(ip)
				pass, _, err := ctx.RateLimiter.CheckToken(bucket)
				if err != nil {
					return err
				}
				if !pass {
					return bucket.BucketError()
				}
			}

			webhookState := ""
			if intentWithWebhook, ok := graph.Intent.(interaction.IntentWithWebhookState); ok {
				webhookState = intentWithWebhook.GetWebhookState()
			}

			// run the effects
			err = ctx.Users.AfterCreate(
				u,
				graph.GetUserNewIdentities(),
				graph.GetUserNewAuthenticators(),
				n.IsAdminAPI,
				webhookState,
			)
			if err != nil {
				return err
			}

			// take the token after running the effects successfully
			if !n.BypassRateLimit {
				err := ctx.RateLimiter.TakeToken(interaction.AntiSpamSignupBucket(ip))
				if err != nil {
					return err
				}

				if isAnonymous {
					err := ctx.RateLimiter.TakeToken(interaction.AntiSpamSignupAnonymousBucket(ip))
					if err != nil {
						return err
					}
				}
			}
			return nil
		}),
	}, nil
}

func (n *NodeDoCreateUser) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoCreateUser) UserID() string {
	return n.CreateUserID
}

func (n *NodeDoCreateUser) NewUserID() string {
	return n.CreateUserID
}
