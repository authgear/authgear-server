package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type IntentSignupFlowStepAuthenticateTarget interface {
	GetOOBOTPClaims(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (map[model.ClaimName]string, error)
}

func init() {
	workflow.RegisterPrivateIntent(&IntentSignupFlowStepAuthenticate{})
}

var IntentSignupFlowStepAuthenticateSchema = validation.NewSimpleSchema(`{}`)

type IntentSignupFlowStepAuthenticate struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentSignupFlowStepAuthenticate{}

func (i *IntentSignupFlowStepAuthenticate) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentSignupFlowStepVerifyTarget = &IntentSignupFlowStepAuthenticate{}

func (*IntentSignupFlowStepAuthenticate) GetVerifiableClaims(_ context.Context, _ *workflow.Dependencies, workflows workflow.Workflows) (map[model.ClaimName]string, error) {
	n, ok := workflow.FindSingleNode[*NodeDoCreateAuthenticator](workflows.Nearest)
	if !ok {
		return nil, fmt.Errorf("NodeDoCreateAuthenticator cannot be found in IntentSignupFlowStepAuthenticate")
	}
	return n.Authenticator.StandardClaims(), nil
}

func (*IntentSignupFlowStepAuthenticate) GetPurpose(_ context.Context, _ *workflow.Dependencies, _ workflow.Workflows) otp.Purpose {
	return otp.PurposeOOBOTP
}

func (i *IntentSignupFlowStepAuthenticate) GetMessageType(_ context.Context, _ *workflow.Dependencies, workflows workflow.Workflows) otp.MessageType {
	authenticationMethod := i.authenticationMethod(workflows.Nearest)
	switch authenticationMethod {
	case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
		return otp.MessageTypeSetupPrimaryOOB
	case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
		return otp.MessageTypeSetupPrimaryOOB
	case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
		return otp.MessageTypeSetupSecondaryOOB
	case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
		return otp.MessageTypeSetupSecondaryOOB
	default:
		panic(fmt.Errorf("workflow: unexpected authentication method: %v", authenticationMethod))
	}
}

var _ workflow.Intent = &IntentSignupFlowStepAuthenticate{}

func (*IntentSignupFlowStepAuthenticate) Kind() string {
	return "workflowconfig.IntentSignupFlowStepAuthenticate"
}

func (*IntentSignupFlowStepAuthenticate) JSONSchema() *validation.SimpleSchema {
	return IntentSignupFlowStepAuthenticateSchema
}

func (*IntentSignupFlowStepAuthenticate) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// Let the input to select which authentication method to use.
	if len(workflows.Nearest.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeAuthenticationMethod{},
		}, nil
	}

	lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
	if lastNode.Type == workflow.NodeTypeSimple {
		switch lastNode.Simple.(type) {
		case *NodeDoCreateAuthenticator:
			// Handle nested steps.
			return nil, nil
		}
	}

	return nil, workflow.ErrEOF
}

func (i *IntentSignupFlowStepAuthenticate) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := signupFlowCurrent(deps, i.SignupFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(workflows.Nearest.Nodes) == 0 {
		var inputTakeAuthenticationMethod inputTakeAuthenticationMethod
		if workflow.AsInput(input, &inputTakeAuthenticationMethod) &&
			// NodeCreateAuthenticatorOOBOTP sometimes does not take any input to proceed when it has target_step.
			// In that case, if the next step is also type: authenticate, then the input will be incorrectly fed to the next step.
			// To protect against this, we require the first input of each step to provide the json pointer to indicate the audience of the input.
			inputTakeAuthenticationMethod.GetJSONPointer().String() == i.JSONPointer.String() {

			authentication := inputTakeAuthenticationMethod.GetAuthenticationMethod()
			var idx int
			idx, err = i.checkAuthenticationMethod(step, authentication)
			if err != nil {
				return nil, err
			}

			switch authentication {
			case config.WorkflowAuthenticationMethodPrimaryPassword:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryPassword:
				return workflow.NewNodeSimple(&NodeCreateAuthenticatorPassword{
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.WorkflowAuthenticationMethodPrimaryPasskey:
				// FIXME(workflow): create primary passkey
			case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
				fallthrough
			case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
				return workflow.NewNodeSimple(&NodeCreateAuthenticatorOOBOTP{
					SignupFlow:     i.SignupFlow,
					JSONPointer:    JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.WorkflowAuthenticationMethodSecondaryTOTP:
				node, err := NewNodeCreateAuthenticatorTOTP(deps, &NodeCreateAuthenticatorTOTP{
					UserID:         i.UserID,
					Authentication: authentication,
				})
				if err != nil {
					return nil, err
				}
				return workflow.NewNodeSimple(node), nil
			}
		}
		return nil, workflow.ErrIncompatibleInput
	}

	lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
	if lastNode.Type == workflow.NodeTypeSimple {
		switch lastNode.Simple.(type) {
		case *NodeDoCreateAuthenticator:
			authentication := i.authenticationMethod(workflows.Nearest)
			return workflow.NewSubWorkflow(&IntentSignupFlowSteps{
				SignupFlow:  i.SignupFlow,
				JSONPointer: i.jsonPointer(step, authentication),
				UserID:      i.UserID,
			}), nil
		}
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentSignupFlowStepAuthenticate) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentSignupFlowStepAuthenticate) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return map[string]interface{}{
		"json_pointer": i.JSONPointer.String(),
	}, nil
}

func (*IntentSignupFlowStepAuthenticate) step(o config.WorkflowObject) *config.WorkflowSignupFlowStep {
	step, ok := o.(*config.WorkflowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}

func (*IntentSignupFlowStepAuthenticate) checkAuthenticationMethod(step *config.WorkflowSignupFlowStep, am config.WorkflowAuthenticationMethod) (idx int, err error) {
	idx = -1
	var allAllowed []config.WorkflowAuthenticationMethod

	for i, branch := range step.OneOf {
		branch := branch
		allAllowed = append(allAllowed, branch.Authentication)
		if am == branch.Authentication {
			idx = i
		}
	}

	if idx >= 0 {
		return
	}

	err = InvalidAuthenticationMethod.NewWithInfo("invalid authentication method", apierrors.Details{
		"expected": allAllowed,
		"actual":   am,
	})
	return
}

func (*IntentSignupFlowStepAuthenticate) authenticationMethod(w *workflow.Workflow) config.WorkflowAuthenticationMethod {
	if len(w.Nodes) == 0 {
		panic(fmt.Errorf("workflow: authentication method not yet selected"))
	}

	switch n := w.Nodes[0].Simple.(type) {
	case *NodeCreateAuthenticatorPassword:
		return n.Authentication
	case *NodeCreateAuthenticatorOOBOTP:
		return n.Authentication
	case *NodeCreateAuthenticatorTOTP:
		return n.Authentication
	default:
		panic(fmt.Errorf("workflow: unexpected node: %T", w.Nodes[0].Simple))
	}
}

func (i *IntentSignupFlowStepAuthenticate) jsonPointer(step *config.WorkflowSignupFlowStep, am config.WorkflowAuthenticationMethod) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("workflow: selected identification method is not allowed"))
}
