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
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

func init() {
	authflow.RegisterIntent(&IntentPromoteFlowStepIdentify{})
}

// IntentPromoteFlowStepIdentify
//   IntentPromoteIdentityLoginID (MilestoneIdentificationMethod, MilestoneFlowCreateIdentity)
//     NodeDoCreateIdentity (MilestoneDoCreateIdentity)
//
//   IntentPromoteIdentityOAuth (MilestoneIdentificationMethod, MilestoneFlowCreateIdentity)
//     NodePromoteIdentityOAuth
//     NodeDoCreateIdentity (MilestoneDoCreateIdentity)

type IntentPromoteFlowStepIdentify struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	Options       []IdentificationOption `json:"options"`
}

var _ authflow.TargetStep = &IntentPromoteFlowStepIdentify{}

func (i *IntentPromoteFlowStepIdentify) GetName() string {
	return i.StepName
}

func (i *IntentPromoteFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentSignupFlowStepVerifyTarget = &IntentPromoteFlowStepIdentify{}

func (*IntentPromoteFlowStepIdentify) GetVerifiableClaims(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error) {
	m1, m1Flows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateIdentity](flows)
	if !ok {
		return nil, fmt.Errorf("MilestoneFlowCreateIdentity cannot be found in IntentPromoteFlowStepIdentify")
	}

	m2, _, ok := m1.MilestoneFlowCreateIdentity(m1Flows)
	if !ok {
		return nil, fmt.Errorf("MilestoneDoCreateIdentity cannot be found in IntentPromoteFlowStepIdentify")
	}

	info := m2.MilestoneDoCreateIdentity()

	return info.IdentityAwareStandardClaims(), nil
}

func (*IntentPromoteFlowStepIdentify) GetPurpose(_ context.Context, _ *authflow.Dependencies, _ authflow.Flows) otp.Purpose {
	return otp.PurposeVerification
}

func (*IntentPromoteFlowStepIdentify) GetMessageType(_ context.Context, _ *authflow.Dependencies, _ authflow.Flows) translation.MessageType {
	return translation.MessageTypeVerification
}

var _ IntentSignupFlowStepCreateAuthenticatorTarget = &IntentPromoteFlowStepIdentify{}

func (n *IntentPromoteFlowStepIdentify) GetOOBOTPClaims(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error) {
	return n.GetVerifiableClaims(ctx, deps, flows)
}

func (n *IntentPromoteFlowStepIdentify) IsSkipped() bool {
	return false
}

var _ authflow.Intent = &IntentPromoteFlowStepIdentify{}
var _ authflow.DataOutputer = &IntentPromoteFlowStepIdentify{}

func NewIntentPromoteFlowStepIdentify(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, i *IntentPromoteFlowStepIdentify, originNode authflow.NodeOrIntent) (*IntentPromoteFlowStepIdentify, error) {
	current, err := i.currentFlowObject(deps, flows, originNode)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	options := []IdentificationOption{}
	for _, b := range step.OneOf {
		switch b.Identification {
		case model.AuthenticationFlowIdentificationEmail:
			fallthrough
		case model.AuthenticationFlowIdentificationPhone:
			fallthrough
		case model.AuthenticationFlowIdentificationUsername:
			c := NewIdentificationOptionLoginID(flows, b.Identification, b.BotProtection, deps.Config.BotProtection)
			options = append(options, c)
		case model.AuthenticationFlowIdentificationOAuth:
			oauthOptions := NewIdentificationOptionsOAuth(
				flows,
				deps.Config.Identity.OAuth,
				deps.FeatureConfig.Identity.OAuth.Providers,
				b.BotProtection,
				deps.Config.BotProtection,
				deps.SSOOAuthDemoCredentials,
			)
			options = append(options, oauthOptions...)
		case model.AuthenticationFlowIdentificationPasskey:
			// Do not support create passkey in signup because
			// passkey is not considered as a persistent identifier.
			break
		}
	}

	i.Options = options
	return i, nil
}

func (*IntentPromoteFlowStepIdentify) Kind() string {
	return "IntentPromoteFlowStepIdentify"
}

