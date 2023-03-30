package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentAuthenticateOOBOTPPhone{})
}

var IntentAuthenticateOOBOTPPhoneSchema = validation.NewSimpleSchema(`{}`)

type IntentAuthenticateOOBOTPPhone struct {
	Authenticator      *authenticator.Info `json:"authenticator,omitempty"`
	IsCaptchaProtected bool                `json:"is_captcha_protected,omitempty"`
}

func (i *IntentAuthenticateOOBOTPPhone) Kind() string {
	return "latte.IntentAuthenticateOOBOTPPhone"
}

func (i *IntentAuthenticateOOBOTPPhone) JSONSchema() *validation.SimpleSchema {
	return IntentAuthenticateOOBOTPPhoneSchema
}

func (i *IntentAuthenticateOOBOTPPhone) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	switch len(w.Nodes) {
	case 0:
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentAuthenticateOOBOTPPhone) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	if i.IsCaptchaProtected && len(workflow.FindSubWorkflows[*IntentVerifyCaptcha](w)) == 0 {
		return workflow.NewSubWorkflow(&IntentVerifyCaptcha{}), nil
	}

	if _, found := workflow.FindSingleNode[*NodeAuthenticateOOBOTPPhone](w); !found {
		authenticator := i.Authenticator
		_, err := (&SendOOBCode{
			WorkflowID:        workflow.GetWorkflowID(ctx),
			Deps:              deps,
			Stage:             authenticatorKindToStage(authenticator.Kind),
			IsAuthenticating:  true,
			AuthenticatorInfo: authenticator,
		}).Do()
		if ratelimit.IsRateLimitErrorWithBucketName(err, "AntiSpamOTPCodeBucket") {
			// Ignore resend cooldown rate limit error for initial sending
		} else if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(&NodeAuthenticateOOBOTPPhone{
			Authenticator: authenticator,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentAuthenticateOOBOTPPhone) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentAuthenticateOOBOTPPhone) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{}, nil
}

func (i *IntentAuthenticateOOBOTPPhone) GetAMR(w *workflow.Workflow) []string {
	node, ok := workflow.FindSingleNode[*NodeVerifiedAuthenticator](w)
	if !ok {
		return []string{}
	}
	return node.GetAMR()
}

var _ AMRGetter = &IntentAuthenticateOOBOTPPhone{}
