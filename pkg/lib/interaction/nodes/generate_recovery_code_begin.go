package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeGenerateRecoveryCodeBegin{})
}

type EdgeGenerateRecoveryCode struct {
	IsRegenerate bool
}

func (e *EdgeGenerateRecoveryCode) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	// Regenerate recovery codes if explicitly requested
	doGenerate := e.IsRegenerate

	if !doGenerate {
		// List all secondary authenticators and see how many of them are new.
		// If all of them are new, the user just enrolled into secondary authentication, we need to (re)generate recovery code for them.

		userID := graph.MustGetUserID()
		ais, err := ctx.Authenticators.List(goCtx,
			userID,
			authenticator.KeepKind(authenticator.KindSecondary),
		)
		if err != nil {
			return nil, err
		}

		newSecondary := authenticator.ApplyFilters(
			graph.GetUserNewAuthenticators(),
			authenticator.KeepKind(authenticator.KindSecondary),
		)

		doGenerate = len(newSecondary) != 0 && len(newSecondary) == len(ais)
	}

	if doGenerate && !*ctx.Config.Authentication.RecoveryCode.Disabled {
		recoveryCodes := ctx.MFA.GenerateRecoveryCodes(goCtx)
		return &NodeGenerateRecoveryCodeBegin{
			RecoveryCodes: recoveryCodes,
		}, nil
	}

	// Otherwise just end it.
	return &NodeGenerateRecoveryCodeEnd{}, nil
}

type NodeGenerateRecoveryCodeBegin struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

func (n *NodeGenerateRecoveryCodeBegin) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeGenerateRecoveryCodeBegin) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeGenerateRecoveryCodeBegin) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeGenerateRecoveryCodeEnd{RecoveryCodes: n.RecoveryCodes},
	}, nil
}

// GetRecoveryCodes implements RecoveryCodeNodes.
func (n *NodeGenerateRecoveryCodeBegin) GetRecoveryCodes() []string {
	return n.RecoveryCodes
}
