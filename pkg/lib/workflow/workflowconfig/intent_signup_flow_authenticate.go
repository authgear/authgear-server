package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentSignupFlowAuthenticate{})
}

var IntentSignupFlowAuthenticateSchema = validation.NewSimpleSchema(`{}`)

type IntentSignupFlowAuthenticate struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentSignupFlowAuthenticate{}

func (i *IntentSignupFlowAuthenticate) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentSignupFlowVerifyTarget = &IntentSignupFlowAuthenticate{}

func (*IntentSignupFlowAuthenticate) GetVerifiableClaims(w *workflow.Workflow) (map[model.ClaimName]string, error) {
	n, ok := workflow.FindSingleNode[*NodeDoCreateAuthenticator](w)
	if !ok {
		return nil, fmt.Errorf("NodeDoCreateAuthenticator cannot be found in IntentSignupFlowAuthenticate")
	}
	return n.Authenticator.StandardClaims(), nil
}

var _ workflow.Intent = &IntentSignupFlowAuthenticate{}

func (*IntentSignupFlowAuthenticate) Kind() string {
	return "workflowconfig.IntentSignupFlowAuthenticate"
}

func (*IntentSignupFlowAuthenticate) JSONSchema() *validation.SimpleSchema {
	return IntentSignupFlowAuthenticateSchema
}

func (*IntentSignupFlowAuthenticate) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
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

func (i *IntentSignupFlowAuthenticate) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := i.current(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(workflows.Nearest.Nodes) == 0 {
		var inputTakeAuthenticationMethod inputTakeAuthenticationMethod
		if workflow.AsInput(input, &inputTakeAuthenticationMethod) {
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
				return workflow.NewNodeSimple(&NodeCreatePassword{
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
				return workflow.NewNodeSimple(&NodeCreateOOBOTPAuthenticator{
					SignupFlow:     i.SignupFlow,
					JSONPointer:    JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.WorkflowAuthenticationMethodSecondaryTOTP:
				// FIXME(workflow): create secondary totp
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

func (*IntentSignupFlowAuthenticate) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentSignupFlowAuthenticate) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *IntentSignupFlowAuthenticate) current(deps *workflow.Dependencies) (config.WorkflowObject, error) {
	root, err := findSignupFlow(deps.Config.Workflow, i.SignupFlow)
	if err != nil {
		return nil, err
	}

	entries, err := Traverse(root, i.JSONPointer)
	if err != nil {
		return nil, err
	}

	current, err := GetCurrentObject(entries)
	if err != nil {
		return nil, err
	}

	return current, nil
}

func (*IntentSignupFlowAuthenticate) step(o config.WorkflowObject) *config.WorkflowSignupFlowStep {
	step, ok := o.(*config.WorkflowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}

func (*IntentSignupFlowAuthenticate) checkAuthenticationMethod(step *config.WorkflowSignupFlowStep, am config.WorkflowAuthenticationMethod) (idx int, err error) {
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

func (*IntentSignupFlowAuthenticate) authenticationMethod(w *workflow.Workflow) config.WorkflowAuthenticationMethod {
	if len(w.Nodes) == 0 {
		panic(fmt.Errorf("workflow: authentication method not yet selected"))
	}

	switch n := w.Nodes[0].Simple.(type) {
	case *NodeCreatePassword:
		return n.Authentication
	case *NodeCreateOOBOTPAuthenticator:
		return n.Authentication
	default:
		panic(fmt.Errorf("workflow: unexpected node: %T", w.Nodes[0].Simple))
	}
}

func (i *IntentSignupFlowAuthenticate) jsonPointer(step *config.WorkflowSignupFlowStep, am config.WorkflowAuthenticationMethod) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("workflow: selected identification method is not allowed"))
}