func (i *IntentPromoteFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// Let the input to select which identification method to use.
	if len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, i)
		if err != nil {
			return nil, err
		}
		shouldBypassBotProtection := ShouldExistingResultBypassBotProtectionRequirement(ctx)
		return &InputSchemaStepIdentify{
			FlowRootObject:            flowRootObject,
			JSONPointer:               i.JSONPointer,
			Options:                   i.Options,
			ShouldBypassBotProtection: shouldBypassBotProtection,
			BotProtectionCfg:          deps.Config.BotProtection,
		}, nil
	}

	_, _, identityCreated := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateIdentity](flows)
	_, _, standardAttributesPopulated := authflow.FindMilestoneInCurrentFlow[MilestoneDoPopulateStandardAttributes](flows)
	_, _, nestedStepHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

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

func (i *IntentPromoteFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	current, err := i.currentFlowObject(deps, flows, i)
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

			syntheticInput := &InputStepIdentify{
				Identification: identification,
			}

			switch identification {
			case model.AuthenticationFlowIdentificationEmail:
				fallthrough
			case model.AuthenticationFlowIdentificationPhone:
				fallthrough
			case model.AuthenticationFlowIdentificationUsername:
				return authflow.NewSubFlow(&IntentPromoteIdentityLoginID{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Identification: identification,
					SyntheticInput: syntheticInput,
				}), nil
			case model.AuthenticationFlowIdentificationOAuth:
				return authflow.NewSubFlow(&IntentPromoteIdentityOAuth{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Identification: identification,
					SyntheticInput: syntheticInput,
				}), nil
			case model.AuthenticationFlowIdentificationPasskey:
				// Cannot create passkey in this step.
				return nil, authflow.ErrIncompatibleInput
			}
		}
	}

	_, _, identityCreated := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateIdentity](flows)
	_, _, standardAttributesPopulated := authflow.FindMilestoneInCurrentFlow[MilestoneDoPopulateStandardAttributes](flows)
	_, _, nestedStepHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case identityCreated && !standardAttributesPopulated && !nestedStepHandled:
		iden := i.identityInfo(flows)
		return authflow.NewNodeSimple(&NodeDoPopulateStandardAttributesInSignup{
			Identity: iden,
		}), nil
	case identityCreated && standardAttributesPopulated && !nestedStepHandled:
		identification := i.identificationMethod(flows)
		return authflow.NewSubFlow(&IntentPromoteFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.jsonPointer(step, identification),
			UserID:        i.UserID,
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentPromoteFlowStepIdentify) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NewIdentificationData(IdentificationData{
		Options: i.Options,
	}), nil
}

func (*IntentPromoteFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentPromoteFlowStepIdentify) checkIdentificationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep, im model.AuthenticationFlowIdentification) (idx int, err error) {
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

func (*IntentPromoteFlowStepIdentify) identificationMethod(flows authflow.Flows) model.AuthenticationFlowIdentification {
	m, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneIdentificationMethod](flows)
	if !ok {
		panic(fmt.Errorf("identification method not yet selected"))
	}

	im := m.MilestoneIdentificationMethod()

	return im
}

func (i *IntentPromoteFlowStepIdentify) jsonPointer(step *config.AuthenticationFlowSignupFlowStep, im model.AuthenticationFlowIdentification) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}

func (*IntentPromoteFlowStepIdentify) identityInfo(flows authflow.Flows) *identity.Info {
	m1, m1Flows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateIdentity](flows)
	if !ok {
		panic(fmt.Errorf("MilestoneFlowCreateIdentity cannot be found in IntentPromoteFlowStepIdentify"))
	}

	m2, _, ok := m1.MilestoneFlowCreateIdentity(m1Flows)
	if !ok {
		panic(fmt.Errorf("MilestoneDoCreateIdentity cannot be found in IntentPromoteFlowStepIdentify"))
	}

	info := m2.MilestoneDoCreateIdentity()
	return info
}

func (i *IntentPromoteFlowStepIdentify) currentFlowObject(deps *authflow.Dependencies, flows authflow.Flows, originNode authflow.NodeOrIntent) (config.AuthenticationFlowObject, error) {
	rootObject, err := findNearestFlowObjectInFlow(deps, flows, originNode)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(rootObject, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	return current, nil
}
