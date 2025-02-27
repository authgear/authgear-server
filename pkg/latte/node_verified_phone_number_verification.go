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
	IsCaptchaProtected bool           `json:"is_captcha_protected,omitempty"`
	IdentitySpec       *identity.Spec `json:"identity_spec,omitempty"`
}

func (n *NodeVerifiedProofOfPhoneNumberVerification) Kind() string {
	return "latte.NodeVerifiedProofOfPhoneNumberVerification"
}

func (n *NodeVerifiedProofOfPhoneNumberVerification) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeVerifiedProofOfPhoneNumberVerification) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	if len(workflows.Nearest.Nodes) == 1 {
		return nil, nil
	}

	return nil, workflow.ErrEOF
}

func (n *NodeVerifiedProofOfPhoneNumberVerification) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	identityInfo, _, err := deps.Identities.SearchBySpec(ctx, n.IdentitySpec)
	if err != nil {
		return nil, err
	}
	if identityInfo == nil {
		intent := &IntentSignup{
			PhoneNumberHint: n.IdentitySpec.LoginID.Value.TrimSpace(),
		}
		intent.IsCaptchaProtected = n.IsCaptchaProtected
		return workflow.NewSubWorkflow(intent), nil
	}

	intent := &IntentLogin{
		Identity:         identityInfo,
		IdentityVerified: true,
	}
	intent.IsCaptchaProtected = n.IsCaptchaProtected
	return workflow.NewSubWorkflow(intent), nil
}

func (n *NodeVerifiedProofOfPhoneNumberVerification) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return map[string]interface{}{}, nil
}
