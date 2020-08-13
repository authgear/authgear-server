package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/mfa"
)

func init() {
	newinteraction.RegisterNode(&NodeDoConsumeRecoveryCode{})
}

type InputConsumeRecoveryCode interface {
	GetRecoveryCode() string
}

type EdgeConsumeRecoveryCode struct{}

func (e *EdgeConsumeRecoveryCode) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputConsumeRecoveryCode)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	recoveryCode := input.GetRecoveryCode()

	rc, err := ctx.MFA.GetRecoveryCode(userID, recoveryCode)
	if errors.Is(err, mfa.ErrRecoveryCodeNotFound) {
		return &NodeAuthenticationEnd{
			Stage: newinteraction.AuthenticationStageSecondary,
		}, nil
	} else if errors.Is(err, mfa.ErrRecoveryCodeConsumed) {
		return &NodeAuthenticationEnd{
			Stage: newinteraction.AuthenticationStageSecondary,
		}, nil
	} else if err != nil {
		return nil, err
	}

	return &NodeDoConsumeRecoveryCode{RecoveryCode: rc}, nil
}

type NodeDoConsumeRecoveryCode struct {
	RecoveryCode *mfa.RecoveryCode `json:"recovery_code"`
}

func (n *NodeDoConsumeRecoveryCode) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoConsumeRecoveryCode) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		return ctx.MFA.ConsumeRecoveryCode(n.RecoveryCode)
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoConsumeRecoveryCode) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{&EdgeAuthenticationEnd{
		Stage:    newinteraction.AuthenticationStageSecondary,
		Optional: true,
	}}, nil
}
