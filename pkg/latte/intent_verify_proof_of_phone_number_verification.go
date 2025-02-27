package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentVerifyProofOfPhoneNumberVerification{})
}

var IntentVerifyProofOfPhoneNumberVerificationSchema = validation.NewSimpleSchema(`{}`)

type IntentVerifyProofOfPhoneNumberVerification struct {
	ProofOfPhoneNumberVerification string
}

func (i *IntentVerifyProofOfPhoneNumberVerification) Kind() string {
	return "latte.IntentVerifyProofOfPhoneNumberVerification"
}

func (i *IntentVerifyProofOfPhoneNumberVerification) JSONSchema() *validation.SimpleSchema {
	return IntentVerifyProofOfPhoneNumberVerificationSchema
}

func (i *IntentVerifyProofOfPhoneNumberVerification) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentVerifyProofOfPhoneNumberVerification) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	resp, err := deps.ProofOfPhoneNumberVerification.Verify(ctx, i.ProofOfPhoneNumberVerification)
	if err != nil {
		return nil, err
	}

	return workflow.NewNodeSimple(&NodeVerifiedProofOfPhoneNumberVerification{
		IdentitySpec: resp.Identity,
	}), nil
}

func (i *IntentVerifyProofOfPhoneNumberVerification) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentVerifyProofOfPhoneNumberVerification) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return map[string]interface{}{}, nil
}
