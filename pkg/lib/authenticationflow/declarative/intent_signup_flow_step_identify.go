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

func (*IntentSignupFlowStepIdentify) Kind() string {
	return "IntentSignupFlowStepIdentify"
}

func (i *IntentSignupFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	// Let the input to select which identification method to use.
	if len(flows.Nearest.Nodes) == 0 {
		var identifications []config.AuthenticationFlowIdentification
		for _, b := range step.OneOf {
			identifications = append(identifications, b.Identification)
		}

		return &InputSchemaStepIdentify{
			JSONPointer:     i.JSONPointer,
			Identifications: identifications,
			OAuthConfig:     deps.Config.Identity.OAuth,
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
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeIdentificationMethod inputTakeIdentificationMethod
		if authflow.AsInput(input, &inputTakeIdentificationMethod) {
			identification := inputTakeIdentificationMethod.GetIdentificationMethod()
			idx, err := i.checkIdentificationMethod(deps, step, identification)
			if err != nil {
				return nil, err
			}

			switch identification {
			case config.AuthenticationFlowIdentificationEmail:
				fallthrough
			case config.AuthenticationFlowIdentificationPhone:
				fallthrough
			case config.AuthenticationFlowIdentificationUsername:
				return authflow.NewNodeSimple(&NodeCreateIdentityLoginID{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Identification: identification,
				}), nil
			case config.AuthenticationFlowIdentificationOAuth:
				// FIXME(authflow): Support oauth in signup.
				return nil, authflow.ErrIncompatibleInput
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

func (*IntentSignupFlowStepIdentify) checkIdentificationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep, im config.AuthenticationFlowIdentification) (idx int, err error) {
	idx = -1

	for index, branch := range step.OneOf {
		branch := branch
		if im == branch.Identification {
			idx = index
		}
	}

	if idx >= 0 {
		return
	}

	err = authflow.ErrIncompatibleInput
	return
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
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
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
