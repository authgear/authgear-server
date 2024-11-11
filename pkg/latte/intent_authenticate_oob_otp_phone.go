package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentAuthenticateOOBOTPPhone{})
}

var IntentAuthenticateOOBOTPPhoneSchema = validation.NewSimpleSchema(`{}`)

type IntentAuthenticateOOBOTPPhone struct {
	CaptchaProtectedIntent
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

func (i *IntentAuthenticateOOBOTPPhone) Kind() string {
	return "latte.IntentAuthenticateOOBOTPPhone"
}

func (i *IntentAuthenticateOOBOTPPhone) JSONSchema() *validation.SimpleSchema {
	return IntentAuthenticateOOBOTPPhoneSchema
}

func (i *IntentAuthenticateOOBOTPPhone) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	if i.IsCaptchaProtected {
		switch len(workflows.Nearest.Nodes) {
		case 0:
			return nil, nil
		case 1:
			return nil, nil
		}
	} else {
		switch len(workflows.Nearest.Nodes) {
		case 0:
			return nil, nil
		}
	}
	return nil, workflow.ErrEOF
}

func (i *IntentAuthenticateOOBOTPPhone) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	if i.IsCaptchaProtected && len(workflow.FindSubWorkflows[*IntentVerifyCaptcha](workflows.Nearest)) == 0 {
		return workflow.NewSubWorkflow(&IntentVerifyCaptcha{}), nil
	}

	if _, found := workflow.FindSingleNode[*NodeAuthenticateOOBOTPPhone](workflows.Nearest); !found {
		authenticator := i.Authenticator
		err := (&SendOOBCode{
			WorkflowID:        workflow.GetWorkflowID(ctx),
			Deps:              deps,
			Stage:             authenticatorKindToStage(authenticator.Kind),
			IsAuthenticating:  true,
			AuthenticatorInfo: authenticator,
			OTPForm:           otp.FormCode,
		}).Do(ctx)
		if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(&NodeAuthenticateOOBOTPPhone{
			Authenticator: authenticator,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentAuthenticateOOBOTPPhone) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentAuthenticateOOBOTPPhone) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
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
