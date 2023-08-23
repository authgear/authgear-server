package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentSignupFlowStepIdentify{})
}

type IntentSignupFlowStepIdentify struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ FlowStep = &IntentSignupFlowStepIdentify{}

func (i *IntentSignupFlowStepIdentify) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentSignupFlowStepVerifyTarget = &IntentSignupFlowStepIdentify{}

func (*IntentSignupFlowStepIdentify) GetVerifiableClaims(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error) {
	m, ok := authflow.FindMilestone[MilestoneDoCreateIdentity](flows.Nearest)
	if !ok {
		return nil, fmt.Errorf("MilestoneDoCreateIdentity cannot be found in IntentSignupFlowStepIdentify")
	}
	info := m.MilestoneDoCreateIdentity()

	return info.IdentityAwareStandardClaims(), nil
}

func (*IntentSignupFlowStepIdentify) GetPurpose(_ context.Context, _ *authflow.Dependencies, _ authflow.Flows) otp.Purpose {
	return otp.PurposeVerification
}

func (*IntentSignupFlowStepIdentify) GetMessageType(_ context.Context, _ *authflow.Dependencies, _ authflow.Flows) otp.MessageType {
	return otp.MessageTypeVerification
}

var _ IntentSignupFlowStepAuthenticateTarget = &IntentSignupFlowStepIdentify{}

func (n *IntentSignupFlowStepIdentify) GetOOBOTPClaims(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error) {
	return n.GetVerifiableClaims(ctx, deps, flows)
}

var _ authflow.Intent = &IntentSignupFlowStepIdentify{}
var _ authflow.Boundary = &IntentSignupFlowStepIdentify{}

func (*IntentSignupFlowStepIdentify) Kind() string {
	return "IntentSignupFlowStepIdentify"
}

func (i *IntentSignupFlowStepIdentify) Boundary() string {
	return i.JSONPointer.String()
}

func (i *IntentSignupFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	current, err := signupFlowCurrent(deps, i.SignupFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	// Let the input to select which identification method to use.
	if len(flows.Nearest.Nodes) == 0 {
		return &InputSchemaSignupFlowStepIdentify{
			OneOf: step.OneOf,
		}, nil
	}

	_, identityCreated := authflow.FindMilestone[MilestoneDoCreateIdentity](flows.Nearest)
	_, standardAttributesPopulated := authflow.FindMilestone[MilestoneDoPopulateStandardAttributes](flows.Nearest)
	_, nestedStepHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case identityCreated && !standardAttributesPopulated && !nestedStepHandled:
		// Populate standard attributes
		return nil, nil
	case identityCreated && standardAttributesPopulated && !nestedStepHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentSignupFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := signupFlowCurrent(deps, i.SignupFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeIdentificationMethod inputTakeIdentificationMethod
		if authflow.AsInput(input, &inputTakeIdentificationMethod) {
			identification := inputTakeIdentificationMethod.GetIdentificationMethod()

			switch identification {
			case config.AuthenticationFlowIdentificationEmail:
				fallthrough
			case config.AuthenticationFlowIdentificationPhone:
				fallthrough
			case config.AuthenticationFlowIdentificationUsername:
				return authflow.NewNodeSimple(&NodeCreateIdentityLoginID{
					UserID:         i.UserID,
					Identification: identification,
				}), nil
			case config.AuthenticationFlowIdentificationOAuth:
				// FIXME(authflow): handle oauth
			case config.AuthenticationFlowIdentificationPasskey:
				// FIXME(authflow): handle passkey
			case config.AuthenticationFlowIdentificationSiwe:
				// FIXME(authflow): handle siwe
			}
		}
		return nil, authflow.ErrIncompatibleInput
	}

	_, identityCreated := authflow.FindMilestone[MilestoneDoCreateIdentity](flows.Nearest)
	_, standardAttributesPopulated := authflow.FindMilestone[MilestoneDoPopulateStandardAttributes](flows.Nearest)
	_, nestedStepHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case identityCreated && !standardAttributesPopulated && !nestedStepHandled:
		iden := i.identityInfo(flows.Nearest)
		return authflow.NewNodeSimple(&NodeDoPopulateStandardAttributes{
			Identity: iden,
		}), nil
	case identityCreated && standardAttributesPopulated && !nestedStepHandled:
		identification := i.identificationMethod(flows.Nearest)
		return authflow.NewSubFlow(&IntentSignupFlowSteps{
			SignupFlow:  i.SignupFlow,
			JSONPointer: i.jsonPointer(step, identification),
			UserID:      i.UserID,
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (*IntentSignupFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentSignupFlowStepIdentify) identificationMethod(w *authflow.Flow) config.AuthenticationFlowIdentification {
	m, ok := authflow.FindMilestone[MilestoneIdentificationMethod](w)
	if !ok {
		panic(fmt.Errorf("identification method not yet selected"))
	}

	im := m.MilestoneIdentificationMethod()

	return im
}

func (i *IntentSignupFlowStepIdentify) jsonPointer(step *config.AuthenticationFlowSignupFlowStep, im config.AuthenticationFlowIdentification) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}

func (*IntentSignupFlowStepIdentify) identityInfo(w *authflow.Flow) *identity.Info {
	m, ok := authflow.FindMilestone[MilestoneDoCreateIdentity](w)
	if !ok {
		panic(fmt.Errorf("MilestoneDoCreateIdentity cannot be found in IntentSignupFlowStepIdentify"))
	}
	info := m.MilestoneDoCreateIdentity()
	return info
}
