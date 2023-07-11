package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentSignupFlowIdentify{})
}

var IntentSignupFlowIdentifySchema = validation.NewSimpleSchema(`{}`)

type IntentSignupFlowIdentify struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentSignupFlowIdentify{}

func (i *IntentSignupFlowIdentify) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentSignupFlowVerifyTarget = &IntentSignupFlowIdentify{}

func (*IntentSignupFlowIdentify) GetVerifiableClaims(_ context.Context, _ *workflow.Dependencies, workflows workflow.Workflows) (map[model.ClaimName]string, error) {
	n, ok := workflow.FindSingleNode[*NodeDoCreateIdentity](workflows.Nearest)
	if !ok {
		return nil, fmt.Errorf("NodeDoCreateIdentity cannot be found in IntentSignupFlowIdentify")
	}
	return n.Identity.IdentityAwareStandardClaims(), nil
}

func (*IntentSignupFlowIdentify) GetPurpose(_ context.Context, _ *workflow.Dependencies, _ workflow.Workflows) otp.Purpose {
	return otp.PurposeVerification
}

func (*IntentSignupFlowIdentify) GetMessageType(_ context.Context, _ *workflow.Dependencies, _ workflow.Workflows) otp.MessageType {
	return otp.MessageTypeVerification
}

var _ IntentSignupFlowAuthenticateTarget = &IntentSignupFlowIdentify{}

func (n *IntentSignupFlowIdentify) GetOOBOTPClaims(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (map[model.ClaimName]string, error) {
	return n.GetVerifiableClaims(ctx, deps, workflows)
}

var _ workflow.Intent = &IntentSignupFlowIdentify{}

func (*IntentSignupFlowIdentify) Kind() string {
	return "workflowconfig.IntentSignupFlowIdentify"
}

func (*IntentSignupFlowIdentify) JSONSchema() *validation.SimpleSchema {
	return IntentSignupFlowIdentifySchema
}

func (*IntentSignupFlowIdentify) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// Let the input to select which identification method to use.
	if len(workflows.Nearest.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeIdentificationMethod{},
		}, nil
	}

	lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
	if lastNode.Type == workflow.NodeTypeSimple {
		switch lastNode.Simple.(type) {
		case *NodeDoCreateIdentity:
			// Populate standard attributes
			return nil, nil
		case *NodePopulateStandardAttributes:
			// Handle nested steps.
			return nil, nil
		}
	}

	return nil, workflow.ErrEOF
}

func (i *IntentSignupFlowIdentify) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := i.current(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(workflows.Nearest.Nodes) == 0 {
		var inputTakeIdentificationMethod inputTakeIdentificationMethod
		if workflow.AsInput(input, &inputTakeIdentificationMethod) {
			identification := inputTakeIdentificationMethod.GetIdentificationMethod()
			err = i.checkIdentificationMethod(step, identification)
			if err != nil {
				return nil, err
			}

			switch identification {
			case config.WorkflowIdentificationMethodEmail:
				fallthrough
			case config.WorkflowIdentificationMethodPhone:
				fallthrough
			case config.WorkflowIdentificationMethodUsername:
				return workflow.NewNodeSimple(&NodeCreateIdentityLoginID{
					UserID:         i.UserID,
					Identification: identification,
				}), nil
			case config.WorkflowIdentificationMethodOAuth:
				// FIXME(workflow): handle oauth
			case config.WorkflowIdentificationMethodPasskey:
				// FIXME(workflow): handle passkey
			case config.WorkflowIdentificationMethodSiwe:
				// FIXME(workflow): handle siwe
			}
		}
		return nil, workflow.ErrIncompatibleInput
	}

	lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
	if lastNode.Type == workflow.NodeTypeSimple {
		switch lastNode.Simple.(type) {
		case *NodeDoCreateIdentity:
			iden := i.identityInfo(workflows.Nearest)
			return workflow.NewNodeSimple(&NodePopulateStandardAttributes{
				Identity: iden,
			}), nil
		case *NodePopulateStandardAttributes:
			identification := i.identificationMethod(workflows.Nearest)
			return workflow.NewSubWorkflow(&IntentSignupFlowSteps{
				SignupFlow:  i.SignupFlow,
				JSONPointer: i.jsonPointer(step, identification),
				UserID:      i.UserID,
			}), nil
		}
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentSignupFlowIdentify) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentSignupFlowIdentify) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *IntentSignupFlowIdentify) current(deps *workflow.Dependencies) (config.WorkflowObject, error) {
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

func (*IntentSignupFlowIdentify) step(o config.WorkflowObject) *config.WorkflowSignupFlowStep {
	step, ok := o.(*config.WorkflowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}

func (*IntentSignupFlowIdentify) checkIdentificationMethod(step *config.WorkflowSignupFlowStep, im config.WorkflowIdentificationMethod) error {
	var allAllowed []config.WorkflowIdentificationMethod

	for _, branch := range step.OneOf {
		branch := branch
		allAllowed = append(allAllowed, branch.Identification)
	}

	for _, allowed := range allAllowed {
		if im == allowed {
			return nil
		}
	}

	return InvalidIdentificationMethod.NewWithInfo("invalid identification method", apierrors.Details{
		"expected": allAllowed,
		"actual":   im,
	})
}

func (*IntentSignupFlowIdentify) identificationMethod(w *workflow.Workflow) config.WorkflowIdentificationMethod {
	if len(w.Nodes) == 0 {
		panic(fmt.Errorf("workflow: identification method not yet selected"))
	}

	switch n := w.Nodes[0].Simple.(type) {
	case *NodeCreateIdentityLoginID:
		return n.Identification
	default:
		panic(fmt.Errorf("workflow: unexpected node: %T", w.Nodes[0].Simple))
	}
}

func (i *IntentSignupFlowIdentify) jsonPointer(step *config.WorkflowSignupFlowStep, im config.WorkflowIdentificationMethod) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("workflow: selected identification method is not allowed"))
}

func (*IntentSignupFlowIdentify) identityInfo(w *workflow.Workflow) *identity.Info {
	node, ok := workflow.FindSingleNode[*NodeDoCreateIdentity](w)
	if !ok {
		panic(fmt.Errorf("workflow: expected NodeCreateIdentity"))
	}
	return node.Identity
}
