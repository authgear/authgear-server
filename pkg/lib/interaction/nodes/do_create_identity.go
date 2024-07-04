package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

func init() {
	interaction.RegisterNode(&NodeDoCreateIdentity{})
}

type EdgeDoCreateIdentity struct {
	Identity   *identity.Info
	IsAddition bool
}

func (e *EdgeDoCreateIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	isAdminAPI := interaction.IsAdminAPI(rawInput)
	createDisabled := e.Identity.CreateDisabled(ctx.Config.Identity)
	if e.IsAddition && !isAdminAPI && createDisabled {
		return nil, api.ErrIdentityModifyDisabled
	}

	skipCreateIdentityEvent := false
	if _, creating := graph.GetNewUserID(); creating {
		skipCreateIdentityEvent = true
	} else {
		// not user signup
		// determine if the flow is user promotion by checking user's identities
		// this node need to be run before removing the anonymous identity
		iis, err := ctx.Identities.ListByUser(graph.MustGetUserID())
		if err != nil {
			return nil, err
		}
		for _, ii := range iis {
			if ii.Type == model.IdentityTypeAnonymous {
				// skip create identity event for anonymous user promotion
				skipCreateIdentityEvent = true
			}
		}
	}

	return &NodeDoCreateIdentity{
		Identity:                e.Identity,
		IsAddition:              e.IsAddition,
		IsAdminAPI:              isAdminAPI,
		SkipCreateIdentityEvent: skipCreateIdentityEvent,
	}, nil
}

type NodeDoCreateIdentity struct {
	Identity                *identity.Info `json:"identity"`
	IsAddition              bool           `json:"is_addition"`
	IsAdminAPI              bool           `json:"is_admin_api"`
	SkipCreateIdentityEvent bool           `json:"skip_create_identity_event"`
}

func (n *NodeDoCreateIdentity) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

// nolint:gocognit
func (n *NodeDoCreateIdentity) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			user, err := ctx.Users.Get(n.Identity.UserID, accesscontrol.RoleGreatest)
			if err != nil {
				return err
			}

			if n.Identity.Type == model.IdentityTypeBiometric && user.IsAnonymous {
				return api.ErrAnonymousUserAddIdentity
			}

			if existing, err := ctx.Identities.CheckDuplicated(n.Identity); err != nil {
				if errors.Is(err, identity.ErrIdentityAlreadyExists) {
					s1 := n.Identity.ToSpec()
					s2 := existing.ToSpec()
					return identityFillDetails(api.ErrDuplicatedIdentity, &s1, &s2)
				}
				return err
			}
			if err := ctx.Identities.Create(n.Identity); err != nil {
				return err
			}

			if !n.IsAddition && ctx.Config.UserProfile.StandardAttributes.Population.Strategy == config.StandardAttributesPopulationStrategyOnSignup {
				err := ctx.StdAttrsService.PopulateStandardAttributes(n.Identity.UserID, n.Identity)
				if err != nil {
					return err
				}
			}

			return nil
		}),
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.SkipCreateIdentityEvent {
				return nil
			}

			var e event.Payload
			switch n.Identity.Type {
			case model.IdentityTypeLoginID:
				loginIDType := n.Identity.LoginID.LoginIDType
				if payload, ok := nonblocking.NewIdentityLoginIDAddedEventPayload(
					model.UserRef{
						Meta: model.Meta{
							ID: n.Identity.UserID,
						},
					},
					n.Identity.ToModel(),
					string(loginIDType),
					n.IsAdminAPI,
				); ok {
					e = payload
				}
			case model.IdentityTypeOAuth:
				e = &nonblocking.IdentityOAuthConnectedEventPayload{
					UserRef: model.UserRef{
						Meta: model.Meta{
							ID: n.Identity.UserID,
						},
					},
					Identity: n.Identity.ToModel(),
					AdminAPI: n.IsAdminAPI,
				}
			case model.IdentityTypeBiometric:
				e = &nonblocking.IdentityBiometricEnabledEventPayload{
					UserRef: model.UserRef{
						Meta: model.Meta{
							ID: n.Identity.UserID,
						},
					},
					Identity: n.Identity.ToModel(),
					AdminAPI: n.IsAdminAPI,
				}
			}

			if e != nil {
				err := ctx.Events.DispatchEventOnCommit(e)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoCreateIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoCreateIdentity) UserNewIdentity() *identity.Info {
	return n.Identity
}
