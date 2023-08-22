package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

type IntentSignupFlowStepVerifyTarget interface {
	GetVerifiableClaims(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (map[model.ClaimName]string, error)
	GetPurpose(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) otp.Purpose
	GetMessageType(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) otp.MessageType
}

func init() {
	workflow.RegisterIntent(&IntentSignupFlowStepVerify{})
}

type IntentSignupFlowStepVerify struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentSignupFlowStepVerify{}

func (i *IntentSignupFlowStepVerify) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepVerify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ workflow.Intent = &IntentSignupFlowStepVerify{}

func (*IntentSignupFlowStepVerify) Kind() string {
	return "workflowconfig.IntentSignupFlowStepVerify"
}

// This intent is NOT a boundary because it reacts to nil input.
// If it were a boundary, the returned flow would stick at wait for nil input.
// var _ workflow.Boundary = &IntentSignupFlowStepVerify{}
// func (i *IntentSignupFlowStepVerify) Boundary() string {
// 	return i.JSONPointer.String()
// }

func (*IntentSignupFlowStepVerify) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
	// Look up the claim to verify
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentSignupFlowStepVerify) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, _ workflow.Input) (*workflow.Node, error) {
	current, err := signupFlowCurrent(deps, i.SignupFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}

	step := i.step(current)
	targetStepID := step.TargetStep

	// Find the target step from the root.
	targetStepWorkflow, err := FindTargetStep(workflows.Root, targetStepID)
	if err != nil {
		return nil, err
	}

	target, ok := targetStepWorkflow.Intent.(IntentSignupFlowStepVerifyTarget)
	if !ok {
		return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
			"target_step": targetStepID,
		})
	}

	claims, err := target.GetVerifiableClaims(ctx, deps, workflows.Replace(targetStepWorkflow))
	if err != nil {
		return nil, err
	}

	if len(claims) == 0 {
		// Nothing to verify. End this workflow.
		return workflow.NewNodeSimple(&NodeSentinel{}), nil
	}

	var claimNames []model.ClaimName
	for claimName := range claims {
		claimNames = append(claimNames, claimName)
	}

	if len(claimNames) > 1 {
		// TODO(workflow): support verify more than 1 claim?
		return nil, InvalidTargetStep.NewWithInfo("target_step contains more than one claim to verify", apierrors.Details{
			"claims": claimNames,
		})
	}

	claimName := claimNames[0]
	switch claimName {
	case model.ClaimEmail:
		break
	case model.ClaimPhoneNumber:
		break
	default:
		return nil, InvalidTargetStep.NewWithInfo("target_step contains a claim that cannot be verified", apierrors.Details{
			"claim_name": claimName,
		})
	}

	purpose := target.GetPurpose(ctx, deps, workflows.Replace(targetStepWorkflow))
	messageType := target.GetMessageType(ctx, deps, workflows.Replace(targetStepWorkflow))
	claimValue := claims[claimName]
	return workflow.NewSubWorkflow(&IntentVerifyClaim{
		UserID:      i.UserID,
		Purpose:     purpose,
		MessageType: messageType,
		ClaimName:   claimName,
		ClaimValue:  claimValue,
	}), nil
}

func (*IntentSignupFlowStepVerify) step(o config.WorkflowObject) *config.WorkflowSignupFlowStep {
	step, ok := o.(*config.WorkflowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}
