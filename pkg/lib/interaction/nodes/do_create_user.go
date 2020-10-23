package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
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

	return &NodeDoCreateUser{
		CreateUserID: uuid.New(),
	}, nil
}

type NodeDoCreateUser struct {
	CreateUserID string `json:"create_user_id"`
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
			return ctx.Users.AfterCreate(u, graph.GetUserNewIdentities())
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
