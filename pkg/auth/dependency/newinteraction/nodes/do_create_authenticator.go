package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeDoCreateAuthenticator{})
}

type EdgeDoCreateAuthenticator struct {
	Stage                newinteraction.AuthenticationStage
	Authenticators       []*authenticator.Info
	PasswordUpdateReason event.PasswordUpdateReason
}

func (e *EdgeDoCreateAuthenticator) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoCreateAuthenticator{
		Stage:          e.Stage,
		Authenticators: e.Authenticators,
	}, nil
}

type NodeDoCreateAuthenticator struct {
	Stage                newinteraction.AuthenticationStage `json:"stage"`
	Authenticators       []*authenticator.Info              `json:"authenticators"`
	PasswordUpdateReason event.PasswordUpdateReason         `json:"password_update_reason"`
}

func (n *NodeDoCreateAuthenticator) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		for _, a := range n.Authenticators {
			if err := ctx.Authenticators.Create(a); err != nil {
				return err
			}
		}

		return nil
	}))
	if err != nil {
		return err
	}

	// Run hooks only if not creating user
	if _, ok := graph.GetNewUserID(); !ok {
		err = perform(newinteraction.EffectOnCommit(func(ctx *newinteraction.Context) error {
			for _, a := range n.Authenticators {
				switch a.Type {
				case authn.AuthenticatorTypePassword:
					user, err := ctx.Users.Get(a.UserID)
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

				default:
					break
				}
			}
			return nil
		}))
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *NodeDoCreateAuthenticator) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeDoCreateAuthenticator) UserAuthenticator() (newinteraction.AuthenticationStage, *authenticator.Info) {
	if len(n.Authenticators) > 1 {
		panic("interaction: expect at most one primary/secondary authenticator")
	}
	if len(n.Authenticators) == 0 {
		return "", nil
	}
	return n.Stage, n.Authenticators[0]
}

func (n *NodeDoCreateAuthenticator) UserNewAuthenticators() []*authenticator.Info {
	return n.Authenticators
}
