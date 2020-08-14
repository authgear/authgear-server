package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
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
	input, ok := rawInput.(InputConsumeRecoveryCode)
	if !ok {
		return nil, interaction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	recoveryCode := input.GetRecoveryCode()

	rc, err := ctx.MFA.GetRecoveryCode(userID, recoveryCode)
	if errors.Is(err, mfa.ErrRecoveryCodeNotFound) {
		return &NodeAuthenticationEnd{
			Stage:  interaction.AuthenticationStageSecondary,
			Result: AuthenticationResultRecoveryCode,
		}, nil
	} else if errors.Is(err, mfa.ErrRecoveryCodeConsumed) {
		return &NodeAuthenticationEnd{
			Stage:  interaction.AuthenticationStageSecondary,
			Result: AuthenticationResultRecoveryCode,
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

func (n *NodeDoConsumeRecoveryCode) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	err := perform(interaction.EffectRun(func(ctx *interaction.Context) error {
		return ctx.MFA.ConsumeRecoveryCode(n.RecoveryCode)
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoConsumeRecoveryCode) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeAuthenticationEnd{
		Stage:        interaction.AuthenticationStageSecondary,
		Result:       AuthenticationResultRecoveryCode,
		RecoveryCode: n.RecoveryCode,
	}}, nil
}
