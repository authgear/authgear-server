package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	newinteraction.RegisterNode(&NodeGenerateRecoveryCodeBegin{})
}

type EdgeGenerateRecoveryCode struct{}

func (e *EdgeGenerateRecoveryCode) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	// List all secondary authenticators and see how many of them are new.
	// If all of them are new, the user just enrolled into secondary authentication, we need to (re)generate recovery code for them.

	userID := graph.MustGetUserID()
	ais, err := ctx.Authenticators.List(
		userID,
		authenticator.KeepTag(authenticator.TagSecondaryAuthenticator),
	)
	if err != nil {
		return nil, err
	}

	newSecondary := filterAuthenticators(
		graph.GetUserNewAuthenticators(),
		authenticator.KeepTag(authenticator.TagSecondaryAuthenticator),
	)

	if len(newSecondary) != 0 && len(newSecondary) == len(ais) {
		recoveryCodes := ctx.MFA.GenerateRecoveryCodes()
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

func (n *NodeGenerateRecoveryCodeBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeGenerateRecoveryCodeBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeGenerateRecoveryCodeBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeGenerateRecoveryCodeEnd{RecoveryCodes: n.RecoveryCodes},
	}, nil
}

// GetRecoveryCodes implements RecoveryCodeNodes.
func (n *NodeGenerateRecoveryCodeBegin) GetRecoveryCodes() []string {
	return n.RecoveryCodes
}
