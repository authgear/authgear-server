package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoConsumeRecoveryCode{})
}

type InputConsumeRecoveryCode interface {
	GetRecoveryCode() string
}

type EdgeConsumeRecoveryCode struct{}

func (e *EdgeConsumeRecoveryCode) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputConsumeRecoveryCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	recoveryCode := input.GetRecoveryCode()

	rc, err := ctx.MFA.VerifyRecoveryCode(userID, recoveryCode)
	if errors.Is(err, api.ErrInvalidCredentials) {
		return &NodeAuthenticationEnd{
			Stage:              authn.AuthenticationStageSecondary,
			AuthenticationType: authn.AuthenticationTypeRecoveryCode,
		}, nil
	} else if err != nil {
		return nil, err
	}

	return &NodeDoConsumeRecoveryCode{RecoveryCode: rc}, nil
}

type NodeDoConsumeRecoveryCode struct {
	RecoveryCode *mfa.RecoveryCode `json:"recovery_code"`
}

func (n *NodeDoConsumeRecoveryCode) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoConsumeRecoveryCode) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			return ctx.MFA.ConsumeRecoveryCode(n.RecoveryCode)
		}),
	}, nil
}

func (n *NodeDoConsumeRecoveryCode) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeAuthenticationEnd{
		Stage:              authn.AuthenticationStageSecondary,
		AuthenticationType: authn.AuthenticationTypeRecoveryCode,
		RecoveryCode:       n.RecoveryCode,
	}}, nil
}

func (n *NodeDoConsumeRecoveryCode) UsedAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool) {
	return config.AuthenticationLockoutMethodRecoveryCode, true
}
