package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeDoUpdateAuthenticator{})
}

type EdgeDoUpdateAuthenticator struct {
	Stage                     newinteraction.AuthenticationStage
	AuthenticatorBeforeUpdate *authenticator.Info
	AuthenticatorAfterUpdate  *authenticator.Info
	PasswordUpdateReason      event.PasswordUpdateReason
}

func (e *EdgeDoUpdateAuthenticator) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoUpdateAuthenticator{
		AuthenticatorBeforeUpdate: e.AuthenticatorBeforeUpdate,
		AuthenticatorAfterUpdate:  e.AuthenticatorAfterUpdate,
	}, nil
}

type NodeDoUpdateAuthenticator struct {
	Stage                     newinteraction.AuthenticationStage `json:"stage"`
	AuthenticatorBeforeUpdate *authenticator.Info                `json:"authenticator_before_update"`
	AuthenticatorAfterUpdate  *authenticator.Info                `json:"authenticator_after_update"`
	PasswordUpdateReason      event.PasswordUpdateReason         `json:"password_update_reason"`
}

func (n *NodeDoUpdateAuthenticator) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		return ctx.Authenticators.Update(n.AuthenticatorAfterUpdate)
	}))
	if err != nil {
		return err
	}

	err = perform(newinteraction.EffectOnCommit(func(ctx *newinteraction.Context) error {
		switch n.AuthenticatorAfterUpdate.Type {
		case authn.AuthenticatorTypePassword:
			user, err := ctx.Users.Get(n.AuthenticatorAfterUpdate.UserID)
			if err != nil {
				return err
			}

			err = ctx.Hooks.DispatchEvent(
				event.PasswordUpdateEvent{
					Reason: n.PasswordUpdateReason,
					User:   *user,
				},
				user,
			)
			if err != nil {
				return err
			}

			return nil
		default:
			return nil
		}
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoUpdateAuthenticator) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeDoUpdateAuthenticator) UserAuthenticator() (newinteraction.AuthenticationStage, *authenticator.Info) {
	return n.Stage, n.AuthenticatorAfterUpdate
}
