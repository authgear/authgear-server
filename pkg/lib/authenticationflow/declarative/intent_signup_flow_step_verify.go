package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type IntentSignupFlowStepVerifyTarget interface {
	GetVerifiableClaims(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error)
	GetPurpose(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) otp.Purpose
	GetMessageType(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) otp.MessageType
}

func init() {
	authflow.RegisterIntent(&IntentSignupFlowStepVerify{})
}

type IntentSignupFlowStepVerify struct {
	StepID      string        `json:"step_id,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ FlowStep = &IntentSignupFlowStepVerify{}

func (i *IntentSignupFlowStepVerify) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepVerify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentSignupFlowStepVerify{}

func (*IntentSignupFlowStepVerify) Kind() string {
	return "IntentSignupFlowStepVerify"
}

func (*IntentSignupFlowStepVerify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// Look up the claim to verify
	if len(flows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentSignupFlowStepVerify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}

	step := i.step(current)
	targetStepID := step.TargetStep

	// Find the target step from the root.
	targetStepFlow, err := FindTargetStep(flows.Root, targetStepID)
	if err != nil {
		return nil, err
	}

	target, ok := targetStepFlow.Intent.(IntentSignupFlowStepVerifyTarget)
	if !ok {
		return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
			"target_step": targetStepID,
		})
	}

	claims, err := target.GetVerifiableClaims(ctx, deps, flows.Replace(targetStepFlow))
	if err != nil {
		return nil, err
	}

	if len(claims) == 0 {
		// Nothing to verify. End this flow.
		return authflow.NewNodeSimple(&NodeSentinel{}), nil
	}

	var claimNames []model.ClaimName
	for claimName := range claims {
		claimNames = append(claimNames, claimName)
	}

	if len(claimNames) > 1 {
		// TODO(authflow): support verify more than 1 claim?
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

	purpose := target.GetPurpose(ctx, deps, flows.Replace(targetStepFlow))
	messageType := target.GetMessageType(ctx, deps, flows.Replace(targetStepFlow))
	claimValue := claims[claimName]
	return authflow.NewSubFlow(&IntentVerifyClaim{
		JSONPointer: i.JSONPointer,
		UserID:      i.UserID,
		Purpose:     purpose,
		MessageType: messageType,
		ClaimName:   claimName,
		ClaimValue:  claimValue,
	}), nil
}

func (*IntentSignupFlowStepVerify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}
