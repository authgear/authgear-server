package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeVerifiedProofOfPhoneNumberVerification{})
}

type NodeVerifiedProofOfPhoneNumberVerification struct {
	IdentitySpec *identity.Spec `json:"identity_spec,omitempty"`
}

func (n *NodeVerifiedProofOfPhoneNumberVerification) Kind() string {
	return "latte.NodeVerifiedProofOfPhoneNumberVerification"
}

func (n *NodeVerifiedProofOfPhoneNumberVerification) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeVerifiedProofOfPhoneNumberVerification) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (n *NodeVerifiedProofOfPhoneNumberVerification) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeVerifiedProofOfPhoneNumberVerification) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return map[string]interface{}{}, nil
}
